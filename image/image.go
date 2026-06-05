package image

import (
	"image"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func drawText(img *image.RGBA, x, y int, text string) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: basicfont.Face7x13,
		Dot:  fixed.P(x, y),
	}
	d.DrawString(text)
}

func toBytes(img *image.RGBA) []byte {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	buf := make([]byte, w*h)
	for y := range h {
		for x := range w {
			r, _, _, _ := img.At(x, y).RGBA()
			pos := y*w + x
			if r < 0x8000 { // dark pixel = set byte to 1
				buf[pos] = 1
			} else {
				buf[pos] = 0
			}
		}
	}
	return buf
}

func Generate(x, y int, text string) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 296, 128))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	drawText(img, x, y, text)
	return toBytes(img)
}

func SavePNG(path string) error {
	img := image.NewRGBA(image.Rect(0, 0, 296, 128))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	drawText(img, 10, 60, "Hello World!")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
