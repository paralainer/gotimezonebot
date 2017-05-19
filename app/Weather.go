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

type GetWeather func(location Location, ch chan<- Weather)

func WrapWeatherWithCache(getWeather GetWeather) GetWeather {
	weatherCache := cache.New(time.Duration(15*time.Minute), time.Duration(1*time.Hour))

	return func(location Location, ch chan<- Weather) {
		result, ok := weatherCache.Get(location.Coordinates)
		if !ok {
			wCh := make(chan Weather)
			go getWeather(location, wCh)
			weather := <-wCh
			weatherCache.SetDefault(location.Coordinates, weather)
			ch <- weather
		} else {
			ch <- result.(Weather)
		}
	}

}

func CreateDarkSkyWeatherFetcher(apiKey string) GetWeather {
	return func(location Location, ch chan<- Weather) {
		result := makeRequest(apiKey, location.Coordinates)

		currentWeather := result["currently"].(map[string]interface{})

		temp := strconv.Itoa(int(fahrenheit2Celsius(currentWeather["temperature"].(float64))))

		emoji, ok := emojis[currentWeather["icon"].(string)]
		if !ok {
			emoji = ""
		}
		ch <- Weather{
			Conditions: " " + emoji + " " + temp + "℃",
			Timezone:   result["timezone"].(string),
			Location:   location,
		}

	}
}

func makeRequest(apiKey string, coordinates string) map[string]interface{} {
	resp, err := http.Get("https://api.darksky.net/forecast/" + apiKey + "/" + coordinates)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var f interface{}
	json.Unmarshal(body, &f)
	return f.(map[string]interface{})
}

func fahrenheit2Celsius(f float64) float64 {
	return (f - 32) / 1.8
}
