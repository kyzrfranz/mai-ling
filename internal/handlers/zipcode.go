package handlers

import (
	"encoding/json"
	v1 "github.com/kyzrfranz/mai-ling/api/v1"
	internalhttp "github.com/kyzrfranz/mai-ling/internal/http"
	"log/slog"
	"net/http"
	"os"
)

type ZipCodeHandler struct {
	logger *slog.Logger
	data   []zipcodeInternal
}

func NewZipCodeHandler(logger *slog.Logger) *ZipCodeHandler {
	data, err := os.ReadFile("./data/zipcodes.de.json")
	if err != nil {
		panic(err)
	}
	var zipcodes []zipcodeInternal
	err = json.Unmarshal(data, &zipcodes)

	return &ZipCodeHandler{
		data:   zipcodes,
		logger: logger,
	}
}

func (h *ZipCodeHandler) Handle(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		h.Get(w, req)
	case http.MethodDelete:
		h.Get(w, req)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func (h *ZipCodeHandler) Get(w http.ResponseWriter, req *http.Request) {
	zipcode := req.PathValue("zipcode")
	if zipcode == "" {
		http.Error(w, "Invalid zipcode", http.StatusBadRequest)
		return
	}
	var foundZipCode *v1.ZipCode
	for _, zip := range h.data {
		if zip.Name == zipcode {
			foundZipCode = &v1.ZipCode{
				ZipCode:      zip.Name,
				CityName:     zip.PlzName,
				CityNameLong: zip.PlzNameLong,
			}
			break
		}
	}
	if foundZipCode == nil {
		http.Error(w, "Zipcode not found", http.StatusNotFound)
		return
	}
	internalhttp.MarshalJsonResponse(w, foundZipCode)
}

type zipcodeInternal struct {
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
