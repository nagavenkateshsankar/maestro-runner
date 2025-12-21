package core

import (
	"testing"
	"time"
)

func TestFlowResult_ComputeSummary(t *testing.T) {
	flow := &FlowResult{
		Name: "test-flow",
		Steps: []StepResult{
			{Index: 0, Status: StatusPassed},
			{Index: 1, Status: StatusPassed},
			{Index: 2, Status: StatusFailed},
			{Index: 3, Status: StatusSkipped},
			{Index: 4, Status: StatusWarned},
			{Index: 5, Status: StatusErrored},
		},
	}

	flow.ComputeSummary()

	if flow.TotalSteps != 6 {
		t.Errorf("TotalSteps = %d, want 6", flow.TotalSteps)
	}
	if flow.PassedSteps != 2 {
		t.Errorf("PassedSteps = %d, want 2", flow.PassedSteps)
	}
	if flow.FailedSteps != 2 { // Failed + Errored
		t.Errorf("FailedSteps = %d, want 2", flow.FailedSteps)
	}
	if flow.SkippedSteps != 1 {
		t.Errorf("SkippedSteps = %d, want 1", flow.SkippedSteps)
	}
	if flow.WarnedSteps != 1 {
		t.Errorf("WarnedSteps = %d, want 1", flow.WarnedSteps)
	}
}

func TestFlowResult_ComputeSummary_Empty(t *testing.T) {
	flow := &FlowResult{Name: "empty-flow"}
	flow.ComputeSummary()

	if flow.TotalSteps != 0 {
		t.Errorf("TotalSteps = %d, want 0", flow.TotalSteps)
	}
}

func TestFlowResult_AggregateStatus_AllPassed(t *testing.T) {
	flow := &FlowResult{
		Steps: []StepResult{
			{Status: StatusPassed},
			{Status: StatusPassed},
		},
	}

	if got := flow.AggregateStatus(); got != StatusPassed {
		t.Errorf("AggregateStatus() = %s, want %s", got, StatusPassed)
	}
}

func TestFlowResult_AggregateStatus_WithWarned(t *testing.T) {
	flow := &FlowResult{
		Steps: []StepResult{
			{Status: StatusPassed},
			{Status: StatusWarned},
			{Status: StatusPassed},
		},
	}

	if got := flow.AggregateStatus(); got != StatusWarned {
		t.Errorf("AggregateStatus() = %s, want %s", got, StatusWarned)
	}
}

func TestFlowResult_AggregateStatus_WithFailed(t *testing.T) {
	flow := &FlowResult{
		Steps: []StepResult{
			{Status: StatusPassed},
			{Status: StatusFailed},
			{Status: StatusSkipped},
		},
	}

	if got := flow.AggregateStatus(); got != StatusFailed {
		t.Errorf("AggregateStatus() = %s, want %s", got, StatusFailed)
	}
}

func TestFlowResult_AggregateStatus_WithErrored(t *testing.T) {
	flow := &FlowResult{
		Steps: []StepResult{
			{Status: StatusPassed},
			{Status: StatusErrored},
		},
	}

	if got := flow.AggregateStatus(); got != StatusFailed {
		t.Errorf("AggregateStatus() = %s, want %s", got, StatusFailed)
	}
}

func TestFlowResult_AggregateStatus_OnFlowStartFailed(t *testing.T) {
	flow := &FlowResult{
		OnFlowStart: []StepResult{
			{Status: StatusFailed},
		},
		Steps: []StepResult{
			{Status: StatusPassed},
		},
	}

	if got := flow.AggregateStatus(); got != StatusFailed {
		t.Errorf("AggregateStatus() = %s, want %s", got, StatusFailed)
	}
}

func TestFlowResult_AggregateStatus_OnFlowCompleteFailed(t *testing.T) {
	flow := &FlowResult{
		Steps: []StepResult{
			{Status: StatusPassed},
		},
		OnFlowComplete: []StepResult{
			{Status: StatusErrored},
		},
	}

	if got := flow.AggregateStatus(); got != StatusFailed {
		t.Errorf("AggregateStatus() = %s, want %s", got, StatusFailed)
	}
}

func TestSuiteResult_ComputeSummary(t *testing.T) {
	suite := &SuiteResult{
		Flows: []FlowResult{
			{Status: StatusPassed},
			{Status: StatusPassed},
			{Status: StatusFailed},
			{Status: StatusWarned},
			{Status: StatusSkipped},
		},
	}

	suite.ComputeSummary()

	if suite.TotalFlows != 5 {
		t.Errorf("TotalFlows = %d, want 5", suite.TotalFlows)
	}
	if suite.PassedFlows != 3 { // Passed + Warned
		t.Errorf("PassedFlows = %d, want 3", suite.PassedFlows)
	}
	if suite.FailedFlows != 1 {
		t.Errorf("FailedFlows = %d, want 1", suite.FailedFlows)
	}
	if suite.SkippedFlows != 1 {
		t.Errorf("SkippedFlows = %d, want 1", suite.SkippedFlows)
	}
}

func TestSuiteResult_Success(t *testing.T) {
	tests := []struct {
		name     string
		flows    []FlowResult
		expected bool
	}{
		{
			name:     "all passed",
			flows:    []FlowResult{{Status: StatusPassed}, {Status: StatusPassed}},
			expected: true,
		},
		{
			name:     "passed and warned",
			flows:    []FlowResult{{Status: StatusPassed}, {Status: StatusWarned}},
			expected: true,
		},
		{
			name:     "one failed",
			flows:    []FlowResult{{Status: StatusPassed}, {Status: StatusFailed}},
			expected: false,
		},
		{
			name:     "empty suite",
			flows:    []FlowResult{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suite := &SuiteResult{Flows: tt.flows}
			if got := suite.Success(); got != tt.expected {
				t.Errorf("Success() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStepResult_Fields(t *testing.T) {
	now := time.Now()
	step := StepResult{
		Index:       0,
		Command:     "tapOn",
		Status:      StatusPassed,
		Category:    ErrCategoryNone,
		StartTime:   now,
		Duration:    100 * time.Millisecond,
		Attempt:     1,
		MaxAttempts: 3,
	}

	if step.Command != "tapOn" {
		t.Errorf("Command = %s, want tapOn", step.Command)
	}
	if step.Attempt != 1 {
		t.Errorf("Attempt = %d, want 1", step.Attempt)
	}
}
