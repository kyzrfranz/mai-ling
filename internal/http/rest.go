package http

import (
	"encoding/json"
	"net/http"
)

func MarshalJsonResponse(w http.ResponseWriter, res interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(res)
	if err != nil {
		return err
	}
	_, writeErr := w.Write(jsonData) // Write the JSON data
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func UnmarshalJsonRequest[T any](r *http.Request) (T, error) {
	var request T
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return request, err
	}
	return request, nil
}
