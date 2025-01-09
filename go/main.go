package main

import (
	"log"
	"net/http"

	"github.com/goodyduru/go-news/controllers"
	_ "github.com/goodyduru/go-news/memory"
	"github.com/goodyduru/go-news/models"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	models.Init()
	mux := controllers.Setup()
	log.Fatal(http.ListenAndServe(":8090", mux))
}
