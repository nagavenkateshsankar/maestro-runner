package core

import (
	"testing"
	"time"
)

func TestBounds_Center(t *testing.T) {
	tests := []struct {
		bounds    Bounds
		expectedX int
		expectedY int
	}{
		{Bounds{X: 0, Y: 0, Width: 100, Height: 100}, 50, 50},
		{Bounds{X: 10, Y: 20, Width: 100, Height: 200}, 60, 120},
		{Bounds{X: 0, Y: 0, Width: 0, Height: 0}, 0, 0},
	}

	for _, tt := range tests {
		x, y := tt.bounds.Center()
		if x != tt.expectedX || y != tt.expectedY {
			t.Errorf("Bounds%+v.Center() = (%d, %d), want (%d, %d)",
				tt.bounds, x, y, tt.expectedX, tt.expectedY)
		}
	}
}

func TestBounds_Contains(t *testing.T) {
	bounds := Bounds{X: 10, Y: 10, Width: 100, Height: 100}

	tests := []struct {
		x, y     int
		expected bool
	}{
		{50, 50, true},    // Center
		{10, 10, true},    // Top-left corner
		{109, 109, true},  // Just inside bottom-right
		{110, 110, false}, // Exactly at boundary (exclusive)
		{0, 0, false},     // Outside
		{200, 200, false}, // Far outside
	}

	for _, tt := range tests {
		if got := bounds.Contains(tt.x, tt.y); got != tt.expected {
			t.Errorf("Bounds.Contains(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.expected)
		}
	}
}

func TestCommandResult_Fields(t *testing.T) {
	result := CommandResult{
		Success:  true,
		Duration: 100 * time.Millisecond,
		Message:  "Tapped on button",
		Element: &ElementInfo{
			ID:      "btn-submit",
			Text:    "Submit",
			Visible: true,
			Enabled: true,
		},
	}

	if !result.Success {
		t.Error("Success should be true")
	}
	if result.Duration != 100*time.Millisecond {
		t.Errorf("Duration = %v, want 100ms", result.Duration)
	}
	if result.Element == nil {
		t.Error("Element should not be nil")
	}
	if result.Element.ID != "btn-submit" {
		t.Errorf("Element.ID = %s, want btn-submit", result.Element.ID)
	}
}

func TestElementInfo_Fields(t *testing.T) {
	elem := ElementInfo{
		ID:                 "elem-1",
		Text:               "Hello",
		Bounds:             Bounds{X: 10, Y: 20, Width: 100, Height: 50},
		Visible:            true,
		Enabled:            true,
		Focused:            false,
		Checked:            true,
		Selected:           false,
		Class:              "android.widget.Button",
		AccessibilityLabel: "Submit Button",
		Attributes: map[string]string{
			"resource-id": "com.app:id/submit",
		},
	}

	if elem.ID != "elem-1" {
		t.Errorf("ID = %s, want elem-1", elem.ID)
	}
	if elem.Bounds.Width != 100 {
		t.Errorf("Bounds.Width = %d, want 100", elem.Bounds.Width)
	}
	if !elem.Visible {
		t.Error("Visible should be true")
	}
	if !elem.Checked {
		t.Error("Checked should be true")
	}
	if elem.Attributes["resource-id"] != "com.app:id/submit" {
		t.Errorf("Attributes[resource-id] = %s, want com.app:id/submit", elem.Attributes["resource-id"])
	}
}

func TestStateSnapshot_Fields(t *testing.T) {
	state := StateSnapshot{
		AppState:        "foreground",
		Orientation:     "portrait",
		KeyboardVisible: true,
		FocusedElement: &ElementInfo{
			ID: "input-email",
		},
		ClipboardText:   "copied text",
		CurrentActivity: "com.app.MainActivity",
		CurrentScreen:   "LoginScreen",
	}

	if state.AppState != "foreground" {
		t.Errorf("AppState = %s, want foreground", state.AppState)
	}
	if !state.KeyboardVisible {
		t.Error("KeyboardVisible should be true")
	}
	if state.FocusedElement == nil {
		t.Error("FocusedElement should not be nil")
	}
	if state.ClipboardText != "copied text" {
		t.Errorf("ClipboardText = %s, want 'copied text'", state.ClipboardText)
	}
}

func TestPlatformInfo_Fields(t *testing.T) {
	info := PlatformInfo{
		Platform:     "android",
		OSVersion:    "14",
		DeviceName:   "Pixel 8",
		DeviceID:     "emulator-5554",
		IsSimulator:  true,
		ScreenWidth:  1080,
		ScreenHeight: 2400,
		AppID:        "com.example.app",
		AppVersion:   "1.2.3",
	}

	if info.Platform != "android" {
		t.Errorf("Platform = %s, want android", info.Platform)
	}
	if !info.IsSimulator {
		t.Error("IsSimulator should be true")
	}
	if info.ScreenWidth != 1080 {
		t.Errorf("ScreenWidth = %d, want 1080", info.ScreenWidth)
	}
}

func TestExecutedByConstants(t *testing.T) {
	if ExecutedByDriver != "driver" {
		t.Errorf("ExecutedByDriver = %s, want driver", ExecutedByDriver)
	}
	if ExecutedByRunner != "runner" {
		t.Errorf("ExecutedByRunner = %s, want runner", ExecutedByRunner)
	}
}

func TestLogEntry_Fields(t *testing.T) {
	now := time.Now()
	entry := LogEntry{
		Timestamp: now,
		Level:     "error",
		Source:    "device",
		Message:   "App crashed",
	}

	if entry.Timestamp != now {
		t.Error("Timestamp mismatch")
	}
	if entry.Level != "error" {
		t.Errorf("Level = %s, want error", entry.Level)
	}
	if entry.Source != "device" {
		t.Errorf("Source = %s, want device", entry.Source)
	}
	if entry.Message != "App crashed" {
		t.Errorf("Message = %s, want 'App crashed'", entry.Message)
	}
}
