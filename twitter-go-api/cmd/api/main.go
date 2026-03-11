package main

import (
	"context"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chanombude/twitter-go-api/internal/config"
	"github.com/chanombude/twitter-go-api/internal/db"
	"github.com/chanombude/twitter-go-api/internal/logger"
	"github.com/chanombude/twitter-go-api/internal/server"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config, err := config.LoadConfig(".")
	if err != nil {
		stdlog.Fatal("cannot load config:", err)
	}
	if err := config.ValidateForRuntime(); err != nil {
		stdlog.Fatal("invalid runtime config:", err)
	}

	logger.InitLogger(config.Environment)

	poolConfig, err := pgxpool.ParseConfig(config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot parse db config")
	}
	if config.DBMaxConns > 0 {
		poolConfig.MaxConns = config.DBMaxConns
	}
	if config.DBMinConns >= 0 {
		poolConfig.MinConns = config.DBMinConns
	}
	if config.DBMaxConnLifetimeMinutes > 0 {
		poolConfig.MaxConnLifetime = time.Duration(config.DBMaxConnLifetimeMinutes) * time.Minute
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot connect to db")
	}

	runDBMigration("file://db/migration", config.DBSource)

	var redisClient *redis.Client
	if config.RedisAddress != "" {
		redisOpt, err := redis.ParseURL(config.RedisAddress)
		if err != nil {
			log.Warn().Err(err).Msg("Invalid REDIS_ADDRESS, starting without Redis")
		} else {
			if config.RedisPassword != "" {
				redisOpt.Password = config.RedisPassword
			}
			client := redis.NewClient(redisOpt)
			pingCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			pingErr := client.Ping(pingCtx).Err()
			cancel()
			if pingErr != nil {
				log.Warn().Err(pingErr).Msg("Redis unavailable, starting without Redis")
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
		log.Fatal().Err(err).Msg("Cannot create server")
	}

	srv := server.HTTPServer(config.HTTPServerAddress)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP listen failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exiting")
}

func runDBMigration(migrationURL string, dbSource string) {
	migration, err := migrate.New(migrationURL, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create new migrate instance")
	}

	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("Failed to run migrate up")
	}

	log.Info().Msg("DB migrated successfully")
}
