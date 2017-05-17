package main

import (
	"net/http"
	"log"
	"io/ioutil"
	"net/url"
	"encoding/json"
	"errors"
	"strconv"
)

const (
	GeoNotFound = "GeoNotFound"
)

type GeoInfo struct {
	LocationDisplayName string
	Lat float64
	Lon float64
}

func GetGeoInfo(locQuery string) (*GeoInfo, error) {
	result := makeGeoRequest(locQuery)
	if len(result) == 0 {
		return nil, errors.New(GeoNotFound)
	}

	placeInfo := result[0].(map[string]interface{})

	lat,_ := strconv.ParseFloat(placeInfo["lat"].(string), 32)
	lon,_ := strconv.ParseFloat(placeInfo["lon"].(string), 32)

	return &GeoInfo{
		LocationDisplayName: placeInfo["display_name"].(string),
		Lat: lat,
		Lon: lon,
	}, nil

}

func makeGeoRequest(query string) []interface{} {
	resp, err := http.Get("http://nominatim.openstreetmap.org/?format=json&addressdetails=1&q=" + url.QueryEscape(query) + "&format=json&limit=1")
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var f interface{}
	json.Unmarshal(body, &f)
	return f.([]interface{})
}
