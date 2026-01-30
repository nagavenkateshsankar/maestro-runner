package simulator

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/devicelab-dev/maestro-runner/pkg/logger"
)

// FindSimctlBinary verifies that xcrun/simctl is available.
func FindSimctlBinary() (string, error) {
	path, err := exec.LookPath("xcrun")
	if err != nil {
		return "", fmt.Errorf("xcrun not found; install Xcode Command Line Tools: xcode-select --install")
	}
	return path, nil
}

// simctlDevicesOutput represents the JSON output from xcrun simctl list devices.
type simctlDevicesOutput struct {
	Devices map[string][]simctlDevice `json:"devices"`
}

type simctlDevice struct {
	Name        string `json:"name"`
	UDID        string `json:"udid"`
	State       string `json:"state"`
	IsAvailable bool   `json:"isAvailable"`
}

// ListSimulators returns all available iOS simulators.
func ListSimulators() ([]SimulatorDevice, error) {
	if _, err := FindSimctlBinary(); err != nil {
		return nil, err
	}

	cmd := exec.Command("xcrun", "simctl", "list", "devices", "available", "-j")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list simulators: %w", err)
	}

	var data simctlDevicesOutput
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse simctl output: %w", err)
	}

	var sims []SimulatorDevice
	for runtime, devices := range data.Devices {
		osVersion := extractOSVersion(runtime)
		for _, dev := range devices {
			if !dev.IsAvailable {
				continue
			}
			sims = append(sims, SimulatorDevice{
				Name:        dev.Name,
				UDID:        dev.UDID,
				Runtime:     runtime,
				OSVersion:   osVersion,
				State:       dev.State,
				IsAvailable: dev.IsAvailable,
			})
		}
	}

	logger.Debug("Found %d available simulators", len(sims))
	return sims, nil
}

// ListShutdownSimulators returns available simulators that are currently shut down.
func ListShutdownSimulators() ([]SimulatorDevice, error) {
	sims, err := ListSimulators()
	if err != nil {
		return nil, err
	}

	var shutdown []SimulatorDevice
	for _, sim := range sims {
		if sim.State == "Shutdown" {
			shutdown = append(shutdown, sim)
		}
	}
	return shutdown, nil
}

// IsSimulator checks if a UDID belongs to a known simulator.
func IsSimulator(udid string) bool {
	sims, err := ListSimulators()
	if err != nil {
		return false
	}
	for _, sim := range sims {
		if sim.UDID == udid {
			return true
		}
	}
	return false
}

// CheckBootStatus checks if a simulator is booted.
func CheckBootStatus(udid string) (*BootStatus, error) {
	sims, err := ListSimulators()
	if err != nil {
		return nil, err
	}
	for _, sim := range sims {
		if sim.UDID == udid {
			return &BootStatus{Booted: sim.State == "Booted"}, nil
		}
	}
	return nil, fmt.Errorf("simulator not found: %s", udid)
}

// WaitForBoot waits for a simulator to reach "Booted" state.
func WaitForBoot(udid string, timeout time.Duration) error {
	logger.Info("Waiting for simulator boot: %s", udid)
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		status, err := CheckBootStatus(udid)
		if err != nil {
			logger.Debug("Boot check error: %v", err)
			<-ticker.C
			continue
		}
		if status.IsReady() {
			logger.Info("Simulator booted: %s", udid)
			return nil
		}
		<-ticker.C
	}

	return fmt.Errorf("simulator boot timeout after %v", timeout)
}

// BootSimulator boots an iOS simulator and waits for it to be ready.
func BootSimulator(udid string, timeout time.Duration) error {
	logger.Info("Booting simulator: %s", udid)

	cmd := exec.Command("xcrun", "simctl", "boot", udid)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Check if already booted
		if strings.Contains(string(output), "current state: Booted") {
			logger.Info("Simulator already booted: %s", udid)
			return nil
		}
		return fmt.Errorf("failed to boot simulator: %s", strings.TrimSpace(string(output)))
	}

	// Wait for boot to complete
	if err := WaitForBoot(udid, timeout); err != nil {
		return err
	}

	// Open the Simulator UI
	openCmd := exec.Command("open", "-a", "Simulator")
	if err := openCmd.Run(); err != nil {
		logger.Debug("Failed to open Simulator app: %v", err)
	}

	return nil
}

// ShutdownSimulator gracefully shuts down a simulator.
func ShutdownSimulator(udid string, timeout time.Duration) error {
	logger.Info("Shutting down simulator: %s", udid)

	cmd := exec.Command("xcrun", "simctl", "shutdown", udid)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Check if already shutdown
		if strings.Contains(string(output), "current state: Shutdown") {
			logger.Info("Simulator already shutdown: %s", udid)
			return nil
		}
		logger.Warn("simctl shutdown failed for %s: %s", udid, strings.TrimSpace(string(output)))
	}

	// Poll until shutdown confirmed
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		status, err := CheckBootStatus(udid)
		if err != nil || !status.Booted {
			logger.Info("Simulator shutdown confirmed: %s", udid)
			return nil
		}
		<-ticker.C
	}

	return fmt.Errorf("simulator shutdown timeout after %v", timeout)
}

// extractOSVersion extracts version from runtime string.
// e.g., "com.apple.CoreSimulator.SimRuntime.iOS-17-2" → "17.2"
func extractOSVersion(runtime string) string {
	// Find "iOS-" prefix and extract version
	idx := strings.LastIndex(runtime, "iOS-")
	if idx == -1 {
		// Try other platforms (watchOS, tvOS, visionOS)
		for _, prefix := range []string{"watchOS-", "tvOS-", "xrOS-"} {
			idx = strings.LastIndex(runtime, prefix)
			if idx != -1 {
				version := runtime[idx+len(prefix):]
				return strings.ReplaceAll(version, "-", ".")
			}
		}
		return ""
	}
	version := runtime[idx+4:] // skip "iOS-"
	return strings.ReplaceAll(version, "-", ".")
}
