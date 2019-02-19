package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

var mongoCollection *mongo.Collection

type MarketInfo struct {
	ID     string  `json:"name" bson:"_id"`
	Low    float64 `json:"low" bson:"low"`
	High   float64 `json:"high" bson:"high"`
	Volume float64 `json:"volume" bson:"volume"`
}

func getMarketInfo(w http.ResponseWriter, r *http.Request) {
	var marketInfo MarketInfo
	name := chi.URLParam(r, "name")

	result := mongoCollection.FindOne(context.Background(), bson.M{"_id": name})
	if result.Err() != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Printf("error querying market info (market %s): %+v", name, result.Err())
		return
	}
	err := result.Decode(&marketInfo)

	if err != nil {
		http.Error(w, http.StatusText(404), 404)
		log.Printf("error decoding the object (market %s): %+v", name, err)
		return
	}

	render.JSON(w, r, marketInfo)
}

func main() {

	mongoAddr := os.Getenv("MONGO_ADDRESS")
	var client *mongo.Client
	var err error

	for i := 0; i < 3; i++ {
		client, err = mongo.NewClient(mongoAddr)
		if err == nil {
			break
		}
		time.Sleep(1000 * time.Millisecond)
	}
	if err != nil {
		log.Fatal("Failed to connect to the database at", mongoAddr)
	}

	err = client.Connect(context.Background())
	if err != nil {
		log.Fatal("Failed to connect to the database")
	}
	defer client.Disconnect(nil)
	mongoCollection = client.Database("exchange").Collection("markets")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Route("/markets", func(r chi.Router) {
		r.Get("/{name}", getMarketInfo)
	})

	err = http.ListenAndServe(":80", r)

	if err != nil {
		log.Fatal("Failed to start the http server")
	}
}
