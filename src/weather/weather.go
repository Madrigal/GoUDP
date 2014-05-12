package weather

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Weather struct {
	Data `json:"weather"`
}

type Data struct {
	Name     string  `json:"name"`
	Temp     float32 `json:"temp"`
	Humidity int     `json:"humidity"`
	Pressure float32 `json:"pressure"`
	TempMin  float32 `json:"temp_min"`
	TempMax  float32 `json:"temp_max"`
}

type ValidAddresses struct {
	Coord struct {
		Lon float32 `json:"lon"`
		Lat float32 `json:"lat"`
	} `json:"coord"`
	Sys struct {
		Message float32 `json:"message"`
		Country string  `json:"country"`
		Sunrise int     `json:"sunrise"`
		Sunset  int     `json:"sunset"`
	} `json:"sys"`
	Weather []struct {
		Id          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Base string `json:"base"`
	Main struct {
		Temp     float32 `json:"temp"`
		Humidity int     `json:"humidity"`
		Pressure float32 `json:"pressure"`
		TempMin  float32 `json:"temp_min"`
		TempMax  float32 `json:"temp_max"`
	} `json:"main"`
	Wind struct {
		Speed float32 `json:"speed"`
		Gust  float32 `json:"gust"`
		Deg   float32 `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt   int    `json:"dt"`
	Id   int    `json:"id"`
	Name string `json:"name"`
	Cod  int    `json:"cod"`
}

// Convert transform OpenWeather info to relevant info
func Convert(t ValidAddresses) Weather {
	w := Weather{
		Data{
			Name:     t.Name,
			Temp:     t.Main.Temp,
			Humidity: t.Main.Humidity,
			Pressure: t.Main.Pressure,
			TempMin:  t.Main.TempMin,
			TempMax:  t.Main.TempMax,
		},
	}
	return w
}

func GetStringWeather(city string) string {
	w := GetWeather(city)
	return fmt.Sprintf("Temp in %s %f Min: %f Max: %f",
		w.Name, w.Temp, w.TempMin, w.TempMax)
}

func GetWeather(city string) Weather {
	t, err := makeRequest(city)
	if err != nil {
		return Weather{}
	}
	w := Convert(*t)
	return w
}

func makeRequest(city string) (*ValidAddresses, error) {
	request := "http://api.openweathermap.org/data/2.5/weather?q="
	request = request + city
	request = request + "&units=metric"
	resp, err := http.Get(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	t := &ValidAddresses{}
	err = json.Unmarshal(body, &t)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return t, err
}
