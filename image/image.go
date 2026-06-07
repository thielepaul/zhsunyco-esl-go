package image

import (
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"log"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed fonts/weathericons-regular-webfont.ttf
var weathericonsTTF []byte

var weathericonsFont *opentype.Font

var colorRed = image.NewUniform(color.RGBA{R: 255, G: 0, B: 0, A: 255})

type Weather struct {
	High int
	Icon string
	Low  int
	Text string
}

func init() {
	var err error
	weathericonsFont, err = opentype.Parse(weathericonsTTF)
	if err != nil {
		log.Fatal("failed to parse weathericons font: " + err.Error())
	}
}

func drawText(img *image.RGBA, x, y int, text string, red bool) {
	src := image.Black
	if red {
		src = colorRed
	}
	d := &font.Drawer{
		Dst:  img,
		Src:  src,
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
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
				if r/3+g/3+b/3 < 0x8000 { // red pixel = set byte to 1
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
	var icon string
	switch iconStr {
	case "clear":
		icon = "\uF00D"
	case "partly-cloudy":
		icon = "\uF002"
	case "cloudy":
		icon = "\uF013"
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
		Size:    48,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic("failed to create weathericons face: " + err.Error())
	}
	d := &font.Drawer{
		Dst:  img,
		Src:  src,
		Face: face,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(icon) // cloud icon
}

func Generate(weather ...Weather) ([]byte, []byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 296, 128))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	for i, w := range weather {
		drawIcon(w.Icon, img, (i * 99), 60, false)
		drawText(img, (i * 99), 120, w.Text, true)
	}
	return toBytes(img)
}
