package v1

type ZipCode struct {
	ZipCode      string      `json:"zip_code"`
	CityName     string      `json:"city_name"`
	CityNameLong string      `json:"city_name_long"`
	Coordinates  Coordinates `json:"coordinates"`
}

type Coordinates struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
