// Package model holds normalized output types shared across source clients.
package model

import "time"

type Meta struct {
	Tool            string            `json:"tool"`
	Version         string            `json:"version"`
	ScannedAt       time.Time         `json:"scanned_at"`
	SonarProjectKey string            `json:"sonar_project_key,omitempty"`
	GitLabProject   string            `json:"gitlab_project,omitempty"`
	Branch          string            `json:"branch,omitempty"`
	CommitSHA       string            `json:"commit_sha,omitempty"`
	MRIID           int               `json:"mr_iid,omitempty"`
	SourceStatus    map[string]string `json:"source_status"`
}

type Issue struct {
	Key          string   `json:"key"`
	Rule         string   `json:"rule"`
	Severity     string   `json:"severity"`
	Type         string   `json:"issue_type"`
	File         string   `json:"file"`
	Line         int      `json:"line,omitempty"`
	EndLine      int      `json:"end_line,omitempty"`
	Message      string   `json:"message"`
	Author       string   `json:"author,omitempty"`
	Effort       string   `json:"effort,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Status       string   `json:"status"`
	RuleDescHTML string   `json:"rule_desc_html,omitempty"`
}

type Hotspot struct {
	Key                      string `json:"key"`
	Rule                     string `json:"rule"`
	VulnerabilityProbability string `json:"vulnerability_probability"`
	Status                   string `json:"status"`
	File                     string `json:"file"`
	Line                     int    `json:"line,omitempty"`
	Message                  string `json:"message"`
	RuleDescHTML             string `json:"rule_desc_html,omitempty"`
}

type Measure struct {
	Metric string  `json:"metric"`
	Value  float64 `json:"value"`
}

type GateCondition struct {
	Metric    string `json:"metric"`
	Op        string `json:"op"`
	Threshold string `json:"threshold"`
	Actual    string `json:"actual"`
	Status    string `json:"status"`
}

type QualityGate struct {
	Status     string          `json:"status"`
	Conditions []GateCondition `json:"conditions,omitempty"`
}

type Violation struct {
	Component   string   `json:"component"`
	Manifest    string   `json:"manifest,omitempty"`
	Line        int      `json:"line,omitempty"`
	Policy      string   `json:"policy"`
	ThreatLevel int      `json:"threat_level"`
	CVEs        []string `json:"cves,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	FixVersion  string   `json:"fix_version,omitempty"`
	Status      string   `json:"status"`
}

type MR struct {
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	Author       string `json:"author,omitempty"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
	WebURL       string `json:"web_url"`
	Draft        bool   `json:"draft"`
	ChangesCount string `json:"changes_count,omitempty"`
}

type DiffSummary struct {
	FilesChanged []string `json:"files_changed"`
	Additions    int      `json:"additions"`
	Deletions    int      `json:"deletions"`
}

type Discussion struct {
	ID       string `json:"id"`
	Author   string `json:"author"`
	Body     string `json:"body"`
	Resolved bool   `json:"resolved"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
}

type Pipeline struct {
	ID        int    `json:"id"`
	Status    string `json:"status"`
	Ref       string `json:"ref"`
	SHA       string `json:"sha"`
	WebURL    string `json:"web_url"`
	CreatedAt string `json:"created_at"`
	Duration  int    `json:"duration"`
}

type Job struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	Stage          string `json:"stage"`
	Duration       int    `json:"duration"`
	WebURL         string `json:"web_url"`
	FailureExcerpt string `json:"failure_excerpt,omitempty"`
}

type SourceError struct {
	Source  string `json:"source"`
	Message string `json:"message"`
}
