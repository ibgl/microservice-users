package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ibgl/microservice-users/internal/app"
	"github.com/ibgl/microservice-users/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	ctx := context.Background()
	viper.AutomaticEnv()

	//init logger
	f, err := os.OpenFile("logs/app-log.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Failed to create logfile " + "app-log.json")
		panic(err)
	}
	defer f.Close()

	logger := &logrus.Logger{
		Out:   io.MultiWriter(f, os.Stdout),
		Level: logrus.DebugLevel,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		},
	}

	postgresDB := viper.GetString("POSTGRES_DB")
	if postgresDB == "" {
		fmt.Println("POSTGRES_DB configuration must be provided")
		os.Exit(1)
	}

	postgresUser := viper.GetString("POSTGRES_USER")
	if postgresUser == "" {
		fmt.Println("POSTGRES_USER configuration must be provided")
		os.Exit(1)
	}

	postgresPass := viper.GetString("POSTGRES_PASSWORD")
	if postgresPass == "" {
		fmt.Println("POSTGRES_PASSWORD configuration must be provided")
		os.Exit(1)
	}

	pgHost := viper.GetString("POSTGRES_HOST")
	if pgHost == "" {
		fmt.Println("POSTGRES_HOST configuration must be provided")
		os.Exit(1)
	}

	postgresString := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", postgresUser, postgresPass, pgHost, postgresDB)

	var pool *pgxpool.Pool

	pool, err = pgxpool.New(ctx, postgresString)
	if err != nil {
		logger.Errorf("pgxpool init error %v", err)
		os.Exit(1)
	}

	jwtSecret := viper.GetString("JWT_SECRET")
	if jwtSecret == "" {
		logger.Errorf("JWT configuration must be provided %v", err)
		os.Exit(1)
	}

	attl := viper.GetInt("JWT_ACCESS_TTL")
	if attl == 0 {
		logger.Errorf("JWT_ACCESS_TTL configuration must be provided %v", err)
		os.Exit(1)
	}

	rttl := viper.GetInt("JWT_REFRESH_TTL")
	if rttl == 0 {
		logger.Errorf("JWT_REFRESH_TTL configuration must be provided %v", err)
		os.Exit(1)
	}

	googleKey := viper.GetString("GOOGLE_KEY")
	if googleKey == "" {
		logger.Errorf("GOOGLE_KEY configuration must be provided %v", err)
		os.Exit(1)
	}

	maxSessions := viper.GetInt("MAX_USER_SESSIONS")
	if err != nil {
		logger.Errorf("MAX_USER_SESSIONS configuration must be provided %v", err)
		os.Exit(1)
	}

	//init application
	appConfig := app.Config{
		JwtSecret:       jwtSecret,
		JwtAccessTTL:    attl,
		JwtRefreshTTL:   rttl,
		MaxUserSessions: maxSessions,
		GoogleKey:       googleKey,
	}

	app, err := app.NewApplication(&appConfig, logger, pool)
	if err != nil {
		logger.Errorf("Application init error %v", err)
		os.Exit(1)
	}

	server := ports.NewHttpServer(app)
	server.Start()
}

type QueryTracer struct {
	logger *logrus.Logger
}

func (h QueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	h.logger.Info("Query log query start", data)
	return ctx
}

func (h QueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	h.logger.Info("Query log query end", data)
}
