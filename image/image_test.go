package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
)

func toImg(bwBytes []byte, redBytes []byte) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 296, 128))
	for y := range img.Bounds().Dy() {
		for x := range img.Bounds().Dx() {
			pos := y*img.Bounds().Dx() + x
			if bwBytes[pos] == 1 {
				img.Set(x, y, image.Black)
			} else if redBytes[pos] == 1 {
				img.Set(x, y, colorRed)
			} else {
				img.Set(x, y, image.White)
			}
		}
	}
	return img
}

// goldenCases doubles as the visual regression corpus and as the sample input
// for the font tests below, so the two can never drift apart.
var goldenCases = []struct {
	name string
	data []Weather
}{
	{"clear partly-cloudy cloudy", []Weather{
		{High: -2, Low: -8, PrecipitationProbability: 0, PrecipitationAmount: 0.0, Icon: "clear", Day: "Mo 31.5."},
		{High: 3, Low: -4, PrecipitationProbability: 20, PrecipitationAmount: 0.5, Icon: "partly-cloudy", Day: "Di 1.6."},
		{High: 1, Low: -6, PrecipitationProbability: 80, PrecipitationAmount: 4.2, Icon: "cloudy", Day: "Mi 22.12."},
	}},
	{"fog wind rain", []Weather{
		{High: 12, Low: 1, PrecipitationProbability: 100, PrecipitationAmount: 5.5, Icon: "fog", Day: "Do 1.1."},
		{High: 18, Low: 9, PrecipitationProbability: 40, PrecipitationAmount: 0.0, Icon: "wind", Day: "Fr 14.7."},
		{High: 15, Low: 7, PrecipitationProbability: 90, PrecipitationAmount: 8.5, Icon: "rain", Day: "Sa 12.8."},
	}},
	{"sleet snow hail", []Weather{
		{High: 0, Low: -5, PrecipitationProbability: 70, PrecipitationAmount: 3.1, Icon: "sleet", Day: "So 13.8."},
		{High: -1, Low: -12, PrecipitationProbability: 60, PrecipitationAmount: 6.4, Icon: "snow", Day: "Mo 22.12."},
		{High: 5, Low: -1, PrecipitationProbability: 30, PrecipitationAmount: 1.2, Icon: "hail", Day: "Di 10.2."},
	}},
	{"thunderstorm null empty", []Weather{
		{High: 33, Low: 22, PrecipitationProbability: 95, PrecipitationAmount: 15.5, Icon: "thunderstorm", Day: "Mi 19.11."},
		{High: 8, Low: 2, PrecipitationProbability: 0, PrecipitationAmount: 0.0, Icon: "null", Day: "Do 3.3."},
		{High: 17, Low: 6, PrecipitationProbability: 10, PrecipitationAmount: 0.1, Icon: "", Day: "Fr 10.10."},
	}},
}

func TestGenerateVisual(t *testing.T) {
	for _, tt := range goldenCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.data == nil {
				t.Errorf("data is nil")
			}
			bwBytes, redBytes, err := Generate(tt.data...)
			if err != nil {
				t.Errorf("failed to generate image: %v", err)
			}
			img := toImg(bwBytes, redBytes)
			path := filepath.Join("image_test", tt.name+".png")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Errorf("failed to create directory: %v", err)
				}
				if err := savePNG(path, img); err != nil {
					t.Errorf("failed to save PNG: %v", err)
				}
			}
			assertPNGEqual(t, path, img)
		})
	}
}

func savePNG(path string, img *image.RGBA) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func assertPNGEqual(t *testing.T, path string, img *image.RGBA) {
	f, err := os.Open(path)
	assert.NoError(t, err)
	defer f.Close()
	expected, err := png.Decode(f)
	assert.NoError(t, err)
	assert.Equal(t, expected, img)
}

// titleFontStrings returns every string the layout can draw with the title
// font. It mirrors formatDayGerman in the main package, which builds a German
// weekday abbreviation plus a "2.1." date - no leading zeros, so days and
// months are 1..31 and 1..12.
func titleFontStrings() []string {
	weekdays := []string{"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"}
	days := []int{1, 2, 9, 10, 11, 19, 20, 21, 28, 29, 30, 31}
	months := []int{1, 2, 9, 10, 11, 12}
	var out []string
	for _, wd := range weekdays {
		for _, d := range days {
			for _, m := range months {
				out = append(out, fmt.Sprintf("%s %d.%d.", wd, d, m))
			}
		}
	}
	return out
}

