package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

const apiBase = "https://api.brightsky.dev/weather"

type Forecast struct {
	Icon                     string
	High                     int
	Low                      int
	PrecipitationAmount      float64
	PrecipitationProbability int
}

// internal types for JSON decoding
type weatherEntry struct {
	Timestamp                  string  `json:"timestamp"`
	Temperature                float64 `json:"temperature"`
	Precipitation              float64 `json:"precipitation"`
	PrecipitationProbability6h int     `json:"precipitation_probability_6h"`
	Icon                       string  `json:"icon"`
}

type weatherResponse struct {
	Weather []weatherEntry `json:"weather"`
}

func GetForecast(lat, lon string, date time.Time) (Forecast, error) {
	targetDate := date.UTC().Format("2006-01-02")
	resp, err := http.Get(fmt.Sprintf("%s?date=%s&lat=%s&lon=%s",
		apiBase, targetDate, lat, lon))
	if err != nil {
		return Forecast{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Forecast{}, fmt.Errorf("failed to get forecast: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Forecast{}, fmt.Errorf("failed to read response: %w", err)
	}

	return parseForecast(body, targetDate)
}

func parseForecast(body []byte, targetDate string) (Forecast, error) {
	var wr weatherResponse
	if err := json.Unmarshal(body, &wr); err != nil {
		return Forecast{}, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(wr.Weather) == 0 {
		return Forecast{}, fmt.Errorf("no weather data for %s", targetDate)
	}

	high := math.Inf(-1)
	low := math.Inf(1)
	var totalPrecip float64
	maxPrecipProb := 0
	iconCounts := make(map[string]int)

	for _, entry := range wr.Weather {
		t, err := time.Parse(time.RFC3339, entry.Timestamp)
		if err != nil {
			continue
		}
		// Skip entries that don't belong to the requested date (API can return
		// a trailing midnight entry for the following day)
		if t.UTC().Format("2006-01-02") != targetDate {
			continue
		}

		if entry.Temperature > high {
			high = entry.Temperature
		}
		if entry.Temperature < low {
			low = entry.Temperature
		}
		totalPrecip += entry.Precipitation
		if entry.PrecipitationProbability6h > maxPrecipProb {
			maxPrecipProb = entry.PrecipitationProbability6h
		}
		// Count icons only during daylight hours (06:00–19:59 UTC) so that
		// night icons don't dominate the representative daily icon.
		if h := t.UTC().Hour(); h >= 6 && h < 20 {
			iconCounts[entry.Icon]++
		}
	}

	if math.IsInf(high, -1) {
		return Forecast{}, fmt.Errorf("no entries found for %s", targetDate)
	}

	// Most frequently occurring daytime icon wins.
	bestIcon, bestCount := "", 0
	for icon, count := range iconCounts {
		if count > bestCount {
			bestCount = count
			bestIcon = icon
		}
	}

	bestIcon = strings.TrimSuffix(bestIcon, "-day")
	bestIcon = strings.TrimSuffix(bestIcon, "-night")

	return Forecast{
		Icon:                     bestIcon,
		High:                     int(math.Round(high)),
		Low:                      int(math.Round(low)),
		PrecipitationAmount:      math.Round(totalPrecip*10) / 10,
		PrecipitationProbability: maxPrecipProb,
	}, nil
}
