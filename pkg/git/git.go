package git

import (
	"os"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/user/wakatime-profile-stats/internal/errors"
	"go.uber.org/zap"
)

const (
	startMarker = "<!--START_SECTION:waka-->"
	endMarker   = "<!--END_SECTION:waka-->"
	filePerm    = 0600
)

const (
	localRepoPath = "/tmp/repo"
)

type Git struct {
	gitRepo *git.Repository
	setup   bool
}

func SetupRepo(repoPath string) (*Git, error) {
	if _, err := os.Stat(localRepoPath); err == nil {
		if err := os.RemoveAll(localRepoPath); err != nil {
			return nil, err
		}
	}

	repo, err := git.PlainClone(localRepoPath, false, &git.CloneOptions{
		URL: repoPath,
	})
	if err != nil {
		return nil, err
	}

	zap.L().Info("Cloned repo")

	return &Git{
		gitRepo: repo,
		setup:   true,
	}, nil
}

func (g *Git) UpdateStats(stats string) error {
	if !g.setup {
		return errors.ErrRepoNotSetup
	}

	readmeBytes, err := os.ReadFile(localRepoPath + "/README.md")
	if err != nil {
		return err
	}

	startIndex := -1
	endIndex := -1
	readme := string(readmeBytes)
	if start := findMarker(readme, startMarker); start != -1 {
		startIndex = start
	}
	if end := findMarker(readme, endMarker); end != -1 {
		endIndex = end
	}

	if startIndex == -1 || endIndex == -1 {
		return errors.ErrTagNotFound
	}

	newReadme := readme[:startIndex+len(startMarker)] + "\n" + stats + "\n" + readme[endIndex:]

	if err := os.WriteFile(localRepoPath+"/README.md", []byte(newReadme), filePerm); err != nil {
		return err
	}

	zap.L().Info("Updated README.md")

	return nil
}

func (g *Git) CommitAndPush() error {
	if !g.setup {
		return errors.ErrRepoNotSetup
	}

	wt, err := g.gitRepo.Worktree()
	if err != nil {
		return err
	}

	if _, err := wt.Add("README.md"); err != nil {
		return err
	}

	botSignature := &object.Signature{
		Name:  "github-actions[bot]",
		Email: "github-actions[bot]@users.noreply.github.com",
		When:  time.Now(),
	}

	// commit the changes as bot
	_, err = wt.Commit("Add debug string to README", &git.CommitOptions{
		Author:    botSignature,
		Committer: botSignature,
	})
	if err != nil {
		return err
	}

	// push the changes
	if err := g.gitRepo.Push(&git.PushOptions{}); err != nil {
		return err
	}

	zap.L().Info("Committed and pushed changes")

	return nil
}

func findMarker(readme, marker string) int {
	for i := 0; i < len(readme); i++ {
		if len(readme)-i < len(marker) {
			break
		}

		if readme[i:i+len(marker)] == marker {
			return i
		}
	}

	return -1
}
