package stats

import (
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v65/github"
	"github.com/user/wakatime-profile-stats/pkg/wakatime"
	"go.uber.org/zap"
)

const (
	codeBlockStart = "```text"
	codeBlockEnd   = "```"
)

func ProcessStats(currentUser string, sevenDaysStats, monthlyStats, yearlyStats, allTimeStats *wakatime.WakaStats, githubRepos []*github.Repository) (*string, error) {
	stats := ""

	if sevenDaysStats == nil && yearlyStats == nil && allTimeStats == nil {
		return nil, nil
	}

	stats += processCodingStats(sevenDaysStats, monthlyStats, yearlyStats, allTimeStats)
	sevenDaysStr := processStats("Last 7 days", currentUser, sevenDaysStats, githubRepos)
	monthlyStr := processStats("Last 30 days", currentUser, monthlyStats, githubRepos)
	yearlyStr := processStats("Last 365 days", currentUser, yearlyStats, githubRepos)
	alltimeStr := processStats("All time", currentUser, allTimeStats, githubRepos)
	stats += getCombinedStats(sevenDaysStr, monthlyStr, yearlyStr, alltimeStr)
	stats += "Updated at " + time.Now().Format("2006-01-02 15:04:05") + " (UTC) using [ZerGo0/wakatime-profile-stats](https://github.com/ZerGo0/wakatime-profile-stats)"

	zap.L().Info("Processed stats")

	return &stats, nil
}

func processCodingStats(sevenDaysStats, monthlyStats, yearlyStats, allTimeStats *wakatime.WakaStats) string {
	if sevenDaysStats == nil || monthlyStats == nil || yearlyStats == nil || allTimeStats == nil {
		zap.L().Warn("Some stats are nil, skipping")
		os.Exit(0)
	}

	codingTimeSevenDays := sevenDaysStats.Data.HumanReadableTotalIncludingOtherLanguage
	codingTimeMonthly := calculateMaxCodingTime(monthlyStats, sevenDaysStats)
	codingTimeYearly := calculateMaxCodingTime(yearlyStats, monthlyStats)
	codingTimeAllTime := calculateMaxCodingTime(allTimeStats, yearlyStats)

	return `Code Time:

` + codeBlockStart + `
Last 7 days:             ` + codingTimeSevenDays + `
Last 30 days:            ` + codingTimeMonthly + `
Last 365 days:           ` + codingTimeYearly + `
All time:                ` + codingTimeAllTime + `
` + codeBlockEnd + `
`
}

func processStats(title, currentUser string, stats *wakatime.WakaStats, githubRepos []*github.Repository) string {
	if stats == nil {
		return ""
	}

	calculateWorkTime(currentUser, stats, githubRepos)
	topProjects := calculateTopProjects(stats)
	topLanguages := calculateTopLanguages(stats)

	return title + `

Projects:
` + formatObjects(topProjects) + `
Languages:
` + formatObjects(topLanguages)
}

func getCombinedStats(sevenDaysStr, monthlyStr, yearlyStr, alltimeStr string) string {
	sevenLongestLine := getLongestLine(sevenDaysStr)
	monthlyLongestLine := getLongestLine(monthlyStr)
	yearlyLongestLine := getLongestLine(yearlyStr)
	alltimeLongestLine := getLongestLine(alltimeStr)

	combinedStats := ""
	sevenDaysLines := strings.Split(sevenDaysStr, "\n")
	monthlyLines := strings.Split(monthlyStr, "\n")
	yearlyLines := strings.Split(yearlyStr, "\n")
	alltimeLines := strings.Split(alltimeStr, "\n")

	for i := 0; i < len(sevenDaysLines)-1 || i < len(monthlyLines)-1 || i < len(yearlyLines)-1 || i < len(alltimeLines)-1; i++ {
		sevenDaysLine := ""
		if i < len(sevenDaysLines) {
			sevenDaysLine = sevenDaysLines[i]
		}

		monthlyLine := ""
		if i < len(monthlyLines) {
			monthlyLine = monthlyLines[i]
		}

		yearlyLine := ""
		if i < len(yearlyLines) {
			yearlyLine = yearlyLines[i]
		}

		alltimeLine := ""
		if i < len(alltimeLines) {
			alltimeLine = alltimeLines[i]
		}

		combinedStats += sevenDaysLine + strings.Repeat(" ", sevenLongestLine-len(sevenDaysLine)) + " | " +
			monthlyLine + strings.Repeat(" ", monthlyLongestLine-len(monthlyLine)) + " | " +
			yearlyLine + strings.Repeat(" ", yearlyLongestLine-len(yearlyLine)) + " | " +
			alltimeLine + strings.Repeat(" ", alltimeLongestLine-len(alltimeLine)) + "\n"
	}

	return "Projects and Languages:\n" + codeBlockStart + "\n" + combinedStats + codeBlockEnd + "\n"
}

func getLongestLine(str string) int {
	longest := 0
	for _, line := range strings.Split(str, "\n") {
		if len(line) > longest {
			longest = len(line)
		}
	}

	return longest
}
