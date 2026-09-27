package image

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/weathericons-regular-webfont.ttf
var weathericonsTTF []byte

// The display is 1-bit (black/red/white), so anti-aliased vector fonts turn
// into ragged edges after thresholding. Both fonts below are pixel fonts whose
// outlines land exactly on the pixel grid at their rendered size (27 px and
// 20 px at 72 dpi), so they rasterize with no anti-aliasing and no upscaling -
// see TestFontGlyphCoverage and TestFontsRenderWithoutAntiAliasing.
//
// Title: Jersey 15 (OFL 1.1), whose native design size is 27 px. At no other
// size do its outlines align with the pixel grid.
//
//go:embed fonts/jersey15-title-subset.ttf
var titleTTF []byte

// Info: ESL Pixel Info (OFL 1.1), a custom pixel font generated for this
// project, 100 font units per design pixel. It carries the glyphs the layout
// draws - %, -, ., 0-9, C, m and ° - plus a space no format string currently
// emits, with 15 px tall digits on 2 px strokes, a half-height "mm" unit and a
// 2x2 period. Native design size 20 px.
//
//go:embed fonts/esl-pixel-info.ttf
var infoTTF []byte

var weathericonsFont *opentype.Font
var titleFont font.Face
var infoFont font.Face

var colorRed = image.NewUniform(color.RGBA{R: 255, G: 0, B: 0, A: 255})

type Weather struct {
	High                     int
	Icon                     string
	Low                      int
	Day                      string
	PrecipitationProbability int
	PrecipitationAmount      float64
}

func init() {
	var err error
	weathericonsFont, err = opentype.Parse(weathericonsTTF)
	if err != nil {
		log.Fatal("failed to parse weathericons font: " + err.Error())
	}

	titleFontData, err := opentype.Parse(titleTTF)
	if err != nil {
		log.Fatal("failed to parse title font: " + err.Error())
	}

	infoFontData, err := opentype.Parse(infoTTF)
	if err != nil {
		log.Fatal("failed to parse info font: " + err.Error())
	}

	titleFont, err = opentype.NewFace(titleFontData, &opentype.FaceOptions{
		Size:    27,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal("failed to create title face: " + err.Error())
	}
	infoFont, err = opentype.NewFace(infoFontData, &opentype.FaceOptions{
		Size:    20,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal("failed to create info face: " + err.Error())
	}
}

func drawText(img *image.RGBA, x, y int, text string, face font.Face, red bool) {
	src := image.Black
	if red {
		src = colorRed
	}
	m := face.Metrics()
	width := font.MeasureString(face, text)
	baseline := y + (m.Ascent-m.Descent).Round()/2
	d := &font.Drawer{
		Dst:  img,
		Src:  src,
		Face: face,
		Dot:  fixed.P(x-width.Round()/2, baseline),
	}
	d.DrawString(text)
}

func toBytes(img image.Image) ([]byte, []byte, error) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	black := make([]byte, w*h)
	red := make([]byte, w*h)
	for y := range h {
		for x := range w {
			r, g, b, _ := img.At(x, y).RGBA()
			pos := y*w + x
			switch {
			case r == g && r == b:
				if r < 0x8000 { // dark pixel = set byte to 1
					black[pos] = 1
				}
			case r > g && r > b:
				if (uint32(g)+uint32(b))/2 < 0x8000 { // red pixel = set byte to 1
					red[pos] = 1
				}
			default:
				return nil, nil, fmt.Errorf("unsupported pixel color: r=%d, g=%d, b=%d", r, g, b)
			}
		}
	}
	return black, red, nil
}

func drawIcon(iconStr string, img *image.RGBA, x, y int, red bool) {
	yOffset := 0
	size := 48.0
	var icon string
	switch iconStr {
	case "clear":
		icon = "\uF00D"
		yOffset = 3
	case "partly-cloudy":
		icon = "\uF002"
		yOffset = 10
	case "cloudy":
		icon = "\uF013"
		yOffset = 10
		size = 56
	case "fog":
		icon = "\uF014"
	case "wind":
		icon = "\uF050"
	case "rain":
		icon = "\uF019"
	case "sleet":
		icon = "\uF0B5"
	case "snow":
		icon = "\uF01B"
	case "hail":
		icon = "\uF015"
	case "thunderstorm":
		icon = "\uF01E"
	default:
		icon = "\uF07B"
	}

	src := image.Black
	if red {
		src = image.NewUniform(color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}
	face, err := opentype.NewFace(weathericonsFont, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic("failed to create weathericons face: " + err.Error())
	}
	advance := font.MeasureString(face, icon)
	d := &font.Drawer{
		Dst:  img,
		Src:  src,
		Face: face,
		Dot:  fixed.P(x+(99-advance.Round())/2, y+yOffset),
	}
	d.DrawString(icon) // cloud icon
}

func Generate(weather ...Weather) ([]byte, []byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 296, 128))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	for i, w := range weather {
		drawText(img, (i*99)+49, 16, w.Day, titleFont, false)
		drawIcon(w.Icon, img, (i * 99), 70, false)
		drawText(img, (i*99)+24, 98, fmt.Sprintf("%d°C", w.Low), infoFont, isHot(w.Low))
		drawText(img, (i*99)+74, 98, fmt.Sprintf("%d°C", w.High), infoFont, isHot(w.High))
		drawText(img, (i*99)+26, 118, fmtRain(w.PrecipitationAmount), infoFont, false)
		drawText(img, (i*99)+74, 118, fmt.Sprintf("%d%%", w.PrecipitationProbability), infoFont, false)
	}
	// separators, running from the title down to the last row of the info
	// text so they end flush with it
	for i := range len(weather) {
		x := i*99 - 1
		for y := 10; y < 124; y++ {
			img.Set(x, y, colorRed)
		}
	}
	return toBytes(img)
}

func fmtRain(p float64) string {
	if p < 10.0 {
		return fmt.Sprintf("%.1fmm", p)
	}
	return fmt.Sprintf("%.0fmm", p)
}

func isHot(temp int) bool {
	return temp >= 20
}
