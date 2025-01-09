package models

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB
var globalSessions *Manager

func Init() {
	var err error
	db, err = sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	globalSessions, _ = NewManager("memory", "hnsessionid", 3600)
	go globalSessions.GC()
}

func StartSession(w http.ResponseWriter, r *http.Request) Session {
	return globalSessions.SessionStart(w, r)
}
