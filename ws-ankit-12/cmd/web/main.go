package main

import (
	"log"
	"net/http"
	"ws-ankit-12/internal/handlers"
)

func main() {
   route := routes();
   log.Println("Started web server on port 8080")
   go handlers.ListenToWsChannel()
   _ = http.ListenAndServe(":8080", route)
}
