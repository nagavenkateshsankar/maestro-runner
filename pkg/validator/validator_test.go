package validator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate_SingleFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.yaml")

	content := `
appId: com.example.app
---
- tapOn: "Login"
- inputText: "username"
`
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	v := New(nil, nil)
	result := v.Validate(file)

	if !result.IsValid() {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
	if len(result.Files) != 1 {
		t.Errorf("expected 1 file, got %d", len(result.Files))
	}
}

func TestValidate_Directory(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"flow1.yaml": `- tapOn: "Button1"`,
		"flow2.yaml": `- tapOn: "Button2"`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	v := New(nil, nil)
	result := v.Validate(dir)

	if !result.IsValid() {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
	if len(result.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(result.Files))
	}
}

func TestValidate_RunFlowResolution(t *testing.T) {
	dir := t.TempDir()

	// Main flow references sub flow
	mainFlow := `
- tapOn: "Start"
- runFlow: subflow.yaml
- tapOn: "End"
`
	subFlow := `
- tapOn: "SubStep1"
- tapOn: "SubStep2"
`

	if err := os.WriteFile(filepath.Join(dir, "main.yaml"), []byte(mainFlow), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "subflow.yaml"), []byte(subFlow), 0644); err != nil {
		t.Fatal(err)
	}

	v := New(nil, nil)
	result := v.Validate(filepath.Join(dir, "main.yaml"))

	if !result.IsValid() {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
	// Should include both main and subflow
	if len(result.Files) != 2 {
		t.Errorf("expected 2 files (main + subflow), got %d: %v", len(result.Files), result.Files)
	}
}

func TestValidate_NestedRunFlow(t *testing.T) {
	dir := t.TempDir()

	// main -> sub1 -> sub2
	mainFlow := `- runFlow: sub1.yaml`
	sub1Flow := `- runFlow: sub2.yaml`
	sub2Flow := `- tapOn: "Deep"`

	if err := os.WriteFile(filepath.Join(dir, "main.yaml"), []byte(mainFlow), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub1.yaml"), []byte(sub1Flow), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub2.yaml"), []byte(sub2Flow), 0644); err != nil {
		t.Fatal(err)
	}

	v := New(nil, nil)
	result := v.Validate(filepath.Join(dir, "main.yaml"))

	if !result.IsValid() {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
	if len(result.Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(result.Files))
	}
}

func TestValidate_CircularDependency(t *testing.T) {
	dir := t.TempDir()

	// a -> b -> a (circular)
	flowA := `- runFlow: b.yaml`
	flowB := `- runFlow: a.yaml`

	if err := os.WriteFile(filepath.Join(dir, "a.yaml"), []byte(flowA), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.yaml"), []byte(flowB), 0644); err != nil {
		t.Fatal(err)
	}

	v := New(nil, nil)
	result := v.Validate(filepath.Join(dir, "a.yaml"))

	if result.IsValid() {
		t.Error("expected circular dependency error")
	}

	// Check for circular dependency error message
	found := false
	for _, err := range result.Errors {
		if strings.Contains(err.Error(), "circular dependency") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected circular dependency error, got: %v", result.Errors)
	}
}

func TestValidate_SelfReference(t *testing.T) {
	dir := t.TempDir()

	// Flow references itself
	flow := `- runFlow: self.yaml`

	if err := os.WriteFile(filepath.Join(dir, "self.yaml"), []byte(flow), 0644); err != nil {
		t.Fatal(err)
	}

	v := New(nil, nil)
	result := v.Validate(filepath.Join(dir, "self.yaml"))

	if result.IsValid() {
		t.Error("expected circular dependency error for self-reference")
	}
}

func TestValidate_MissingRunFlowFile(t *testing.T) {
	dir := t.TempDir()

	flow := `- runFlow: nonexistent.yaml`

	if err := os.WriteFile(filepath.Join(dir, "main.yaml"), []byte(flow), 0644); err != nil {
		t.Fatal(err)
	}

	v := New(nil, nil)
	result := v.Validate(filepath.Join(dir, "main.yaml"))

	if result.IsValid() {
		t.Error("expected error for missing runFlow file")
	}
}

