package to9

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lesur-ai/DCMOTO-Moderne/internal/keyboard"
	"github.com/Lesur-ai/DCMOTO-Moderne/internal/machine"
)

func romTestPath() string { return filepath.Join("..", "..", "..", "rom", "to9.rom") }

var testBootDate = time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC)

func TestProfileRegistered(t *testing.T) {
	p, ok := machine.ByID("to9")
	if !ok {
		t.Fatal("profil to9 non enregistré")
	}
	if p.Name != "Thomson TO9" || p.Family != machine.FamilyTOGateArray {
		t.Fatalf("profil to9 = {Name:%q Family:%d}", p.Name, p.Family)
	}
}

func TestNewFromConfigErrors(t *testing.T) {
	if _, err := newFromConfig(machine.Config{}); err == nil {
		t.Error("ROM absente : erreur attendue")
	}
	if _, err := newFromConfig(machine.Config{machine.KeyROM: "/inexistant.rom"}); err == nil {
		t.Error("ROM introuvable : erreur attendue")
	}
	if _, err := newFromROM(make([]byte, 1024), testBootDate); err == nil {
		t.Error("taille ROM invalide : erreur attendue")
	}
}

func TestSplitROMCopiesAndLayout(t *testing.T) {
	blob := make([]byte, romTotalSize)
	blob[0] = 0x42
	blob[romBasicSize] = 0x43
	blob[romBasicSize+romSoftwareSize] = 0x44

	romBasic, romSoftware, romMon, err := splitROM(blob)
	if err != nil {
		t.Fatalf("splitROM: %v", err)
	}
	if len(romBasic) != romBasicSize || len(romSoftware) != romSoftwareSize || len(romMon) != romMonSize {
		t.Fatalf("tailles split = basic %d software %d mon %d", len(romBasic), len(romSoftware), len(romMon))
	}
	blob[0], blob[romBasicSize], blob[romBasicSize+romSoftwareSize] = 0xaa, 0xbb, 0xcc
	if romBasic[0] == 0xaa || romSoftware[0] == 0xbb || romMon[0] == 0xcc {
		t.Fatal("splitROM aliase le blob appelant au lieu de copier les segments")
	}
}

func TestSplitROMMatchesTrackedReference(t *testing.T) {
	blob, err := os.ReadFile(romTestPath())
	if err != nil {
		t.Fatalf("lecture ROM TO9 : %v", err)
	}
	romBasic, romSoftware, romMon, err := splitROM(blob)
	if err != nil {
		t.Fatalf("split ROM réelle : %v", err)
	}
	if string(romBasic[:16]) != " BASIC 128 MICRO" {
		t.Fatalf("signature BASIC TO9 inattendue : %q", string(romBasic[:16]))
	}
	if string(romSoftware[:16]) != " FICHES & DOSSIE" {
		t.Fatalf("signature logiciel TO9 inattendue : %q", string(romSoftware[:16]))
	}
	reset := uint16(romMon[0x1ffe])<<8 | uint16(romMon[0x1fff])
	if reset != 0xec19 {
		t.Fatalf("vecteur reset = 0x%04x, attendu 0xEC19 pour rom/to9.rom", reset)
	}
}

func TestNewFromROMWiresROMIntoGateArray(t *testing.T) {
	blob, err := os.ReadFile(romTestPath())
	if err != nil {
		t.Fatalf("lecture ROM TO9 : %v", err)
	}
	romBasic, _, romMon, err := splitROM(blob)
	if err != nil {
		t.Fatalf("split ROM réelle : %v", err)
	}
	m, err := newFromROM(blob, testBootDate)
	if err != nil {
		t.Fatalf("newFromROM: %v", err)
	}
	if w, h := m.FrameSize(); w != 672 || h != 216 {
		t.Fatalf("FrameSize = %dx%d, attendu 672x216", w, h)
	}
	if km := m.KeyboardModel(); km != keyboard.TO9PModel() {
		t.Fatalf("KeyboardModel = %+v, attendu TO9PModel %+v", km, keyboard.TO9PModel())
	}
	a, ok := m.(*adapter)
	if !ok {
		t.Fatalf("machine concrète = %T, attendu *adapter", m)
	}
	if got, want := a.ga.Read8(0xfffe), romMon[0x1ffe]; got != want {
		t.Fatalf("moniteur câblé à 0xFFFE = 0x%02x, attendu romMon[0x1ffe]=0x%02x", got, want)
	}
	a.ga.Write8(0xe7c3, 0x04)
	if got, want := a.ga.Read8(0x0000), romBasic[0]; got != want {
		t.Fatalf("BASIC câblé à 0x0000 = 0x%02x, attendu romBasic[0]=0x%02x", got, want)
	}
	a.ga.SetKey(0x02, true)
	if got := a.ga.Read8(0xe7de); got != 0x01 {
		t.Fatalf("TO9 E7DE après frappe = 0x%02x, attendu 0x01", got)
	}
	if got := a.ga.Read8(0xe7df); got != 0x59 {
		t.Fatalf("TO9 E7DF après frappe Y = 0x%02x, attendu 0x59", got)
	}
}

func TestResetVectorWired(t *testing.T) {
	blob, err := os.ReadFile(romTestPath())
	if err != nil {
		t.Fatalf("lecture ROM TO9 : %v", err)
	}
	m, err := newFromROM(blob, testBootDate)
	if err != nil {
		t.Fatalf("newFromROM: %v", err)
	}
	if pc := m.CPUSnapshot().PC; pc != 0xEC19 {
		t.Fatalf("PC au reset = 0x%04x, attendu le vecteur reset TO9 0xEC19", pc)
	}
}
