package main

import (
	"log"
	"os"
	"time"
	_ "time/tzdata"

	"github.com/thielepaul/zhsunyco-esl-go/connect"
	"github.com/thielepaul/zhsunyco-esl-go/image"
	"github.com/thielepaul/zhsunyco-esl-go/protocol"
	"github.com/thielepaul/zhsunyco-esl-go/weather"
)

func main() {
	macStr := getEnv("MAC")
	lat := getEnv("LAT")
	lon := getEnv("LON")

	device, err := connect.NewESLDevice(macStr)
	if err != nil {
		log.Fatalf("failed to create ESL device: %v", err)
	}

	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		log.Fatalf("Failed to load Berlin timezone: %v", err)
	}

	for {
		now := time.Now().In(loc)
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, loc)
		if now.After(nextRun) {
			nextRun = nextRun.AddDate(0, 0, 1)
		}
		durationUntilNextRun := time.Until(nextRun)
		log.Printf("Next run scheduled for: %v (Sleeping for %v)\n", nextRun.Format(time.RFC1123), durationUntilNextRun)
		time.Sleep(durationUntilNextRun)

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

		for i := range 10 {
			if err := device.Update(packets); err != nil {
				log.Printf("failed to update (%d/%d): %v", i+1, 10, err)
				time.Sleep(time.Minute)
				continue
			}
			break
		}
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
