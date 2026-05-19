package gitlab

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/RandomCodeSpace/qctx/internal/model"
)

var (
	keyREDFlag = regexp.MustCompile(`-Dsonar\.projectKey=([A-Za-z0-9._:\-]+)`)
	keyREProps = regexp.MustCompile(`(?m)^[ \t]*sonar\.projectKey[ \t]*[=:][ \t]*([A-Za-z0-9._:\-]+)`)
)

// DiscoverSonarProjectKey scans pipeline job traces for sonar.projectKey.
// Jobs are scanned in parallel (last 64 KiB of trace per job). First match wins.
func (c *Client) DiscoverSonarProjectKey(projectPath string, pipelineID int) (string, error) {
	jobs, err := c.ListPipelineJobs(projectPath, pipelineID)
	if err != nil {
		return "", fmt.Errorf("discover: list jobs: %w", err)
	}
	// Prioritize jobs with "sonar" in their name.
	prio := make([]model.Job, 0, len(jobs))
	rest := make([]model.Job, 0, len(jobs))
	for _, j := range jobs {
		if strings.Contains(strings.ToLower(j.Name), "sonar") {
			prio = append(prio, j)
		} else {
			rest = append(rest, j)
		}
	}
	jobs = append(prio, rest...)

	type res struct {
		key string
		err error
	}
	resCh := make(chan res, len(jobs))
	var wg sync.WaitGroup
	for _, j := range jobs {
		wg.Add(1)
		go func(jobID int) {
			defer wg.Done()
			trace, err := c.GetJobTraceTail(projectPath, jobID, 64<<10)
			if err != nil {
				resCh <- res{err: err}
				return
			}
			if k := findProjectKey(trace); k != "" {
				resCh <- res{key: k}
				return
			}
			resCh <- res{}
		}(j.ID)
	}
	go func() { wg.Wait(); close(resCh) }()

	for r := range resCh {
		if r.key != "" {
			return r.key, nil
		}
	}
	return "", fmt.Errorf("could not discover sonar.projectKey from pipeline %d; pass --project KEY explicitly", pipelineID)
}

func findProjectKey(s string) string {
	if m := keyREDFlag.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	if m := keyREProps.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return ""
}
