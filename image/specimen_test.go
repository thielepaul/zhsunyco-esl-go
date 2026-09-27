package image

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

// The font specimen: a two-line sheet showing every character the app can draw,
// once in the title font and once in the info font. It is a test artifact rather
// than panel output, so it lives beside the other golden images.
//
// Vertical layout is 3 + 15 + 3 + 15 + 3 = 39 px. Both fonts put digits and
// capitals exactly 15 px above the baseline and nothing below it, so 15 px is
// each line's real ink height, and the three 3 px bands are the top padding,
// the gap between the lines and the bottom padding.
const (
	specimenPadding = 3
	specimenLineInk = 15

	// Half-open ink bands, each ending on that line's baseline.
	specimenTitleBandY0 = specimenPadding                       // 3
	specimenTitleBandY1 = specimenPadding + specimenLineInk     // 18
	specimenInfoBandY0  = 2*specimenPadding + specimenLineInk   // 21
	specimenInfoBandY1  = 2*specimenPadding + 2*specimenLineInk // 36

	specimenHeight = 2*specimenPadding + 2*specimenLineInk + specimenPadding // 39

	// The first glyph of both fonts has a 1 px left side bearing, so a pen at
	// x=2 puts the first inked column on the 3 px padding.
	specimenOriginX = specimenPadding - 1
)

// The characters the app can draw in each font.
//
// The shared block comes first and is laid out on a single grid, so these 12
// characters are drawn from the same x in both lines: the digits, the period of
// the "2.1." date, and the space that separates the weekday from the date.
const specimenShared = "0123456789. "

// Title only: the German weekday abbreviations from formatDayGerman.
const specimenTitleTail = "DFMSaior"

// Info only: "C" and "m" out of "%d°C" and "%.1fmm", then the other symbols.
const specimenInfoTail = "Cm%-°"

const (
	specimenTitleLine = specimenShared + specimenTitleTail
	specimenInfoLine  = specimenShared + specimenInfoTail
)

// specimenGrid is one row of cells: where each character's pen goes and how
// wide its cell is.
type specimenGrid struct {
	pen   []int
	width []int
}

// specimenSharedGrid is the shared prefix's grid. It uses the wider of the two
// fonts' advances per column, so the shared characters are drawn from the same x
// in both lines and neither font's glyph can spill sideways into the next
// column. Building it once, here, is what makes that agreement structural: both
// layouts consume this exact grid rather than each deriving its own and relying
// on a test to notice they had diverged.
//
// As it happens the max is redundant today: the two fonts advance every digit
// identically, and the title period and space are the wider of the pair, so
// max(title, info) equals the title advance for all twelve shared characters.
// It stays because the digits are the part most likely to be redrawn, and a
// wider info digit would otherwise spill into the next column. TestFontSpecimenSharedGrid
// is what keeps that guarantee honest.
//
// It is a OnceValue rather than a plain var because the fonts are themselves
// package-level vars filled in by init, and var initializers run before init.
var specimenSharedGrid = sync.OnceValue(func() specimenGrid {
	var g specimenGrid
	x := specimenOriginX
	for _, r := range specimenShared {
		w := max(specimenAdvance(titleFont, r), specimenAdvance(infoFont, r))
		g.pen, g.width = append(g.pen, x), append(g.width, w)
		x += w
	}
	return g
})

// specimenLayout is one line of the sheet: where each character's pen goes, how
// wide its cell is, and the band of rows it lives in.
type specimenLayout struct {
	name   string
	line   string
	face   font.Face
	grid   specimenGrid
	shared int // how many leading cells the two lines have in common
	band   [2]int
}

func (l specimenLayout) endX() int {
	return l.grid.pen[len(l.grid.pen)-1] + l.grid.width[len(l.grid.width)-1]
}

func specimenAdvance(face font.Face, r rune) int {
	return font.MeasureString(face, string(r)).Round()
}

func titleSpecimenLayout() specimenLayout {
	return specimenLayoutOf("title", specimenTitleTail, titleFont,
		[2]int{specimenTitleBandY0, specimenTitleBandY1})
}

func infoSpecimenLayout() specimenLayout {
	return specimenLayoutOf("info", specimenInfoTail, infoFont,
		[2]int{specimenInfoBandY0, specimenInfoBandY1})
}