func TestValidate_InvalidYAML(t *testing.T) {
	dir := t.TempDir()

	flow := `- tapOn: [invalid yaml`

	if err := os.WriteFile(filepath.Join(dir, "invalid.yaml"), []byte(flow), 0644); err != nil {
		t.Fatal(err)
	}

	v := New(nil, nil)
	result := v.Validate(filepath.Join(dir, "invalid.yaml"))

	if result.IsValid() {
		t.Error("expected parse error for invalid YAML")
	}
}

func TestValidate_TagFiltering(t *testing.T) {
	dir := t.TempDir()

	smokeFlow := `
appId: com.example
tags:
  - smoke
---
- tapOn: "Smoke"
`
	regressionFlow := `
appId: com.example
tags:
  - regression
---
- tapOn: "Regression"
`

	if err := os.WriteFile(filepath.Join(dir, "smoke.yaml"), []byte(smokeFlow), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "regression.yaml"), []byte(regressionFlow), 0644); err != nil {
		t.Fatal(err)
	}

	// Test include tags
	v := New([]string{"smoke"}, nil)
	result := v.Validate(dir)

	if !result.IsValid() {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
	if len(result.Files) != 1 {
		t.Errorf("expected 1 file with smoke tag, got %d", len(result.Files))
	}

	// Test exclude tags
	v = New(nil, []string{"regression"})
	result = v.Validate(dir)

	if !result.IsValid() {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
	if len(result.Files) != 1 {
		t.Errorf("expected 1 file without regression tag, got %d", len(result.Files))
	}
}

func TestValidate_NonExistentPath(t *testing.T) {
	v := New(nil, nil)
	result := v.Validate("/nonexistent/path")

	if result.IsValid() {
		t.Error("expected error for nonexistent path")
	}
}

func TestValidate_RunFlowInSubdir(t *testing.T) {
	dir := t.TempDir()
	subdir := filepath.Join(dir, "subflows")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatal(err)
	}

	mainFlow := `- runFlow: subflows/helper.yaml`
	helperFlow := `- tapOn: "Helper"`

	if err := os.WriteFile(filepath.Join(dir, "main.yaml"), []byte(mainFlow), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "helper.yaml"), []byte(helperFlow), 0644); err != nil {
		t.Fatal(err)
	}

	v := New(nil, nil)
	result := v.Validate(filepath.Join(dir, "main.yaml"))

	if !result.IsValid() {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
	if len(result.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(result.Files))
	}
}

func TestValidate_RetryWithFile(t *testing.T) {
	dir := t.TempDir()

	mainFlow := `
- retry:
    maxRetries: "3"
    file: helper.yaml
`
	helperFlow := `- tapOn: "Retry helper"`

	if err := os.WriteFile(filepath.Join(dir, "main.yaml"), []byte(mainFlow), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "helper.yaml"), []byte(helperFlow), 0644); err != nil {
		t.Fatal(err)
	}

	v := New(nil, nil)
	result := v.Validate(filepath.Join(dir, "main.yaml"))

	if !result.IsValid() {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
	if len(result.Files) != 2 {
		t.Errorf("expected 2 files (main + retry file), got %d", len(result.Files))
	}
}

func TestValidate_SharedDependency(t *testing.T) {
	dir := t.TempDir()

	// Both a and b reference shared
	flowA := `- runFlow: shared.yaml`
	flowB := `- runFlow: shared.yaml`
	sharedFlow := `- tapOn: "Shared"`

	if err := os.WriteFile(filepath.Join(dir, "a.yaml"), []byte(flowA), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.yaml"), []byte(flowB), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "shared.yaml"), []byte(sharedFlow), 0644); err != nil {
		t.Fatal(err)
	}

	v := New(nil, nil)
	result := v.Validate(dir)

	if !result.IsValid() {
		t.Errorf("expected valid result, got errors: %v", result.Errors)
	}
	// shared.yaml should only appear once
	sharedCount := 0
	for _, f := range result.Files {
		if strings.HasSuffix(f, "shared.yaml") {
			sharedCount++
		}
	}
	if sharedCount != 1 {
		t.Errorf("expected shared.yaml once, got %d times", sharedCount)
	}
}

func TestResult_IsValid(t *testing.T) {
	r := &Result{}
	if !r.IsValid() {
		t.Error("empty result should be valid")
	}

	r.Errors = append(r.Errors, &ValidationError{File: "test", Message: "error"})
	if r.IsValid() {
		t.Error("result with errors should not be valid")
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		File:    "test.yaml",
		Message: "something went wrong",
	}

	expected := "test.yaml: something went wrong"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}
