package main

import (
	"fmt"
	"os"

	"context"
	"github.com/jackc/pgx/v5"
)


func Connect() (*pgx.Conn, error) {
	containerName := os.Getenv("DB_CONTAINER_NAME")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PWD")
	schema := os.Getenv("DB_SCHEMA")
	port := os.Getenv("DB_PORT")

	url := "postgres://" + user + ":" + pass + "@" + containerName + ":" + port + "/" + schema

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
