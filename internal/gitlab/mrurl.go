package gitlab

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// MRURL is the parsed shape of a GitLab merge-request URL.
type MRURL struct {
	Host        string // "https://gitlab.example.com[:port]"
	ProjectPath string // "group/sub/proj"
	IID         int
}

var mrPathRE = regexp.MustCompile(`^/(.+?)/-/merge_requests/(\d+)/?$`)

// ParseMRURL extracts host, project path, and IID from a GitLab MR URL.
func ParseMRURL(s string) (MRURL, error) {
	u, err := url.Parse(strings.TrimSpace(s))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return MRURL{}, fmt.Errorf("not a URL: %q", s)
	}
	m := mrPathRE.FindStringSubmatch(u.Path)
	if m == nil {
		return MRURL{}, fmt.Errorf("not a GitLab MR URL: %q", s)
	}
	iid, err := strconv.Atoi(m[2])
	if err != nil {
		return MRURL{}, fmt.Errorf("invalid IID: %w", err)
	}
	return MRURL{Host: u.Scheme + "://" + u.Host, ProjectPath: m[1], IID: iid}, nil
}

// EncodedProjectPath returns the URL-encoded project path for /api/v4/projects/:id/... usage.
func (m MRURL) EncodedProjectPath() string {
	return url.PathEscape(m.ProjectPath)
}
