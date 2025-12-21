// Package validator validates Maestro flow files before execution.
// It parses all files upfront, resolves runFlow references, and detects errors.
package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/devicelab-dev/maestro-runner/pkg/flow"
)

// ValidationError represents a validation error with context.
type ValidationError struct {
	File    string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.File, e.Message)
}

// Result contains the validation result.
type Result struct {
	// Files is the list of flow file paths in execution order.
	Files []string
	// Errors contains all validation errors found.
	Errors []error
}

// IsValid returns true if there are no validation errors.
func (r *Result) IsValid() bool {
	return len(r.Errors) == 0
}

// Validator validates flow files.
type Validator struct {
	includeTags []string
	excludeTags []string
}

// New creates a new Validator.
func New(includeTags, excludeTags []string) *Validator {
	return &Validator{
		includeTags: includeTags,
		excludeTags: excludeTags,
	}
}

// Validate validates a file or directory.
// It parses all flows, resolves runFlow references, and returns validation results.
func (v *Validator) Validate(path string) *Result {
	result := &Result{}

	info, err := os.Stat(path)
	if err != nil {
		result.Errors = append(result.Errors, &ValidationError{
			File:    path,
			Message: fmt.Sprintf("cannot access: %v", err),
		})
		return result
	}

	var files []string
	if info.IsDir() {
		files, err = v.collectFlowFiles(path)
		if err != nil {
			result.Errors = append(result.Errors, &ValidationError{
				File:    path,
				Message: fmt.Sprintf("failed to scan directory: %v", err),
			})
			return result
		}
	} else {
		files = []string{path}
	}

	// Validate each file and resolve dependencies
	validated := make(map[string]bool)
	for _, file := range files {
		v.validateFile(file, result, validated, nil)
	}

	return result
}

// collectFlowFiles finds all .yaml/.yml files in a directory.
func (v *Validator) collectFlowFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yaml" || ext == ".yml" {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// validateFile validates a single file and its runFlow dependencies.
func (v *Validator) validateFile(filePath string, result *Result, validated map[string]bool, chain []string) {
	// Check for circular dependency
	for _, ancestor := range chain {
		if ancestor == filePath {
			cycle := append(chain, filePath)
			result.Errors = append(result.Errors, &ValidationError{
				File:    filePath,
				Message: fmt.Sprintf("circular dependency detected: %s", strings.Join(cycle, " -> ")),
			})
			return
		}
	}

	// Skip if already validated
	if validated[filePath] {
		return
	}

	// Parse the file
	f, err := flow.ParseFile(filePath)
	if err != nil {
		result.Errors = append(result.Errors, &ValidationError{
			File:    filePath,
			Message: fmt.Sprintf("parse error: %v", err),
		})
		return
	}

	// Check tag filters (only for top-level files, not runFlow targets)
	if len(chain) == 0 && !flow.ShouldIncludeFlow(f, v.includeTags, v.excludeTags) {
		return
	}

	// Mark as validated and add to result
	validated[filePath] = true
	result.Files = append(result.Files, filePath)

	// Recursively validate runFlow dependencies
	newChain := append(chain, filePath)
	v.validateRunFlowSteps(f.Steps, filePath, result, validated, newChain)

	// Also validate lifecycle hooks
	v.validateRunFlowSteps(f.Config.OnFlowStart, filePath, result, validated, newChain)
	v.validateRunFlowSteps(f.Config.OnFlowComplete, filePath, result, validated, newChain)
}

// validateRunFlowSteps finds and validates runFlow references in steps.
func (v *Validator) validateRunFlowSteps(steps []flow.Step, parentFile string, result *Result, validated map[string]bool, chain []string) {
	parentDir := filepath.Dir(parentFile)

	for _, step := range steps {
		switch s := step.(type) {
		case *flow.RunFlowStep:
			if s.File != "" {
				refPath := resolveFilePath(parentDir, s.File)
				v.validateFile(refPath, result, validated, chain)
			}
			// Also check inline commands
			v.validateRunFlowSteps(s.Steps, parentFile, result, validated, chain)

		case *flow.RepeatStep:
			v.validateRunFlowSteps(s.Steps, parentFile, result, validated, chain)

		case *flow.RetryStep:
			if s.File != "" {
				refPath := resolveFilePath(parentDir, s.File)
				v.validateFile(refPath, result, validated, chain)
			}
			v.validateRunFlowSteps(s.Steps, parentFile, result, validated, chain)
		}
	}
}

// resolveFilePath resolves a file path relative to a base directory.
func resolveFilePath(baseDir, filePath string) string {
	if filepath.IsAbs(filePath) {
		return filePath
	}
	return filepath.Join(baseDir, filePath)
}