// specimenLayoutOf places one line: the shared grid first, then the line's own
// tail flowing on its own font's advances, which lets the two tails differ in
// both content and width.
func specimenLayoutOf(name, tail string, face font.Face, band [2]int) specimenLayout {
	shared := specimenSharedGrid()
	l := specimenLayout{
		name:   name,
		line:   specimenShared + tail,
		face:   face,
		shared: len([]rune(specimenShared)),
		band:   band,
		grid: specimenGrid{
			pen:   append([]int(nil), shared.pen...),
			width: append([]int(nil), shared.width...),
		},
	}
	x := l.endX()
	for _, r := range tail {
		w := specimenAdvance(face, r)
		l.grid.pen, l.grid.width = append(l.grid.pen, x), append(l.grid.width, w)
		x += w
	}
	return l
}

// fontSpecimen renders the sheet.
func fontSpecimen() *image.RGBA {
	title, info := titleSpecimenLayout(), infoSpecimenLayout()
	width := max(title.endX(), info.endX()) + specimenPadding

	img := image.NewRGBA(image.Rect(0, 0, width, specimenHeight))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	for _, l := range []specimenLayout{title, info} {
		// Range over runes, not the string: l.grid is indexed by rune, and the
		// byte offsets of a string do not line up with them once a multi-byte
		// glyph such as the degree sign is in play.
		for i, r := range []rune(l.line) {
			(&font.Drawer{
				Dst: img, Src: image.Black, Face: l.face,
				Dot: fixed.P(l.grid.pen[i], l.band[1]),
			}).DrawString(string(r))
		}
	}
	return img
}

// TestFontSpecimen is the golden image for the sheet. Like TestGenerateVisual it
// writes the file when it is missing, so deleting it re-blesses the sheet.
func TestFontSpecimen(t *testing.T) {
	img := fontSpecimen()
	path := filepath.Join("image_test", "font-specimen.png")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
		require.NoError(t, savePNG(path, img))
	}
	assertPNGEqual(t, path, img)
}

// TestFontSpecimenGeometry pins the sheet to 3 + 15 + 3 + 15 + 3 = 39 px and
// proves the ink stays inside the two 15 px bands, so the gap and the padding
// really are empty and neither line is clipped.
func TestFontSpecimenGeometry(t *testing.T) {
	img := fontSpecimen()
	b := img.Bounds()
	assert.Equal(t, 39, b.Dy(), "the sheet must be 39 px tall")
	assert.Equal(t, specimenPadding, specimenInfoBandY0-specimenTitleBandY1,
		"the gap between the lines must be 3 px")
	assert.Equal(t, specimenPadding, b.Dy()-specimenInfoBandY1,
		"the bottom padding must be 3 px")

	inked := func(x, y int) bool {
		r, _, _, _ := img.At(x, y).RGBA()
		return r < 0x8000
	}

	// The left padding is 3 px of ink-free columns. The pen starts one px
	// earlier, at x=2, because the first glyph carries a 1 px side bearing.
	firstInked := -1
	for x := range b.Dx() {
		for y := range b.Dy() {
			if inked(x, y) {
				firstInked = x
				break
			}
		}
		if firstInked >= 0 {
			break
		}
	}
	assert.Equal(t, specimenPadding, firstInked, "the first inked column must be 3 px in")
	for x := range specimenPadding {
		for y := range b.Dy() {
			assert.False(t, inked(x, y), "ink in the left padding at (%d,%d)", x, y)
		}
	}
	inBand := func(y int) bool {
		return (y >= specimenTitleBandY0 && y < specimenTitleBandY1) ||
			(y >= specimenInfoBandY0 && y < specimenInfoBandY1)
	}
	for _, band := range [][2]int{
		{specimenTitleBandY0, specimenTitleBandY1},
		{specimenInfoBandY0, specimenInfoBandY1},
	} {
		any := false
		for y := band[0]; y < band[1]; y++ {
			for x := range b.Dx() {
				any = any || inked(x, y)
			}
		}
		assert.True(t, any, "band y=%d..%d is empty", band[0], band[1]-1)
	}
	for y := range b.Dy() {
		if inBand(y) {
			continue
		}
		for x := range b.Dx() {
			assert.False(t, inked(x, y), "ink at (%d,%d), outside both 15 px bands", x, y)
		}
	}
}

