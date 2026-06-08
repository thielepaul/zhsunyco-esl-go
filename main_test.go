package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatTime(t *testing.T) {
	cases := []struct {
		date     time.Time
		expected string
	}{
		{time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC), "So 14.6."},
		{time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC), "Mo 15.6."},
		{time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC), "Di 16.6."},
		{time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC), "Mi 17.6."},
		{time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC), "Do 18.6."},
		{time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC), "Fr 19.6."},
		{time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC), "Sa 20.6."},
	}
	for _, tt := range cases {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatDayGerman(tt.date))
		})
	}
}
