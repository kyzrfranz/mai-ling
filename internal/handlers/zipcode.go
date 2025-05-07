package handlers

import (
	internalhttp "github.com/kyzrfranz/mai-ling/internal/http"
	"github.com/kyzrfranz/mai-ling/internal/zipcode"
	"log/slog"
	"net/http"
)

type ZipCodeHandler struct {
	logger *slog.Logger
	zc     *zipcode.ZipCode
}

func NewZipCodeHandler(logger *slog.Logger) *ZipCodeHandler {
	return &ZipCodeHandler{
		zc:     zipcode.NewZipCode(),
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

	foundZipCode, err := h.zc.FindByZipCode(zipcode)
	if err != nil {
		h.logger.Error("Failed to find zipcode", "error", err)
		http.Error(w, "zipcode not found", http.StatusNotFound)
		return
	}

	internalhttp.MarshalJsonResponse(w, foundZipCode)
}
