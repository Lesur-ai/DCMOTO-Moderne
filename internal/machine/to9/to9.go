// Package to9 expose le Thomson TO9 derrière le contrat internal/machine.
//
// Le TO9 appartient à la famille gate-array comme le TO9+, mais sa ROM livrée n'a
// pas le même conteneur : 128 Ko de ROM interne/logiciels suivis d'un moniteur
// 8 Ko. Ce premier support câble le BASIC principal 64 Ko et le moniteur 8 Ko sur
// le Device gate-array partagé, avec le chemin clavier ASCII TO9/TO9+.
//
// État volontaire : ce paquet n'est pas encore importé par cmd/dcmoto. Le boot réel
// TO9 ne rend pas avec ce seul câblage ; il faut d'abord modéliser les banques ROM
// logicielles spécifiques au TO9 avant d'exposer la machine au launcher/CLI.
package to9

import (
	"fmt"
	"os"
	"time"

	"github.com/Lesur-ai/DCMOTO-Moderne/internal/cpu6809"
	"github.com/Lesur-ai/DCMOTO-Moderne/internal/engine"
	"github.com/Lesur-ai/DCMOTO-Moderne/internal/keyboard"
	"github.com/Lesur-ai/DCMOTO-Moderne/internal/machine"
	"github.com/Lesur-ai/DCMOTO-Moderne/internal/machine/gatearray"
	"github.com/Lesur-ai/DCMOTO-Moderne/internal/media"
)

const (
	romBasicSize    = 0x10000
	romSoftwareSize = 0x10000
	romMonSize      = 0x2000
	romTotalSize    = romBasicSize + romSoftwareSize + romMonSize
)

type adapter struct {
	*engine.Engine
	ga *gatearray.GateArray
}

var _ machine.Machine = (*adapter)(nil)

func (a *adapter) Reset() {
	a.ga.Reset()
	a.Engine.Reset()
}

func (a *adapter) Initprog()                          { a.ga.Initprog() }
func (a *adapter) SetKey(k machine.Key, pressed bool) { a.ga.SetKey(int(k), pressed) }
func (a *adapter) SetJoystick(j machine.JoystickInput) {
	a.ga.SetJoystick(j.Position, j.Action)
}

func (a *adapter) SetPointer(p machine.PointerInput) {
	x, y := gatearray.PenFromFramebuffer(p.X, p.Y)
	a.ga.SetPointer(x, y, p.Button)
}

func (a *adapter) MountTape(t media.Tape)           { a.ga.MountTape(t) }
func (a *adapter) EjectTape()                       { a.ga.EjectTape() }
func (a *adapter) MountDisk(d media.Disk)           { a.ga.MountDisk(d) }
func (a *adapter) EjectDisk()                       { a.ga.EjectDisk() }
func (a *adapter) MountCartridge(c media.Cartridge) { a.ga.MountCartridge(c) }
func (a *adapter) EjectCartridge()                  { a.ga.EjectCartridge() }
func (a *adapter) MountPrinter(p media.PrinterSink) { a.ga.MountPrinter(p) }
func (a *adapter) EjectPrinter()                    { a.ga.EjectPrinter() }

func (a *adapter) CPUSnapshot() cpu6809.Snapshot  { return a.Engine.CPU().Snapshot() }
func (a *adapter) KeyboardModel() *keyboard.Model { return keyboard.TO9PModel() }

func splitROM(blob []byte) (romBasic, romSoftware, romMon []byte, err error) {
	if len(blob) != romTotalSize {
		return nil, nil, nil, fmt.Errorf("to9: taille ROM inattendue %d octets (attendu %d : BASIC %d + logiciels %d + moniteur %d)",
			len(blob), romTotalSize, romBasicSize, romSoftwareSize, romMonSize)
	}
	romBasic = append([]byte(nil), blob[:romBasicSize]...)
	romSoftware = append([]byte(nil), blob[romBasicSize:romBasicSize+romSoftwareSize]...)
	romMon = append([]byte(nil), blob[romBasicSize+romSoftwareSize:]...)
	return romBasic, romSoftware, romMon, nil
}

func newFromROM(blob []byte, _ time.Time) (machine.Machine, error) {
	romBasic, _, romMon, err := splitROM(blob)
	if err != nil {
		return nil, err
	}
	ga := gatearray.NewTO9(romMon, romBasic)
	eng := engine.New(ga, 0)
	ga.AttachCPU(eng.CPU())
	ga.AttachBeam(eng.VideoBeam)

	a := &adapter{Engine: eng, ga: ga}
	a.Reset()
	return a, nil
}

func init() {
	machine.Register(machine.MachineProfile{
		ID:     "to9",
		Name:   "Thomson TO9",
		Family: machine.FamilyTOGateArray,
		Params: []machine.Param{
			{Key: machine.KeyROM, Label: "ROM TO9 (BASIC/logiciels + moniteur)", Kind: machine.ParamFile, FileExt: []string{".rom"}, Required: true},
			{Key: machine.KeyTape, Label: "Cassette", Kind: machine.ParamFile, FileExt: []string{".k7"}, LiveMutable: true},
			{Key: machine.KeyDisk, Label: "Disquette", Kind: machine.ParamFile, FileExt: []string{".fd"}, LiveMutable: true},
			{Key: machine.KeyCart, Label: "Cartouche", Kind: machine.ParamFile, FileExt: []string{".rom"}, LiveMutable: true},
		},
		New: newFromConfig,
	})
}

func newFromConfig(cfg machine.Config) (machine.Machine, error) {
	path, _ := cfg[machine.KeyROM].(string)
	if path == "" {
		return nil, fmt.Errorf("to9: ROM TO9 requise (paramètre %q)", machine.KeyROM)
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("to9: ROM: %w", err)
	}
	return newFromROM(blob, time.Now())
}
