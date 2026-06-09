package image

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/weathericons-regular-webfont.ttf
var weathericonsTTF []byte

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

	goBoldFont, err := opentype.Parse(gobold.TTF)
	if err != nil {
		log.Fatal("failed to parse go bold font: " + err.Error())
	}

	goRegularFont, err := opentype.Parse(goregular.TTF)
	if err != nil {
		log.Fatal("failed to parse go regular font: " + err.Error())
	}

	titleFont, err = opentype.NewFace(goBoldFont, &opentype.FaceOptions{
		Size:    20,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	infoFont, err = opentype.NewFace(goRegularFont, &opentype.FaceOptions{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingFull,
	})
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
		drawText(img, (i*99)+24, 95, fmt.Sprintf("%d°C", w.Low), infoFont, isHot(w.Low))
		drawText(img, (i*99)+74, 95, fmt.Sprintf("%d°C", w.High), infoFont, isHot(w.High))
		drawText(img, (i*99)+28, 115, fmtRain(w.PrecipitationAmount), infoFont, false)
		drawText(img, (i*99)+74, 115, fmt.Sprintf("%d%%", w.PrecipitationProbability), infoFont, false)
	}
	// separators
	for i := range len(weather) {
		x := i*99 - 1
		for y := 10; y < 118; y++ {
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
