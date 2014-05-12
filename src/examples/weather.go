package main

import (
	"encoding/json"
	"fmt"
	"weather"
)

func main() {
	w := TestStaticData()
	fmt.Println(w)
	s := weather.GetStringWeather("Guadalajara")
	fmt.Println("Weather in GDL", s)
}

func TestStaticData() weather.Weather {
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
	t := &weather.ValidAddresses{}
	err := json.Unmarshal(str, &t)
	if err != nil {
		fmt.Println(err)
	}
	w := weather.Convert(*t)
	return w
}
