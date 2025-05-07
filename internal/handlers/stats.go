package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	v1 "github.com/kyzrfranz/mai-ling/api/v1"
	"github.com/kyzrfranz/mai-ling/internal/cache"
	"github.com/kyzrfranz/mai-ling/internal/db"
	internalHttp "github.com/kyzrfranz/mai-ling/internal/http"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var zipCache *lru.Cache

type StatsHandler struct {
	collection *mongo.Collection
	logger     *slog.Logger
	cache      *cache.Cache
}

func NewStatsHandler(collection *mongo.Collection, logger *slog.Logger, bCache *cache.Cache) *StatsHandler {
	return &StatsHandler{
		collection: collection,
		logger:     logger,
		cache:      bCache,
	}
}

func (h *StatsHandler) Handle(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		h.handleGet(w, req)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func (h *StatsHandler) handleGet(w http.ResponseWriter, req *http.Request) {
	// Handle GET request
	special := req.PathValue("special")

	switch special {
	case "by-location":
		h.handleGetByCity(w, req)
		return
	case "by-creation":
		h.handleByCreation(w, req)
		return
	default:
		h.handleStats(w, req)
		return
	}
}

func (h *StatsHandler) handleStats(w http.ResponseWriter, req *http.Request) {

	top := req.URL.Query().Get("top")
	topIds, err := strconv.Atoi(top)
	if err != nil {
		topIds = 10
	}

	pipeline := db.DefaultPipeline(topIds)

	// Set a timeout for the aggregation operation.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := h.collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal("Aggregation error: ", err)
	}
	defer cursor.Close(ctx)

	var results []v1.Stats
	if err = cursor.All(ctx, &results); err != nil {
		h.logger.Error("Failed to decode", "error", err)
		http.Error(w, "Failed to decode", http.StatusInternalServerError)
	}

	if err := internalHttp.MarshalJsonResponse(w, results); err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
}

func (h *StatsHandler) handleGetByCity(w http.ResponseWriter, req *http.Request) {
	groupBy := req.URL.Query().Get("groupBy")

	if groupBy == "" {
		groupBy = "city"
	}

	pipeline, err := db.ByLocationPipeline(groupBy)
	if err != nil {
		h.logger.Error("Failed to create pipeline", "error", err)
		http.Error(w, "Invalid groupBy parameter", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := h.collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal("Aggregation error: ", err)
	}
	defer cursor.Close(ctx)

	var results []v1.StatsByLocation
	if err = cursor.All(ctx, &results); err != nil {
		h.logger.Error("Failed to decode", "error", err)
		http.Error(w, "Failed to decode", http.StatusInternalServerError)
	}

	for i := range results {
		if results[i].ZipCode == "" {
			continue // Skip if no ZipCode, e.g., grouped by city
		}

		zip := results[i].ZipCode
		response, err := h.fromCacheOrUrl(zip, "https://api.zippopotam.us/de/")
		if err != nil {
			h.logger.Error("Failed to fetch city name", "error", err)
			http.Error(w, "Failed to fetch city name", http.StatusInternalServerError)
			return
		}
		if response != nil && len(response.Places) > 0 {
			results[i].City = response.Places[0].PlaceName
			results[i].Lat = response.Places[0].Latitude
			results[i].Lng = response.Places[0].Longitude
		} else {
			h.logger.Info("City name not found for zip code", "zip", zip)
		}

	}

	if err := internalHttp.MarshalJsonResponse(w, results); err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
}

func (h *StatsHandler) handleByCreation(w http.ResponseWriter, req *http.Request) {
	pipeline := db.ByCreationPipeline()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := h.collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal("Aggregation error: ", err)
	}
	defer cursor.Close(ctx)

	var results []v1.StatsByCreationDate
	if err = cursor.All(ctx, &results); err != nil {
		h.logger.Error("Failed to decode", "error", err)
		http.Error(w, "Failed to decode", http.StatusInternalServerError)
	}

	if err := internalHttp.MarshalJsonResponse(w, results); err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
}

func (h *StatsHandler) fromCacheOrUrl(key string, fetchUrl string) (*hpResponse, error) {
	var cached hpResponse
	hit, err := h.cache.Get(key, &cached)
	if err != nil {
		h.logger.Error("Cache error", "error", err)
		return nil, err
	}

	if hit {
		h.logger.Info("Cache hit", "key", key)
		return &cached, nil
	}

	slog.Info("Cache miss", "key", key)
	u, _ := url.Parse(fmt.Sprint(fetchUrl, key))
	body, _ := internalHttp.FetchUrl(u)

	var resp hpResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		h.logger.Error("Failed to unmarshal response", "error", err)
		_ = h.cache.Set(key, nil) // prevent repeated fetches
		return nil, nil
	}

	_ = h.cache.Set(key, &resp)
	return &resp, nil
}

type cityResponse struct {
	PostalCode   string `json:"postalCode"`
	Name         string `json:"name"`
	Municipality struct {
		Key  string `json:"key"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"municipality"`
	FederalState struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"federalState"`
}

type hpResponse struct {
	PostCode            string `json:"post code"`
	Country             string `json:"country"`
	CountryAbbreviation string `json:"country abbreviation"`
	Places              []struct {
		PlaceName         string `json:"place name"`
		Longitude         string `json:"longitude"`
		State             string `json:"state"`
		StateAbbreviation string `json:"state abbreviation"`
		Latitude          string `json:"latitude"`
	} `json:"places"`
}