// infoFontStrings returns every string the layout can draw with the info font.
// Rather than trust a hand-maintained list, the ranges are swept: any input
// that can reach Generate ends up here.
func infoFontStrings() []string {
	var out []string
	for _, tt := range goldenCases {
		for _, w := range tt.data {
			out = append(out, fmt.Sprintf("%d°C", w.Low), fmt.Sprintf("%d°C", w.High))
			out = append(out, fmt.Sprintf("%d%%", w.PrecipitationProbability), fmtRain(w.PrecipitationAmount))
		}
	}
	for temp := -99; temp <= 99; temp++ {
		out = append(out, fmt.Sprintf("%d°C", temp))
	}
	for prob := 0; prob <= 100; prob++ {
		out = append(out, fmt.Sprintf("%d%%", prob))
	}
	// fmtRain switches format at 10, and %.1f of 9.999 rounds up to 10.0, so
	// the narrow band just below the threshold is the interesting one.
	for _, amount := range []float64{
		-1.5, 0, 0.04, 0.1, 0.5, 4.2, 5.5, 8.5, 9.9, 9.94, 9.99, 9.999,
		10, 10.5, 15.5, 99, 100, 999,
	} {
		out = append(out, fmtRain(amount))
	}
	return out
}

// TestFontGlyphCoverage guards both fonts against a missing glyph. font.Drawer
// substitutes .notdef - a visible box - for anything the cmap does not map, so
// an over-aggressive subset would quietly turn into boxes on the panel.
func TestFontGlyphCoverage(t *testing.T) {
	// The title font is deliberately kept at full ASCII plus the degree sign so
	// the layout can be reworded without re-subsetting. Assert that intent.
	title := newCMAPChecker(t, titleTTF)
	for r := rune(0x20); r <= 0x7e; r++ {
		assert.True(t, title.has(r), "title font is missing %q (U+%04X); it is meant to keep full ASCII", r, r)
	}
	assert.True(t, title.has('°'), "title font is missing the degree sign")

	// The info font is a purpose-built subset, so check it against the strings
	// the layout can actually produce.
	info := newCMAPChecker(t, infoTTF)
	for _, s := range infoFontStrings() {
		for _, r := range s {
			assert.True(t, info.has(r), "info font is missing %q, needed to draw %q", r, s)
		}
	}
}

// TestFontsRenderWithoutAntiAliasing is the reason both fonts are pixel fonts.
// toBytes thresholds at 1 bit, so a single intermediate gray pixel becomes a
// ragged edge on the panel. Every string the layout draws must rasterize to
// pure black, pure red or white.
func TestFontsRenderWithoutAntiAliasing(t *testing.T) {
	faces := []struct {
		name    string
		face    font.Face
		strings []string
	}{
		{"title", titleFont, titleFontStrings()},
		{"info", infoFont, infoFontStrings()},
	}
	allowed := map[color.RGBA]bool{
		{255, 255, 255, 255}: true, // white
		{0, 0, 0, 255}:       true, // black text
		{255, 0, 0, 255}:     true, // red text (isHot)
	}
	for _, f := range faces {
		for _, s := range f.strings {
			for _, red := range []bool{false, true} {
				// Size the canvas to the string instead of using a fixed one:
				// this test is the suite's hot spot, and a canvas that is merely
				// big enough scans several times fewer pixels. drawText centres
				// on x, so x = width/2 + 2 puts the pen at 2 and leaves room for
				// the 1 px side bearing without clipping either edge.
				width := font.MeasureString(f.face, s).Round()
				const height = 48 // fits 15 px of ink at either font's size
				img := image.NewRGBA(image.Rect(0, 0, width+4, height))
				draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
				drawText(img, width/2+2, height/2, s, f.face, red)

				// Only the bounding box of the non-white pixels can hold a
				// blend, and pure white is allowed, so scanning that box is
				// equivalent to scanning the canvas.
				minX, minY, maxX, maxY, inked := width, height, 0, 0, false
				for y := range height {
					for x := range width + 4 {
						if img.RGBAAt(x, y) != (color.RGBA{255, 255, 255, 255}) {
							inked = true
							minX, maxX = min(minX, x), max(maxX, x)
							minY, maxY = min(minY, y), max(maxY, y)
						}
					}
				}
				if !inked {
					continue // nothing drawn, so nothing to be blurry
				}
				for y := minY; y <= maxY; y++ {
					for x := minX; x <= maxX; x++ {
						if c := img.RGBAAt(x, y); !allowed[c] {
							t.Fatalf("%s font drew %q with anti-aliased pixel %v at (%d,%d)",
								f.name, s, c, x, y)
						}
					}
				}
			}
		}
	}
}

// cmapChecker answers whether a font maps a rune to a real glyph. sfnt returns
// glyph index 0 - the .notdef box - for anything the cmap does not cover.
type cmapChecker struct {
	font *sfnt.Font
	buf  sfnt.Buffer
}

func newCMAPChecker(t *testing.T, ttf []byte) *cmapChecker {
	t.Helper()
	f, err := sfnt.Parse(ttf)
	require.NoError(t, err)
	return &cmapChecker{font: f}
}

func (c *cmapChecker) has(r rune) bool {
	index, err := c.font.GlyphIndex(&c.buf, r)
	return err == nil && index != 0
}
