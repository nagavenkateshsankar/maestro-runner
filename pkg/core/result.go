package core

import (
	"time"
)

// StepResult captures the complete outcome of executing a single step
type StepResult struct {
	// Identity
	Index   int    `json:"index"`   // 0-based position in flow
	Command string `json:"command"` // Command type: tapOn, assertVisible, etc.

	// Status
	Status   StepStatus    `json:"status"`
	Category ErrorCategory `json:"errorCategory,omitempty"`

	// Timing
	StartTime time.Time     `json:"startTime"`
	Duration  time.Duration `json:"duration"`

	// Error Details
	Error    string `json:"error,omitempty"`    // Technical error message
	Message  string `json:"message,omitempty"`  // Human-readable explanation
	Expected string `json:"expected,omitempty"` // What was expected (for assertions)
	Actual   string `json:"actual,omitempty"`   // What was found

	// Retry Tracking
	Attempt     int      `json:"attempt"`               // Current attempt (1-based)
	MaxAttempts int      `json:"maxAttempts"`           // Configured max retries + 1
	RetryErrors []string `json:"retryErrors,omitempty"` // Errors from previous attempts

	// Debug Artifacts
	Attachments []Attachment `json:"attachments,omitempty"`
}

// FlowResult captures the complete outcome of executing a flow
type FlowResult struct {
	// Identity
	Name     string   `json:"name"`
	FilePath string   `json:"filePath"`
	Tags     []string `json:"tags,omitempty"`

	// Status (aggregated from steps)
	Status StepStatus `json:"status"`

	// Timing
	StartTime time.Time     `json:"startTime"`
	Duration  time.Duration `json:"duration"`

	// Results
	Steps          []StepResult `json:"steps"`
	OnFlowStart    []StepResult `json:"onFlowStart,omitempty"`
	OnFlowComplete []StepResult `json:"onFlowComplete,omitempty"`

	// Summary (computed)
	TotalSteps   int `json:"totalSteps"`
	PassedSteps  int `json:"passedSteps"`
	FailedSteps  int `json:"failedSteps"`
	SkippedSteps int `json:"skippedSteps"`
	WarnedSteps  int `json:"warnedSteps"`

	// Error info (if flow failed)
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// ComputeSummary calculates step counts from the Steps slice
func (f *FlowResult) ComputeSummary() {
	f.TotalSteps = len(f.Steps)
	f.PassedSteps = 0
	f.FailedSteps = 0
	f.SkippedSteps = 0
	f.WarnedSteps = 0

	for _, step := range f.Steps {
		switch step.Status {
		case StatusPassed:
			f.PassedSteps++
		case StatusFailed, StatusErrored:
			f.FailedSteps++
		case StatusSkipped:
			f.SkippedSteps++
		case StatusWarned:
			f.WarnedSteps++
		}
	}
}

// AggregateStatus determines the flow status from step results
// Rules:
// - Any failed/errored step (non-optional) → StatusFailed
// - All passed (with optional warned) → StatusPassed
// - onFlowStart failed → StatusFailed (steps skipped)
func (f *FlowResult) AggregateStatus() StepStatus {
	// Check onFlowStart first
	for _, step := range f.OnFlowStart {
		if step.Status == StatusFailed || step.Status == StatusErrored {
			return StatusFailed
		}
	}

	// Check main steps
	hasWarned := false
	for _, step := range f.Steps {
		if step.Status == StatusFailed || step.Status == StatusErrored {
			return StatusFailed
		}
		if step.Status == StatusWarned {
			hasWarned = true
		}
	}

	// Check onFlowComplete
	for _, step := range f.OnFlowComplete {
		if step.Status == StatusFailed || step.Status == StatusErrored {
			return StatusFailed
		}
	}

	// All steps executed without hard failure
	if hasWarned {
		return StatusWarned
	}
	return StatusPassed
}

// SuiteResult captures the complete outcome of executing multiple flows
type SuiteResult struct {
	// Identity
	Name  string `json:"name"`
	RunID string `json:"runId"` // Unique execution ID (timestamp or UUID)

	// Environment
	Platform string `json:"platform"` // ios, android
	Device   string `json:"device"`
	AppID    string `json:"appId"`

	// Timing
	StartTime time.Time     `json:"startTime"`
	Duration  time.Duration `json:"duration"`

	// Results
	Flows []FlowResult `json:"flows"`

	// Summary
	TotalFlows   int `json:"totalFlows"`
	PassedFlows  int `json:"passedFlows"`
	FailedFlows  int `json:"failedFlows"`
	SkippedFlows int `json:"skippedFlows"`
}

// ComputeSummary calculates flow counts from the Flows slice
func (s *SuiteResult) ComputeSummary() {
	s.TotalFlows = len(s.Flows)
	s.PassedFlows = 0
	s.FailedFlows = 0
	s.SkippedFlows = 0

	for _, flow := range s.Flows {
		switch flow.Status {
		case StatusPassed, StatusWarned:
			s.PassedFlows++
		case StatusFailed, StatusErrored:
			s.FailedFlows++
		case StatusSkipped:
			s.SkippedFlows++
		}
	}
}

// Success returns true if all flows passed (including warned)
func (s *SuiteResult) Success() bool {
	for _, flow := range s.Flows {
		if !flow.Status.IsSuccess() {
			return false
		}
	}
	return len(s.Flows) > 0
}