// TestFontSpecimenCells pins the sheet down glyph by glyph: every cell must be
// pixel-identical to that character drawn on its own, in that font, from the same
// dot. The blank space cell is what proves no glyph spills sideways.
func TestFontSpecimenCells(t *testing.T) {
	img := fontSpecimen()
	for _, l := range []specimenLayout{titleSpecimenLayout(), infoSpecimenLayout()} {
		for i, r := range []rune(l.line) {
			// Same canvas size and same absolute dot as the sheet, so the only
			// thing that differs is that this one character was drawn alone.
			want := image.NewRGBA(img.Bounds())
			draw.Draw(want, want.Bounds(), image.White, image.Point{}, draw.Src)
			(&font.Drawer{
				Dst: want, Src: image.Black, Face: l.face,
				Dot: fixed.P(l.grid.pen[i], l.band[1]),
			}).DrawString(string(r))

			for y := l.band[0]; y < l.band[1]; y++ {
				for x := l.grid.pen[i]; x < l.grid.pen[i]+l.grid.width[i]; x++ {
					if img.RGBAAt(x, y) != want.RGBAAt(x, y) {
						t.Fatalf("%s line cell %d (%q) spans x=%d..%d, but pixel (%d,%d) is %v, want %v",
							l.name, i, r, l.grid.pen[i], l.grid.pen[i]+l.grid.width[i]-1, x, y,
							img.RGBAAt(x, y), want.RGBAAt(x, y))
					}
				}
			}
		}
	}
}

// TestFontSpecimenSharedGrid pins down the rule the shared prefix's grid has to
// obey, recomputing it here from the two faces rather than reading back the
// value the sheet was built from. Each cell must be as wide as the wider of the
// two fonts' advances - any narrower and the wider font's glyph would spill
// into the next column - and the pens must form a running sum from the origin.
func TestFontSpecimenSharedGrid(t *testing.T) {
	title, info := titleSpecimenLayout(), infoSpecimenLayout()
	assert.Equal(t, title.shared, info.shared, "both lines share the same prefix length")
	assert.Equal(t, len([]rune(specimenShared)), title.shared,
		"the shared block is the whole prefix, not part of it")

	x := specimenOriginX
	for i, r := range []rune(specimenShared) {
		want := max(specimenAdvance(titleFont, r), specimenAdvance(infoFont, r))
		for _, l := range []specimenLayout{title, info} {
			assert.Equal(t, x, l.grid.pen[i], "%s cell %d (%q) pen", l.name, i, r)
			assert.Equal(t, want, l.grid.width[i],
				"%s cell %d (%q) must be as wide as the wider of the two advances", l.name, i, r)
		}
		x += want
	}
}

// TestFontSpecimenShowsExactlyTheAppsCharacters ties the sheet to the layout: the
// two lines between them must hold every character Generate and formatDayGerman
// can produce, and nothing else.
func TestFontSpecimenShowsExactlyTheAppsCharacters(t *testing.T) {
	title := specimenTitleRepertoire()
	info := specimenInfoRepertoire()
	// The info font also carries a space glyph (see the note on infoTTF) even
	// though no current format string emits one, so the sheet shows it too.
	info[' '] = true

	assert.Equal(t, title, specimenRepertoire(specimenTitleLine),
		"the title line must show exactly the characters the title font is given")
	assert.Equal(t, info, specimenRepertoire(specimenInfoLine),
		"the info line must show exactly the characters the info font is given")
}

// specimenTitleRepertoire mirrors formatDayGerman: a weekday abbreviation, a
// space, then a "2.1." date with no leading zeros, so days run 1..31 and months
// 1..12. The full ranges are swept because titleFontStrings only samples a few
// of them, which is enough to prove glyph coverage but not to prove that a
// character is unreachable.
func specimenTitleRepertoire() map[rune]bool {
	weekdays := []string{"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"}
	out := map[rune]bool{}
	for _, wd := range weekdays {
		for day := 1; day <= 31; day++ {
			for month := 1; month <= 12; month++ {
				for _, r := range fmt.Sprintf("%s %d.%d.", wd, day, month) {
					out[r] = true
				}
			}
		}
	}
	return out
}

// specimenInfoRepertoire reuses infoFontStrings, which already sweeps the full
// input range of the four format strings Generate draws with.
func specimenInfoRepertoire() map[rune]bool {
	return specimenRepertoire(infoFontStrings()...)
}

func specimenRepertoire(strs ...string) map[rune]bool {
	out := map[rune]bool{}
	for _, s := range strs {
		for _, r := range s {
			out[r] = true
		}
	}
	return out
}
