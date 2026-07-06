package app

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/goregular"
)

var (
	mediaIndicatorFace      = *loadFace(goregular.TTF, 10)
	mediaIndicatorFaceSmall = *loadFace(goregular.TTF, 8)
)

type mediaIndicatorLayout struct {
	w, h, margin, spacing              float32
	labelX, labelY, circleDX, circleDY float32
	outerR, innerR                     float32
	textFace                           text.Face
}

func mediaIndicatorLayoutFor(logW, logH int) mediaIndicatorLayout {
	scale := float32(logW) / 672
	if hScale := float32(logH) / 432; hScale < scale {
		scale = hScale
	}
	if scale <= 0 || scale > 1 {
		scale = 1
	}
	textFace := mediaIndicatorFace
	if scale < 0.75 {
		scale = 0.62
		textFace = mediaIndicatorFaceSmall
	}
	return mediaIndicatorLayout{
		w:        96 * scale,
		h:        22 * scale,
		margin:   6 * scale,
		spacing:  44 * scale,
		labelX:   9 * scale,
		labelY:   5 * scale,
		circleDX: 27 * scale,
		circleDY: 7 * scale,
		outerR:   5.5 * scale,
		innerR:   4.3 * scale,
		textFace: textFace,
	}
}

func (a *App) drawMediaIndicators(screen *ebiten.Image) {
	bounds := screen.Bounds()
	layout := mediaIndicatorLayoutFor(bounds.Dx(), bounds.Dy())
	sw := float32(bounds.Dx())
	x := sw - layout.w - layout.margin
	if x < layout.margin {
		x = layout.margin
	}
	y := layout.margin
	vector.DrawFilledRect(screen, x, y, layout.w, layout.h, color.RGBA{R: 0, G: 0, B: 0, A: 150}, false)

	now := time.Now()
	a.drawSingleMediaIndicator(screen, layout, x+layout.labelX, y+layout.labelY, "K7", mediaMounted(a.tapeName), a.tapeActivity, now)
	a.drawSingleMediaIndicator(screen, layout, x+layout.labelX+layout.spacing, y+layout.labelY, "FD", mediaMounted(a.diskName), a.diskActivity, now)
}

func (a *App) drawSingleMediaIndicator(screen *ebiten.Image, layout mediaIndicatorLayout, x, y float32, label string, mounted bool, activity *MediaActivity, now time.Time) {
	drawText(screen, label, x, y, layout.textFace, color.RGBA{R: 235, G: 238, B: 244, A: 0xFF})
	cx := x + layout.circleDX
	cy := y + layout.circleDY
	vector.FillCircle(screen, cx, cy, layout.outerR, color.RGBA{R: 20, G: 24, B: 30, A: 0xFF}, true)
	vector.FillCircle(screen, cx, cy, layout.innerR, mediaIndicatorColor(mounted, mounted && activity.active(now)), true)
}

func drawText(screen *ebiten.Image, s string, x, y float32, face text.Face, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, s, face, op)
}

func mediaMounted(name string) bool {
	return name != "" && name != "."
}

func mediaIndicatorColor(mounted, active bool) color.RGBA {
	switch {
	case active:
		return color.RGBA{R: 225, G: 52, B: 52, A: 0xFF}
	case mounted:
		return color.RGBA{R: 34, G: 197, B: 94, A: 0xFF}
	default:
		return color.RGBA{R: 74, G: 85, B: 104, A: 0xFF}
	}
}
