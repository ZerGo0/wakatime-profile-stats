package main

import (
	"fmt"
	"log"
	"os"

	"github.com/user/wakatime-profile-stats/internal/errors"
	"github.com/user/wakatime-profile-stats/pkg/git"
	"github.com/user/wakatime-profile-stats/pkg/github"
	"github.com/user/wakatime-profile-stats/pkg/stats"
	"github.com/user/wakatime-profile-stats/pkg/wakatime"

	go_github "github.com/google/go-github/v65/github"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

func main() {
	env := os.Getenv("ENV")
	var logger *zap.Logger
	var err error
	if env == "prod" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		log.Fatalf("failed to setup logger: %v", err)
	}
	zap.ReplaceGlobals(logger)

	if err := run(); err != nil {
		zap.L().Error("an error occurred", zap.Error(err))
		os.Exit(1)
	}
}

func run() error {
	_, err := maxprocs.Set(maxprocs.Logger(func(s string, i ...interface{}) {
		zap.L().Debug(fmt.Sprintf(s, i...))
	}))
	if err != nil {
		return err
	}

	zap.L().Info("Starting Wakatime Profile Stats")

	wakaAPIKey, githubToken, err := retrieveEnvVars()
	if err != nil {
		return err
	}

	currentUser, repos, repoPath, err := getGithubRepos(githubToken)
	if err != nil {
		return err
	}

	sevenDaysStats, monthlyStats, yearlyStats, allTimeStats, err := getWakaStats(wakaAPIKey)
	if err != nil {
		return err
	}

	err = updateReadmeStats(repoPath, currentUser, sevenDaysStats, monthlyStats, yearlyStats, allTimeStats, repos)
	if err != nil {
		return err
	}

	zap.L().Info("Wakatime Profile Stats completed successfully")

	return nil
}

func retrieveEnvVars() (string, string, error) {
	wakaAPIKey := os.Getenv("INPUT_WAKATIME_API_KEY")
	if wakaAPIKey == "" {
		return "", "", errors.ErrWakatimeAPIKeyRequired
	}

	githubToken := os.Getenv("INPUT_GH_TOKEN")
	if githubToken == "" {
		return "", "", errors.ErrGithubTokenRequired
	}

	zap.L().Info("Environment variables are present and valid")
	return wakaAPIKey, githubToken, nil
}

func getGithubRepos(githubToken string) (string, []*go_github.Repository, string, error) {
	gClient, err := github.NewGithubClient(githubToken)
	if err != nil {
		return "", nil, "", err
	}

	user, err := gClient.GetUser()
	if err != nil {
		return "", nil, "", err
	}

	repos, err := gClient.GetRepos()
	if err != nil {
		return "", nil, "", err
	}

	remoteName := (*user.Login) + "/" + (*user.Login)
	repoPath := "https://" + githubToken + "@github.com/" + remoteName + ".git"

	zap.L().Info("Github login was successful", zap.Int("repos", len(repos)))

	return *user.Login, repos, repoPath, nil
}

func getWakaStats(wakaAPIKey string) (*wakatime.WakaStats, *wakatime.WakaStats, *wakatime.WakaStats, *wakatime.WakaStats, error) {
	wClient := wakatime.NewClient("https://wakatime.com/api/v1", wakaAPIKey)
	sevenDaysStats, err := wClient.GetStats("last_7_days")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	monthlyStats, err := wClient.GetStats("last_30_days")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	yearlyStats, err := wClient.GetStats("last_year")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	allTimeStats, err := wClient.GetStats("all_time")
	if err != nil {
		return nil, nil, nil, nil, err
	}

	zap.L().Info("Got Wakatime stats")
	return sevenDaysStats, monthlyStats, yearlyStats, allTimeStats, nil
}

func updateReadmeStats(repoPath, currentUser string, sevenDaysStats, monthlyStats, yearlyStats, allTimeStats *wakatime.WakaStats, githubRepos []*go_github.Repository) error {
	g, err := git.SetupRepo(repoPath)
	if err != nil {
		return err
	}

	textStats, err := stats.ProcessStats(currentUser, sevenDaysStats, monthlyStats, yearlyStats, allTimeStats, githubRepos)
	if err != nil {
		return err
	}

	err = g.UpdateStats(*textStats)
	if err != nil {
		return err
	}

	err = g.CommitAndPush()
	if err != nil {
		return err
	}

	return nil
}
