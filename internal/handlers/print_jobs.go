package handlers

import (
	"context"
	v1 "github.com/kyzrfranz/mai-ling/api/v1"
	internalHttp "github.com/kyzrfranz/mai-ling/internal/http"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"net/http"
	"time"
)

type PrintJobHandler struct {
	collection *mongo.Collection
	logger     *slog.Logger
	authKey    string
}

func NewPrintJobHandler(collection *mongo.Collection, logger *slog.Logger, authKey string) *PrintJobHandler {
	return &PrintJobHandler{
		collection: collection,
		logger:     logger,
		authKey:    authKey,
	}
}

func (h *PrintJobHandler) Handle(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		h.Generate(w, req)
	case http.MethodGet:
		h.List(w, req)
	case http.MethodPatch:
		h.Patch(w, req)
	case http.MethodDelete:
		h.Delete(w, req)
	default:
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func (h *PrintJobHandler) Generate(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	job, err := internalHttp.UnmarshalJsonRequest[v1.PrintJob](req)
	if err != nil {
		h.logger.Error("Failed to marshal", "error", err)
		http.Error(w, "request object is invalid ", http.StatusBadRequest)
		return
	}

	actionParam := req.URL.Query().Get("action")
	if actionParam == "queue" {
		job.Status = v1.PrintJobStatus(v1.PrintJobStatusQueued)
		job.CreationDate = time.Now()
		job.Id = primitive.NewObjectID().Hex()
		id, err := h.collection.InsertOne(context.Background(), job)
		if err != nil {
			h.logger.Error("Failed to queue", "error", err)
			http.Error(w, "Failed to queue", http.StatusInternalServerError)
			return
		}
		h.logger.Info("Queued", "id", id, "collection", h.collection.Name())
		w.WriteHeader(http.StatusOK)

		return
	}
	http.Error(w, "Invalid action", http.StatusBadRequest)
}

func (h *PrintJobHandler) List(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if h.authKey == "" {
		h.logger.Error("Auth key is empty")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//check Authorization
	if req.Header.Get("Authorization") != "Bearer "+h.authKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var jobs []v1.PrintJob
	opts := options.Find().SetSort(bson.D{{"created_date", -1}})
	cursor, err := h.collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		h.logger.Error("Failed to list", "error", err)
		http.Error(w, "Failed to list", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		var letter v1.PrintJob
		if err := cursor.Decode(&letter); err != nil {
			h.logger.Error("Failed to decode", "error", err)
			http.Error(w, "Failed to decode", http.StatusInternalServerError)
			return
		}
		jobs = append(jobs, letter)
	}
	if err := cursor.Err(); err != nil {
		h.logger.Error("Failed to list", "error", err)
		http.Error(w, "Failed to list", http.StatusInternalServerError)
		return
	}

	if err := internalHttp.MarshalJsonResponse(w, jobs); err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}

}

func (h *PrintJobHandler) Patch(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPatch {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if req.Header.Get("Authorization") != "Bearer "+h.authKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	id := req.PathValue("id")
	if id == "" {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	patchRequest, err := internalHttp.UnmarshalJsonRequest[map[string]interface{}](req)
	if err != nil {
		h.logger.Error("Failed to decode", "error", err)
		http.Error(w, "Failed to decode", http.StatusInternalServerError)
		return
	}

	update := bson.M{"$set": patchRequest}
	res, err := h.collection.UpdateOne(context.Background(),
		bson.M{"_id": id}, update)
	if err != nil {
		objectId, err := primitive.ObjectIDFromHex(id)
		res, err = h.collection.UpdateOne(context.Background(),
			bson.M{"_id": objectId}, update)
		if err != nil {
			h.logger.Error("Failed to update", "error", err)
			http.Error(w, "Failed to update", http.StatusInternalServerError)
			return
		}
	}
	if res.MatchedCount == 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *PrintJobHandler) Delete(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if req.Header.Get("Authorization") != "Bearer "+h.authKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	id := req.PathValue("id")
	if id == "" {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	res, err := h.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if res.DeletedCount == 0 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if err != nil {
		h.logger.Error("Failed to delete", "error", err)
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
