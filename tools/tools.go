package main

import (
	"encoding/json"
	"os"
)

func main() {
	data, err := os.ReadFile("./data/zipcodes.de.json")
	if err != nil {
		panic(err)
	}
	var zipcodes []ZipCode
	err = json.Unmarshal(data, &zipcodes)
	println(zipcodes)

	bData, _ := json.Marshal(zipcodes)
	os.WriteFile("./data/zipcodes.de.nocoord.json", bData, 0644)
}

type ZipCode struct {
	Name        string `json:"name"`
	PlzName     string `json:"plz_name"`
	PlzNameLong string `json:"plz_name_long"`
	//Geometry    struct {
	//	Type     string `json:"type"`
	//	Geometry struct {
	//		Coordinates [][][]float64 `json:"coordinates"`
	//		Type        string        `json:"type"`
	//	} `json:"geometry"`
	//	Properties struct {
	//	} `json:"properties"`
	//} `json:"geometry"`
}
