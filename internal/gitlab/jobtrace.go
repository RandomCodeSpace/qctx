package gitlab

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// GetJobTraceTail fetches the last `bytes` of a job trace using Range: bytes=-N.
// Falls back to full body on 416 Range Not Satisfiable.
func (c *Client) GetJobTraceTail(projectPath string, jobID, bytesTail int) (string, error) {
	enc := encode(projectPath)
	path := fmt.Sprintf("/api/v4/projects/%s/jobs/%d/trace", enc, jobID)
	headers := http.Header{}
	if bytesTail > 0 {
		headers.Set("Range", fmt.Sprintf("bytes=-%d", bytesTail))
	}
	resp, err := c.GetRaw(path, nil, headers)
	if err != nil {
		if strings.Contains(err.Error(), "416") {
			return c.getJobTraceFull(projectPath, jobID)
		}
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *Client) getJobTraceFull(projectPath string, jobID int) (string, error) {
	enc := encode(projectPath)
	path := fmt.Sprintf("/api/v4/projects/%s/jobs/%d/trace", enc, jobID)
	resp, err := c.GetRaw(path, nil, nil)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FailureExcerpt returns the last `nLines` of trace.
func FailureExcerpt(trace string, nLines int) string {
	if trace == "" || nLines <= 0 {
		return ""
	}
	lines := strings.Split(strings.TrimRight(trace, "\n"), "\n")
	if len(lines) <= nLines {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[len(lines)-nLines:], "\n")
}
