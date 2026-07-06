package app

import (
	"sync/atomic"
	"time"

	"github.com/Lesur-ai/DCMOTO-Moderne/internal/media"
)

const mediaActivityHold = 180 * time.Millisecond

type MediaActivity struct {
	lastUnixNano atomic.Int64
}

func NewMediaActivity() *MediaActivity { return &MediaActivity{} }

func (a *MediaActivity) mark() {
	if a != nil {
		a.lastUnixNano.Store(time.Now().UnixNano())
	}
}

func (a *MediaActivity) active(now time.Time) bool {
	if a == nil {
		return false
	}
	last := a.lastUnixNano.Load()
	return last != 0 && now.Sub(time.Unix(0, last)) <= mediaActivityHold
}

type activityTape struct {
	inner    media.Tape
	activity *MediaActivity
}

func WrapTapeActivity(t media.Tape, activity *MediaActivity) media.Tape {
	if t == nil || activity == nil {
		return t
	}
	return &activityTape{inner: t, activity: activity}
}

func (t *activityTape) ReadByte() (byte, error) {
	t.activity.mark()
	return t.inner.ReadByte()
}

func (t *activityTape) WriteByte(b byte) error {
	t.activity.mark()
	return t.inner.WriteByte(b)
}

func (t *activityTape) Rewind() error {
	t.activity.mark()
	return t.inner.Rewind()
}

func (t *activityTape) Position() int64 { return t.inner.Position() }

type activityDisk struct {
	inner    media.Disk
	activity *MediaActivity
}

func WrapDiskActivity(d media.Disk, activity *MediaActivity) media.Disk {
	if d == nil || activity == nil {
		return d
	}
	return &activityDisk{inner: d, activity: activity}
}

func (d *activityDisk) ReadSector(unit, track, sector int) ([256]byte, error) {
	d.activity.mark()
	return d.inner.ReadSector(unit, track, sector)
}

func (d *activityDisk) WriteSector(unit, track, sector int, data [256]byte) error {
	d.activity.mark()
	return d.inner.WriteSector(unit, track, sector, data)
}

func (d *activityDisk) FormatUnit(unit int) error {
	d.activity.mark()
	return d.inner.FormatUnit(unit)
}
