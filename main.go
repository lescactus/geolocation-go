package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/lescactus/geolocation-go/controllers"
	"github.com/lescactus/geolocation-go/internal/api/ipapi"
	"github.com/lescactus/geolocation-go/repositories"
)

func main() {
	// Create in-memory database
	mdb := repositories.NewInMemoryDB()

	// Create redis database client
	rdb, err := repositories.NewRedisDB("redis://localhost:6379")
	if err != nil {
		log.Fatalln(err)
	}

	// Create ip-api client
	ipapi := ipapi.NewIPAPIClient(ipapi.DefaultBaseURL, http.DefaultClient)

	r := mux.NewRouter()
	h := controllers.NewBaseHandler(mdb, rdb, ipapi)

	r.Handle("/rest/v1/{ip}", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(h.GetGeoIP))).Methods("GET")
	r.Handle("/ready", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(h.Healthz))).Methods("GET")
	r.Handle("/alive", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(h.Healthz))).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", r))
}
