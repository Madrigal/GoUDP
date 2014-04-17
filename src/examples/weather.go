package main

import "fmt"
import "encoding/json"

type Weather struct {
	Data `json:"weather"`
}

type Data struct {
	Name     string  `json:"name"`
	Temp     float32 `json:"temp"`
	Humidity int     `json:"humidity"`
	Pressure int     `json:"pressure"`
	TempMin  float32 `json:"temp_min"`
	TempMax  int     `json:"temp_max"`
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
		Pressure int     `json:"pressure"`
		TempMin  float32 `json:"temp_min"`
		TempMax  int     `json:"temp_max"`
	} `json:"main"`
	Wind struct {
		Speed float32 `json:"speed"`
		Gust  float32 `json:"gust"`
		Deg   int     `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Dt   int    `json:"dt"`
	Id   int    `json:"id"`
	Name string `json:"name"`
	Cod  int    `json:"cod"`
}

func main() {

	str := []byte(`{
   "coord":{
      "lon":-0.13,
      "lat":51.51
   },
   "sys":{
      "message":0.0719,
      "country":"GB",
      "sunrise":1397710747,
      "sunset":1397761249
   },
   "weather":[
      {
         "id":721,
         "main":"Haze",
         "description":"haze",
         "icon":"50n"
      }
   ],
   "base":"cmc stations",
   "main":{
      "temp":6.24,
      "humidity":48,
      "pressure":1021,
      "temp_min":2.78,
      "temp_max":10
   },
   "wind":{
      "speed":1.54,
      "gust":4.11,
      "deg":193
   },
   "clouds":{
      "all":24
   },
   "dt":1397702359,
   "id":2643743,
   "name":"London",
   "cod":200
}`)
	t := &ValidAddresses{}
	err := json.Unmarshal(str, &t)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(t.Main)
	fmt.Println(t.Name)

	fmt.Println(Convert(*t))
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
