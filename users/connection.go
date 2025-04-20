package users

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
)

type Connection struct {
	Db         	*sql.DB
	Ctx         context.Context
	Queries     *Queries
}

var (
	postgresConnectionURL string
	conn *Connection
)

func DatabaseConnection() (*Connection, error) {
	if conn != nil {
		return conn, nil
	}
	// Load connection string
	postgresConnectionURL = os.Getenv("POSTGRES_CONNECTION")
	if postgresConnectionURL == "" {
		return nil, errors.New("no postgres connection string found")
	}

	// Open DB
	db, err := sql.Open("postgres", postgresConnectionURL)
	if err != nil {
		return nil, err
	}

	// Ping check
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to database")

	conn = &Connection{
		Db:      db,
		Ctx:     context.Background(),
		Queries: New(db),
	}
	return conn, nil
}