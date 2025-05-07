package handlers

import (
	"context"
	lru "github.com/hashicorp/golang-lru"
	v1 "github.com/kyzrfranz/mai-ling/api/v1"
	"github.com/kyzrfranz/mai-ling/internal/cache"
	"github.com/kyzrfranz/mai-ling/internal/db"
	internalHttp "github.com/kyzrfranz/mai-ling/internal/http"
	"github.com/kyzrfranz/mai-ling/internal/zipcode"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

var zipCache *lru.Cache

type StatsHandler struct {
	collection *mongo.Collection
	logger     *slog.Logger
	cache      *cache.Cache
	zc         *zipcode.ZipCode
}

func NewStatsHandler(collection *mongo.Collection, logger *slog.Logger, bCache *cache.Cache) *StatsHandler {

	return &StatsHandler{
		collection: collection,
		logger:     logger,
		cache:      bCache,
		zc:         zipcode.NewZipCode(),
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
		response, err := h.zc.FindByZipCode(zip)
		if err != nil {
			h.logger.Error("Failed to fetch city name", "error", err)
		}
		if response != nil {
			results[i].City = response.CityName
			results[i].Lat = response.Coordinates.Lat
			results[i].Lng = response.Coordinates.Lng
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
