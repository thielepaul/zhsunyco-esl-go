package image

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestGenerate(t *testing.T) {
	cases := []struct {
		name string
		data []Weather
	}{
		{"clear partly-cloudy cloudy", []Weather{
			{High: 10, Low: 20, Icon: "clear", Text: "Hello World!"},
			{High: 10, Low: 20, Icon: "partly-cloudy", Text: "Hello World!"},
			{High: 10, Low: 20, Icon: "cloudy", Text: "Hello World!"},
		}},
		{"fog wind rain", []Weather{
			{High: 10, Low: 20, Icon: "fog", Text: "Hello World!"},
			{High: 10, Low: 20, Icon: "wind", Text: "Hello World!"},
			{High: 10, Low: 20, Icon: "rain", Text: "Hello World!"},
		}},
		{"sleet snow hail", []Weather{
			{High: 10, Low: 20, Icon: "sleet", Text: "Hello World!"},
			{High: 10, Low: 20, Icon: "snow", Text: "Hello World!"},
			{High: 10, Low: 20, Icon: "hail", Text: "Hello World!"},
		}},
		{"thunderstorm null empty", []Weather{
			{High: 10, Low: 20, Icon: "thunderstorm", Text: "Hello World!"},
			{High: 10, Low: 20, Icon: "null", Text: "Hello World!"},
			{High: 10, Low: 20, Icon: "", Text: "Hello World!"},
		}},
	}
	for _, tt := range cases {
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
