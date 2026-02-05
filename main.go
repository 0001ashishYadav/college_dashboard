package main

import (
	"context"
	"dashboard/api"
	"dashboard/db/pgdb"
	"dashboard/token"
	"dashboard/utils"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// config
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("failed to read env ", err)
	}
	// db
	poolConfig, err := pgxpool.ParseConfig(config.DatabaseURL)
	if err != nil {
		log.Fatalln("Unable to parse DATABASE_URL", err)
	}
	dbConn, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalln("Unable to createt connection pool", err)
	}

	store := pgdb.NewStore(dbConn)
	defer dbConn.Close()

	// token
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		log.Fatal("failed to create token maker", err)
	}

	server, err := api.NewServer(config, store, tokenMaker)
	if err != nil {
		log.Fatal("cannot start server", err)
	}

	err = server.Start(config.Port)
	if err != nil {
		log.Fatal("connot start server", err)
	}
}
