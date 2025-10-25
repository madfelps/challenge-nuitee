package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	liteapi "github.com/liteapi-travel/go-sdk/v3"
	"github.com/madfelps/challenge-nuitee/internal/jsonlog"

	_ "github.com/lib/pq"
)

const version = "1.0.0"
const LITE_API_URL = "https://api.liteapi.travel/v3.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}

	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}

	apiKey string
}

type application struct {
	config    config
	logger    *jsonlog.Logger
	apiClient *liteapi.APIClient
	db        *sql.DB

	wg sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "API server port")

	cfg.db.dsn = os.Getenv("DATABASE_DSN")
	if cfg.db.dsn == "" {
		log.Fatal("Environment variable DATABASE_DSN is not set")
	}

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Parse()

	configuration := liteapi.NewConfiguration()

	apiKey := os.Getenv("LITE_API_KEY")

	if apiKey == "" {
		log.Fatal("Environment variable LITE_API_KEY is not set")
	}

	cfg.apiKey = apiKey

	configuration.AddDefaultHeader("X-API-KEY", apiKey)
	apiClient := liteapi.NewAPIClient(configuration)

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	logger.PrintInfo("database connection pool established with success", nil)

	app := &application{
		config:    cfg,
		logger:    logger,
		apiClient: apiClient,
		db:        db,
	}

	app.wg.Add(1)
	go func() {
		defer app.wg.Done()
		app.StartPriceMonitor()
	}()

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	logger.PrintFatal(err, nil)

}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}
