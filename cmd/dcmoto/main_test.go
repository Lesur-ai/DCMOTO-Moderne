package main

import (
	"testing"

	"github.com/Lesur-ai/DCMOTO-Moderne/internal/app/config"
)

func TestInitialJoystickKeyboardPreferenceForcesMO5Off(t *testing.T) {
	store, err := config.NewStoreAt(t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	if err := config.PersistJoystickKeyboard(store, true); err != nil {
		t.Fatalf("PersistJoystickKeyboard: %v", err)
	}

	if got := initialJoystickKeyboardPreference("mo5", store); got {
		t.Fatal("initialJoystickKeyboardPreference(mo5) = true, want false even when global preference is true")
	}
	if got := initialJoystickKeyboardPreference("to8d", store); !got {
		t.Fatal("initialJoystickKeyboardPreference(to8d) = false, want persisted true")
	}
}
