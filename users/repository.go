package users

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"os"
)

type Repository struct {
	Db         	*sql.DB
	Ctx         context.Context
	Queries     *Queries
}

var (
	POSTGRES_CONNECTION string
	Repo *Repository
)

func DBConnection() (*Repository, error) {
	if Repo != nil {
		return Repo, nil
	}
	// Load connection string
	POSTGRES_CONNECTION = os.Getenv("POSTGRES_CONNECTION")
	if POSTGRES_CONNECTION == "" {
		return nil, errors.New("no postgres connection string found")
	}

	// Open DB
	db, err := sql.Open("postgres", POSTGRES_CONNECTION)
	if err != nil {
		return nil, err
	}

	// Ping check
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	log.Println("Successfully connected to database")

	Repo = &Repository{
		Db:      db,
		Ctx:     context.Background(),
		Queries: New(db),
	}
	return Repo, nil
}