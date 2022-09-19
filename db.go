package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"os"
)

func Init() *sql.DB {
	e := godotenv.Load()
	if e != nil {
		panic("failed to load .env")
	}

	dbURL := "postgres://" +
		os.Getenv("DB_USERNAME") + ":" +
		os.Getenv("DB_PASSWORD") + "@" +
		os.Getenv("DB_HOST") + ":" +
		os.Getenv("DB_PORT") + "/" +
		os.Getenv("DB_NAME")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic("failed to connect database")
	}

return db
}

