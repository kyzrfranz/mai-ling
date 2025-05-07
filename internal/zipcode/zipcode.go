package zipcode

import (
	"encoding/json"
	"fmt"
	v1 "github.com/kyzrfranz/mai-ling/api/v1"
)

type ZipCode struct {
	data []zipcodeInternal
}

func NewZipCode(data []byte) (*ZipCode, error) {
	var zipcodes []zipcodeInternal
	err := json.Unmarshal(data, &zipcodes)
	if err != nil {
		return nil, err
	}
	return &ZipCode{
		data: zipcodes,
	}, nil
}

func (h *ZipCode) FindByZipCode(zipcode string) (*v1.ZipCode, error) {
	var foundZipCode *v1.ZipCode
	for _, zip := range h.data {
		if zip.Name == zipcode {
			foundZipCode = &v1.ZipCode{
				ZipCode:      zip.Name,
				CityName:     zip.PlzName,
				CityNameLong: zip.PlzNameLong,
				Coordinates: v1.Coordinates{
					Lat: zip.Geometry.Geometry.Coordinates[0][0][1],
					Lng: zip.Geometry.Geometry.Coordinates[0][0][0],
				},
			}
			break
		}
	}
	if foundZipCode == nil {
		return nil, fmt.Errorf("zipcode not found")
	}

	return foundZipCode, nil
}

type zipcodeInternal struct {
	Name        string `json:"name"`
	PlzName     string `json:"plz_name"`
	PlzNameLong string `json:"plz_name_long"`
	Geometry    struct {
		Type     string `json:"type"`
		Geometry struct {
			Coordinates [][][]float64 `json:"coordinates"`
			Type        string        `json:"type"`
		} `json:"geometry"`
		Properties struct {
		} `json:"properties"`
	} `json:"geometry"`
}
