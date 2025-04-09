package handlers

import (
	"context"
	v1 "github.com/kyzrfranz/mai-ling/api/v1"
	internalHttp "github.com/kyzrfranz/mai-ling/internal/http"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

type StatsHandler struct {
	collection *mongo.Collection
	logger     *slog.Logger
}

func NewStatsHandler(collection *mongo.Collection, logger *slog.Logger) *StatsHandler {
	return &StatsHandler{
		collection: collection,
		logger:     logger,
	}
}

func (h *StatsHandler) Handle(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		h.Stats(w, req)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func (h *StatsHandler) Stats(w http.ResponseWriter, req *http.Request) {
	top := req.URL.Query().Get("top")
	topIds, err := strconv.Atoi(top)
	if err != nil {
		topIds = 10
	}

	pipeline := mongo.Pipeline{
		{{"$facet", bson.D{
			// Facet for unique summary based on address & ids
			{"uniqueSummary", bson.A{
				bson.D{{"$addFields", bson.D{
					{"idsCount", bson.D{{"$size", "$ids"}}},
				}}},
				bson.D{{"$group", bson.D{
					{"_id", bson.D{
						{"address", "$address"},
						{"ids", "$ids"},
					}},
					{"requestIdsCount", bson.D{{"$first", "$idsCount"}}},
				}}},
				bson.D{{"$group", bson.D{
					{"_id", nil},
					{"uniqueRequests", bson.D{{"$sum", 1}}},
					{"totalIds", bson.D{{"$sum", "$requestIdsCount"}}},
				}}},
				bson.D{{"$project", bson.D{
					{"_id", 0},
					{"uniqueRequests", 1},
					{"totalIds", 1},
				}}},
			}},
			{"topIds", bson.A{
				bson.D{{"$unwind", "$ids"}},
				bson.D{{"$group", bson.D{
					{"_id", "$ids"},
					{"count", bson.D{{"$sum", 1}}},
				}}},
				bson.D{{"$sort", bson.D{{"count", -1}}}},
				bson.D{{"$limit", topIds}},
			}},
			// Facet for status counts with deduplication.
			{"statusCounts", bson.A{
				// First, add the idsCount field to each document.
				bson.D{{"$addFields", bson.D{
					{"idsCount", bson.D{{"$size", "$ids"}}},
				}}},
				// Deduplicate by composite key (address and ids), capturing the first status and idsCount.
				bson.D{{"$group", bson.D{
					{"_id", bson.D{
						{"address", "$address"},
						{"ids", "$ids"},
					}},
					{"status", bson.D{{"$first", "$status"}}},
					{"idsCount", bson.D{{"$first", "$idsCount"}}},
				}}},
				// Now group all unique jobs to count statuses and accumulate sentLetters.
				bson.D{{"$group", bson.D{
					{"_id", nil},
					{"queued", bson.D{{"$sum", bson.D{{"$cond", bson.A{
						bson.D{{"$eq", bson.A{"$status", "queued"}}},
						1,
						0,
					}}}}}},
					{"sent", bson.D{{"$sum", bson.D{{"$cond", bson.A{
						bson.D{{"$eq", bson.A{"$status", "sent"}}},
						1,
						0,
					}}}}}},
					{"sentLetters", bson.D{{"$sum", bson.D{{"$cond", bson.A{
						bson.D{{"$eq", bson.A{"$status", "sent"}}},
						"$idsCount",
						0,
					}}}}}},
				}}},
				bson.D{{"$project", bson.D{
					{"_id", 0},
					{"queued", 1},
					{"sent", 1},
					{"sentLetters", 1},
				}}},
			}},
		}}},
		{{"$project", bson.D{
			{"uniqueRequests", bson.D{{"$arrayElemAt", bson.A{"$uniqueSummary.uniqueRequests", 0}}}},
			{"totalIds", bson.D{{"$arrayElemAt", bson.A{"$uniqueSummary.totalIds", 0}}}},
			{"topIds", 1},
			{"sentIds", bson.D{{"$arrayElemAt", bson.A{"$statusCounts.sentLetters", 0}}}},
			{"statusCounts", bson.D{{"$arrayElemAt", bson.A{"$statusCounts", 0}}}},
		}}},
	}

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
