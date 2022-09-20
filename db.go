package main

import (
	"database/sql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"os"
)

var db *sql.DB //база данных

func init() {
	e := godotenv.Load()
	if e != nil {
		panic("failed to load .env")
	}

	dbURL := "postgres://" +
		os.Getenv("DB_USERNAME") + ":" +
		os.Getenv("DB_PASSWORD") + "@" +
		os.Getenv("DB_HOST") + ":" +
		os.Getenv("DB_PORT") + "/" +
		os.Getenv("DB_NAME") +
		os.Getenv("DB_NAME") + "?sslmode=disable"

	conn, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic("failed to connect database")
	}

	db = conn
}

// Init возвращает дескриптор объекта DB
func Init() *sql.DB {
	return db
}
