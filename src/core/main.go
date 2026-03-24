package main

import (
	"log"
	"net/http"
)

func main() {
	registerRoutes()

	log.Println("Git-Canvas-Go running on http://localhost:6060")
	log.Fatal(http.ListenAndServe(":6060", nil))
}
