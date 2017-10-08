package app

import (
	"net/http"
	"io/ioutil"
	"log"
	"encoding/json"
	"strconv"
	"github.com/patrickmn/go-cache"
	"time"
)

type Weather struct {
	Location   Location
	Conditions string
	Timezone   string
}

var emojis = map[string]string{
	"clear-day":           "☀",
	"clear-night":         "🌛",
	"rain":                "\U0001f327",
	"snow":                "❄️ ☃",
	"sleet":               "\U0001f327 \U0001f328",
	"wind":                "\U0001f390",
	"fog":                 "\U0001f32b",
	"cloudy":              "☁️",
	"partly-cloudy-day":   "🌤",
	"partly-cloudy-night": "☁️🌛",
	"hail":                "\U0001f327",
	"thunderstorm":        "⛈",
	"tornado":             "\U0001f32a",
}

type GetWeather func(location Location, ch chan<- WeatherResult)

func WrapWeatherWithCache(getWeather GetWeather) GetWeather {
	weatherCache := cache.New(time.Duration(15*time.Minute), time.Duration(1*time.Hour))

	return func(location Location, ch chan<- WeatherResult) {
		result, ok := weatherCache.Get(location.Coordinates)
		if !ok {
			wCh := make(chan WeatherResult)
			go getWeather(location, wCh)
			weather := <-wCh
			if weather.Error != nil {
				weatherCache.SetDefault(location.Coordinates, weather.Weather)
			}
			ch <- weather
		} else {
			ch <- WeatherResult{
				Weather: result.(Weather),
			}
		}
	}

}

type WeatherResult struct {
	Weather Weather
	Error   error
}

func CreateDarkSkyWeatherFetcher(apiKey string) GetWeather {
	return func(location Location, ch chan<- WeatherResult) {
		result, err := makeRequest(apiKey, location.Coordinates)

		if err != nil {
			ch <- WeatherResult{
				Error: err,
			}
		}

		currentWeather := result["currently"].(map[string]interface{})

		temp := strconv.Itoa(int(fahrenheit2Celsius(currentWeather["temperature"].(float64))))

		emoji, ok := emojis[currentWeather["icon"].(string)]
		if !ok {
			emoji = ""
		}
		ch <- WeatherResult{
			Weather: Weather{
				Conditions: " " + emoji + " " + temp + "℃",
				Timezone:   result["timezone"].(string),
				Location:   location,
			},
		}
	}
}

func makeRequest(apiKey string, coordinates string) (map[string]interface{}, error) {
	resp, err := http.Get("https://api.darksky.net/forecast/" + apiKey + "/" + coordinates)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var f interface{}
	json.Unmarshal(body, &f)
	return f.(map[string]interface{}), nil
}

func fahrenheit2Celsius(f float64) float64 {
	return (f - 32) / 1.8
}
