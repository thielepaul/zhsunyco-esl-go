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

func TestGenerateVisual(t *testing.T) {
	cases := []struct {
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
