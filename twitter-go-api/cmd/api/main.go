package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chanombude/twitter-go-api/internal/config"
	"github.com/chanombude/twitter-go-api/internal/db"
	"github.com/chanombude/twitter-go-api/internal/logger"
	"github.com/chanombude/twitter-go-api/internal/server"
	"github.com/redis/go-redis/v9"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	logger.InitLogger(config.Environment)

	poolConfig, err := pgxpool.ParseConfig(config.DBSource)
	if err != nil {
		log.Fatal("cannot parse db config:", err)
	}
	poolConfig.MaxConns = 25
	poolConfig.MinConns = 25
	poolConfig.MaxConnLifetime = 5 * time.Minute

	conn, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	runDBMigration("file://db/migration", config.DBSource)

	var redisClient *redis.Client
	if config.RedisAddress != "" {
		redisOpt, err := redis.ParseURL(config.RedisAddress)
		if err != nil {
			log.Printf("warning: invalid REDIS_ADDRESS, starting without redis: %v", err)
		} else {
			if config.RedisPassword != "" {
				redisOpt.Password = config.RedisPassword
			}
			client := redis.NewClient(redisOpt)
			pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			pingErr := client.Ping(pingCtx).Err()
			cancel()
			if pingErr != nil {
				log.Printf("warning: redis unavailable, starting without redis: %v", pingErr)
				_ = client.Close()
			} else {
				redisClient = client
				defer redisClient.Close()
			}
		}
	}

	store := db.NewStore(conn)
	server, err := server.NewServer(config, store, redisClient)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	srv := server.HTTPServer(config.HTTPServerAddress)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

func runDBMigration(migrationURL string, dbSource string) {
	migrateURL := strings.Replace(dbSource, "postgresql://", "pgx5://", 1)
	migration, err := migrate.New(migrationURL, migrateURL)
	if err != nil {
		log.Fatal("cannot create new migrate instance:", err)
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("failed to run migrate up:", err)
	}

	log.Println("db migrated successfully")
}
