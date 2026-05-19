package model

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// JSONLWriter writes type-discriminated JSON records, one per line.
type JSONLWriter struct {
	w *bufio.Writer
}

// NewJSONLWriter wraps w with line buffering.
func NewJSONLWriter(w io.Writer) *JSONLWriter {
	return &JSONLWriter{w: bufio.NewWriter(w)}
}

// Flush flushes buffered bytes to the underlying writer.
func (j *JSONLWriter) Flush() error { return j.w.Flush() }

// writeTyped merges {"type": t} with payload's JSON object and emits one line.
func (j *JSONLWriter) writeTyped(t string, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if len(raw) == 0 || raw[0] != '{' {
		return fmt.Errorf("payload must marshal to JSON object, got %s", string(raw))
	}
	tb, err := json.Marshal(t)
	if err != nil {
		return err
	}
	var out []byte
	if string(raw) == "{}" {
		out = append([]byte(`{"type":`), tb...)
		out = append(out, '}')
	} else {
		out = append([]byte(`{"type":`), tb...)
		out = append(out, ',')
		out = append(out, raw[1:]...) // payload without leading '{'
	}
	out = append(out, '\n')
	_, err = j.w.Write(out)
	return err
}

// WriteMeta emits the run-meta line.
func (j *JSONLWriter) WriteMeta(m Meta) error { return j.writeTyped("meta", m) }

// WriteIssue emits one sonar.issue line.
func (j *JSONLWriter) WriteIssue(i Issue) error { return j.writeTyped("sonar.issue", i) }

// WriteHotspot emits one sonar.hotspot line.
func (j *JSONLWriter) WriteHotspot(h Hotspot) error { return j.writeTyped("sonar.hotspot", h) }

// WriteMeasure emits one sonar.measure line.
func (j *JSONLWriter) WriteMeasure(m Measure) error { return j.writeTyped("sonar.measure", m) }

// WriteQualityGate emits the sonar.quality_gate line.
func (j *JSONLWriter) WriteQualityGate(qg QualityGate) error {
	return j.writeTyped("sonar.quality_gate", qg)
}

// WriteNexusViolation emits one nexus.violation line.
func (j *JSONLWriter) WriteNexusViolation(v Violation) error {
	return j.writeTyped("nexus.violation", v)
}

// WriteMR emits the gitlab.mr line.
func (j *JSONLWriter) WriteMR(m MR) error { return j.writeTyped("gitlab.mr", m) }

// WriteDiffSummary emits the gitlab.mr.diff_summary line.
func (j *JSONLWriter) WriteDiffSummary(d DiffSummary) error {
	return j.writeTyped("gitlab.mr.diff_summary", d)
}

// WriteDiscussion emits one gitlab.mr.discussion line.
func (j *JSONLWriter) WriteDiscussion(d Discussion) error {
	return j.writeTyped("gitlab.mr.discussion", d)
}

// WritePipeline emits the gitlab.pipeline line.
func (j *JSONLWriter) WritePipeline(p Pipeline) error {
	return j.writeTyped("gitlab.pipeline", p)
}

// WriteJob emits one gitlab.job line.
func (j *JSONLWriter) WriteJob(jb Job) error { return j.writeTyped("gitlab.job", jb) }

// WriteError emits one error line.
func (j *JSONLWriter) WriteError(e SourceError) error { return j.writeTyped("error", e) }
