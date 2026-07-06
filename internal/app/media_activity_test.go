package app

import (
	"image/color"
	"testing"
	"time"

	"github.com/Lesur-ai/DCMOTO-Moderne/internal/machine"
	"github.com/Lesur-ai/DCMOTO-Moderne/internal/media"
	"github.com/Lesur-ai/DCMOTO-Moderne/internal/overlay"
)

type activityTestDisk struct {
	reads  int
	writes int
}

func (d *activityTestDisk) ReadSector(_, _, _ int) ([256]byte, error) {
	d.reads++
	return [256]byte{}, nil
}

func (d *activityTestDisk) WriteSector(_, _, _ int, _ [256]byte) error {
	d.writes++
	return nil
}

func (d *activityTestDisk) FormatUnit(int) error { return nil }

var _ media.Disk = (*activityTestDisk)(nil)

type activityTestTape struct {
	reads int
}

func (t *activityTestTape) ReadByte() (byte, error) {
	t.reads++
	return 0x42, nil
}

func (t *activityTestTape) WriteByte(byte) error { return nil }
func (t *activityTestTape) Rewind() error        { return nil }
func (t *activityTestTape) Position() int64      { return 0 }

var _ media.Tape = (*activityTestTape)(nil)

func TestWrapDiskActivityMarksOnSectorAccess(t *testing.T) {
	activity := NewMediaActivity()
	disk := &activityTestDisk{}
	wrapped := WrapDiskActivity(disk, activity)
	if activity.active(time.Now()) {
		t.Fatal("activité disque active avant tout accès")
	}

	if _, err := wrapped.ReadSector(0, 0, 1); err != nil {
		t.Fatalf("ReadSector: %v", err)
	}
	if disk.reads != 1 {
		t.Fatalf("ReadSector interne appelé %d fois, want 1", disk.reads)
	}
	if !activity.active(time.Now()) {
		t.Fatal("activité disque inactive juste après ReadSector")
	}
}

func TestWrapTapeActivityMarksOnByteAccess(t *testing.T) {
	activity := NewMediaActivity()
	tape := &activityTestTape{}
	wrapped := WrapTapeActivity(tape, activity)

	if _, err := wrapped.ReadByte(); err != nil {
		t.Fatalf("ReadByte: %v", err)
	}
	if tape.reads != 1 {
		t.Fatalf("ReadByte interne appelé %d fois, want 1", tape.reads)
	}
	if !activity.active(time.Now()) {
		t.Fatal("activité cassette inactive juste après ReadByte")
	}
}

func TestMediaActivityExpires(t *testing.T) {
	activity := NewMediaActivity()
	now := time.Now()
	activity.lastUnixNano.Store(now.Add(-mediaActivityHold - time.Millisecond).UnixNano())
	if activity.active(now) {
		t.Fatal("activité média encore active après expiration")
	}
}

func TestMediaIndicatorColor(t *testing.T) {
	if got, want := mediaIndicatorColor(false, false), (color.RGBA{R: 74, G: 85, B: 104, A: 0xFF}); got != want {
		t.Fatalf("couleur absent = %+v, want %+v", got, want)
	}
	if got, want := mediaIndicatorColor(true, false), (color.RGBA{R: 34, G: 197, B: 94, A: 0xFF}); got != want {
		t.Fatalf("couleur monté = %+v, want %+v", got, want)
	}
	if got, want := mediaIndicatorColor(true, true), (color.RGBA{R: 225, G: 52, B: 52, A: 0xFF}); got != want {
		t.Fatalf("couleur accès = %+v, want %+v", got, want)
	}
}

func TestMediaIndicatorLayoutScalesWithLogicalDisplay(t *testing.T) {
	mo5 := mediaIndicatorLayoutFor(336, 216)
	if mo5.w < 59 || mo5.w > 60 || mo5.h < 13 || mo5.h > 14 || mo5.textFace != mediaIndicatorFaceSmall {
		t.Fatalf("layout MO5 = w:%g h:%g face:%p, want about 59.5x13.6 with small face %p", mo5.w, mo5.h, mo5.textFace, mediaIndicatorFaceSmall)
	}

	gate := mediaIndicatorLayoutFor(672, 432)
	if gate.w != 96 || gate.h != 22 || gate.textFace != mediaIndicatorFace {
		t.Fatalf("layout TO = w:%g h:%g face:%p, want 96x22 with regular face %p", gate.w, gate.h, gate.textFace, mediaIndicatorFace)
	}
}

func TestToggleMediaIndicatorsPersistsAndRefreshesOverlay(t *testing.T) {
	model := &overlay.Model{}
	profile := machine.MachineProfile{ID: "test", Name: "Test"}
	ui := newOverlayUI(profile, nil, model, nil, newUIKit())
	ui.open(profile, ".", machine.Config{}, false, true)

	var persisted []bool
	a := &App{
		overlayUI:              ui,
		mediaIndicatorsEnabled: true,
		onMediaIndicatorsChange: func(enabled bool) {
			persisted = append(persisted, enabled)
		},
	}

	a.toggleMediaIndicators()
	if a.mediaIndicatorsEnabled {
		t.Fatal("mediaIndicatorsEnabled = true après premier toggle, want false")
	}
	if ui.mediaIndicatorsEnabled {
		t.Fatal("overlay mediaIndicatorsEnabled = true après premier toggle, want false")
	}
	if len(persisted) != 1 || persisted[0] {
		t.Fatalf("persisted après premier toggle = %+v, want [false]", persisted)
	}

	a.toggleMediaIndicators()
	if !a.mediaIndicatorsEnabled {
		t.Fatal("mediaIndicatorsEnabled = false après second toggle, want true")
	}
	if !ui.mediaIndicatorsEnabled {
		t.Fatal("overlay mediaIndicatorsEnabled = false après second toggle, want true")
	}
	if len(persisted) != 2 || !persisted[1] {
		t.Fatalf("persisted après second toggle = %+v, want [false true]", persisted)
	}
}

func TestApplyMachineInputDefaultsForcesMO5JoystickOff(t *testing.T) {
	a := &App{joystickKBEnabled: true}
	a.applyMachineInputDefaults(machine.MachineProfile{ID: "mo5"})
	if a.joystickKBEnabled {
		t.Fatal("joystickKBEnabled reste true pour MO5, want false")
	}

	a = &App{joystickKBEnabled: true}
	a.applyMachineInputDefaults(machine.MachineProfile{ID: "to8d"})
	if !a.joystickKBEnabled {
		t.Fatal("joystickKBEnabled forcé à false pour TO8D, want état préservé")
	}
}
