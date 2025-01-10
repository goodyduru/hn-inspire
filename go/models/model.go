package models

import (
	"container/list"
	"database/sql"
	"log"
	"os"

	"github.com/goodyduru/go-news/sessions"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB

func Init() {
	var err error
	db, err = sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	pder.sessions = make(map[string]*list.Element, 0)
	sessions.Register("memory", pder)
}
