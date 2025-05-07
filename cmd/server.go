package main

import (
	internalCache "github.com/kyzrfranz/mai-ling/internal/cache"
	"github.com/kyzrfranz/mai-ling/internal/db"
	"github.com/kyzrfranz/mai-ling/internal/handlers"
	"github.com/kyzrfranz/mai-ling/internal/http"
	"log/slog"
	"os"
)

var (
	logger          *slog.Logger
	mongoUri        string
	mongoDbName     string
	mongoCollection string
	authKey         string
)

func main() {

	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mongoUri = stringOrEnv("MONGO_URI", "")

	mongoCollection = stringOrEnv("MONGO_COLLECTION", "test")
	mongoDbName = stringOrEnv("MONGO_DB_NAME", "test")
	authKey = stringOrEnv("AUTH_KEY", "")
	cli, err := db.NewV1MongoClient(db.WithUri(mongoUri))
	if err != nil {
		logger.Error("failed to connect to mongo", "error", err)
		os.Exit(1)
	}

	apiServer := http.NewApiServer(8080, logger)

	apiServer.Use(http.MiddlewareRecovery)
	apiServer.Use(http.MiddlewareCORS)

	cache, err := internalCache.NewCache("./.cache")
	if err != nil {
		bail("init cache", err)
	}

	collection := cli.Database(mongoDbName).Collection(mongoCollection)
	letterHandler := handlers.NewPrintJobHandler(collection, logger, authKey)
	statsHandler := handlers.NewStatsHandler(collection, logger, cache)
	zipcodeHandler := handlers.NewZipCodeHandler(logger)
	apiServer.AddHandler("/letters/{id}", letterHandler.Handle)
	apiServer.AddHandler("/letters", letterHandler.Handle)
	apiServer.AddHandler("/zipcodes/{zipcode}", zipcodeHandler.Handle)
	apiServer.AddHandler("/stats", statsHandler.Handle)
	apiServer.AddHandler("/stats/{special}", statsHandler.Handle)
	apiServer.AddStaticHandler("/static/", "static")

	apiServer.ListenAndServe()
}

func bail(stage string, err error) {
	logger.Error("server bailing out", slog.String("stage", stage), "error", err)
	os.Exit(1)
}

func stringOrEnv(key string, defaultVal string) (s string) {
	s = os.Getenv(key)
	if s != "" {
		defaultVal = s
	}

	return defaultVal
}
