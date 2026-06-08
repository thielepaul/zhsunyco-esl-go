package main

import (
	"log"
	"os"
	"time"

	"github.com/thielepaul/zhsunyco-esl-go/connect"
	"github.com/thielepaul/zhsunyco-esl-go/image"
	"github.com/thielepaul/zhsunyco-esl-go/protocol"
	"github.com/thielepaul/zhsunyco-esl-go/weather"
)

func main() {
	macStr := getEnv("MAC")
	lat := getEnv("LAT")
	lon := getEnv("LON")

	for {
		weatherDays := []image.Weather{}
		for i := range 3 {
			date := time.Now().AddDate(0, 0, i)
			forecast, err := weather.GetForecast(lat, lon, date)
			if err != nil {
				log.Fatalf("failed to get forecast: %v", err)
			}
			weatherDays = append(weatherDays, image.Weather{
				Icon: forecast.Icon, Day: formatDayGerman(date),
				High: forecast.High, Low: forecast.Low,
				PrecipitationProbability: forecast.PrecipitationProbability,
				PrecipitationAmount:      forecast.PrecipitationAmount,
			})
		}

		imgBytesBw, imgBytesRed, err := image.Generate(weatherDays...)
		if err != nil {
			log.Fatalf("failed to generate image: %v", err)
		}

		packets, err := protocol.Marshal(imgBytesBw, imgBytesRed, macStr)
		if err != nil {
			log.Fatalf("failed to marshal: %v", err)
		}

		if connect.Update(macStr, packets); err != nil {
			log.Fatalf("failed to update: %v", err)
		}

		// TODO: update at 6 in the morning each day
		time.Sleep(24 * time.Hour)
	}
}

func formatDayGerman(date time.Time) string {
	weekdays := []string{"So", "Mo", "Di", "Mi", "Do", "Fr", "Sa"}
	return weekdays[date.Weekday()] + " " + date.Format("2.1.")
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("environment variable %s is not set", key)
	}
	return value
}
