package main

import (
	"context"
	"github.com/DABronskikh/ago-5/cmd/service/app"
	"github.com/DABronskikh/ago-5/pkg/business"
	"github.com/DABronskikh/ago-5/pkg/security"
	"github.com/go-chi/chi"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"net/http"
	"os"
)

const (
	defaultPort = "8080"
	defaultHost = "0.0.0.0"

	PGdefaultDSN = "postgres://app:pass@localhost:5432/" + defaultDB
	MGdefaultDSN = "mongodb://app:pass@localhost:27017/" + defaultDB
	defaultCacheDSN = "redis://localhost:6379/0"
	defaultDB    = "db"
)

func main() {
	port, ok := os.LookupEnv("APP_PORT")
	if !ok {
		port = defaultPort
	}

	host, ok := os.LookupEnv("APP_HOST")
	if !ok {
		host = defaultHost
	}

	log.Println(host)
	log.Println(port)

	PGdsn, ok := os.LookupEnv("APP_DSN_PG")
	if !ok {
		PGdsn = PGdefaultDSN
	}

	MGdsn, ok := os.LookupEnv("APP_DSN_MG")
	if !ok {
		MGdsn = MGdefaultDSN
	}

	db, ok := os.LookupEnv("APP_DB")
	if !ok {
		db = defaultDB
	}

	cacheDSN, ok := os.LookupEnv("APP_CACHE_DSN")
	if !ok {
		cacheDSN = defaultCacheDSN
	}

	if err := execute(net.JoinHostPort(host, port), PGdsn, MGdsn, db, cacheDSN); err != nil {
		os.Exit(1)
	}
}

func execute(addr string, dsnPG string, dsnMG string, dbMG string, cacheDSN string) error {
	ctx := context.Background()
	pool, err := pgxpool.Connect(ctx, dsnPG)
	if err != nil {
		log.Print(err)
		return err
	}

	clientMG, err := mongo.Connect(ctx, options.Client().ApplyURI(dsnMG))
	if err != nil {
		log.Print(err)
		return err
	}
	databaseMG := clientMG.Database(dbMG)

	securitySvc := security.NewService(pool, databaseMG)
	businessSvc := business.NewService(pool)
	mux := chi.NewMux()

	cache := &redis.Pool{
		DialContext: func(ctx context.Context) (redis.Conn, error) {
			return redis.DialURL(cacheDSN)
		},
	}

	application := app.NewServer(securitySvc, businessSvc, mux, cache)
	err = application.Init()
	if err != nil {
		log.Print(err)
		return err
	}

	server := &http.Server{
		Addr:    addr,
		Handler: application,
	}
	return server.ListenAndServe()
}
