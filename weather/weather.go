package weather

import (
	"fmt"
	"net/http"
	"time"
)

const apiBase = "https://api.brightsky.dev/weather"

type Forecast struct {
	Icon                     string
	Day                      string
	High                     int
	Low                      int
	PrecipitationAmount      float64
	PrecipitationProbability int
}

// ❯ curl -s 'https://api.brightsky.dev/weather?date=2026-06-07&lat=48.171&lon=11.564' --header 'Accept: application/json' | jq
// {
//   "weather": [
//     {
//       "timestamp": "2026-06-07T00:00:00+00:00",
//       "source_id": 46586,
//       "precipitation": 0.0,
//       "pressure_msl": 1019.1,
//       "sunshine": 0.0,
//       "temperature": 18.1,
//       "wind_direction": 240,
//       "wind_speed": 7.9,
//       "cloud_cover": 100,
//       "dew_point": 10.8,
//       "relative_humidity": 62,
//       "visibility": 66140,
//       "wind_gust_direction": 230,
//       "wind_gust_speed": 16.2,
//       "condition": "dry",
//       "precipitation_probability": null,
//       "precipitation_probability_6h": null,
//       "solar": 0.0,
//       "fallback_source_ids": {
//         "solar": 46598
//       },
//       "icon": "cloudy"
//     },

func GetForecast(lat, lon string, date time.Time) (Forecast, error) {
	data, err := http.Get(fmt.Sprintf("%s?date=%s&lat=%s&lon=%s", apiBase, date.Format("2006-01-02"), lat, lon))
	if err != nil {
		return Forecast{}, err
	}
	defer data.Body.Close()
	if data.StatusCode != 200 {
		return Forecast{}, fmt.Errorf("failed to get forecast: %d", data.StatusCode)
	}
	return Forecast{
		Icon: "clear",
	}, nil
}
