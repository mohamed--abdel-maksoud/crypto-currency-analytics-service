package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

var mongoCollection *mongo.Collection

type MarketInfo struct {
	Market string  `json:"market" bson:"market"`
	Low    float64 `json:"low" bson:"low"`
	High   float64 `json:"high" bson:"high"`
	Volume float64 `json:"volume" bson:"volume"`
	Time   int64   `json:"time" bson:"time"`
}

type OutputMarketInfo struct {
	Market string  `json:"market"`
	Low    float64 `json:"low"`
	High   float64 `json:"high"`
	Volume float64 `json:"volume"`
}

type OutputMarketInfoInterval struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Data   []OutputMarketInfo `json:"data"`
}


func getMarketInfo(w http.ResponseWriter, r *http.Request) {
    var interval int64 = 5
	var marketInfo MarketInfo
	market := chi.URLParam(r, "market")

    log.Printf("from=%s", r.URL.Query().Get("from"))
    timeMinObj, err :=  time.Parse(time.RFC3339, r.URL.Query().Get("from"))
    if err != nil {
		http.Error(w, "invalid `from` parameter, needs to be in RFC3339 ", 400)
        return
    }

    timeMaxObj, err :=  time.Parse(time.RFC3339, r.URL.Query().Get("to"))
    if err != nil {
		http.Error(w, "invalid `to` parameter, needs to be in RFC3339 ", 400)
        return
    }

    timeMin := timeMinObj.Unix()
    timeMax := timeMaxObj.Unix()

    query := bson.M{
        "time": bson.M{
            "$gte": timeMin,
            "$lte": timeMax,
        },
    }

    if len(market) > 0 {
        query["market"] = market
    }

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := mongoCollection.Find(ctx, query)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Printf("error querying market info (query %+v): %+v", query, err)
		return
	}
	defer cursor.Close(ctx)

    // try to preallocate with correct capacity
    marketInfosMap := make(map[int64]*OutputMarketInfoInterval, (timeMax - timeMin) / 5)

	for cursor.Next(ctx) {
        err := cursor.Decode(&marketInfo)
        if err != nil {
            log.Printf("error decoding the object (query %+v): %+v", query, err)
            continue
        }
        t := marketInfo.Time / interval
        _, found := marketInfosMap[t]
        if !found {
            marketInfosMap[t] = &OutputMarketInfoInterval{
                From: time.Unix(marketInfo.Time, 0).UTC().Format(time.RFC3339),
                To: time.Unix(marketInfo.Time + interval, 0).UTC().Format(time.RFC3339),
                Data: make([]OutputMarketInfo, 0, 1),
            }
        }

        marketInfosMap[t].Data = append(marketInfosMap[t].Data, OutputMarketInfo{
            Market: marketInfo.Market,
            Low: marketInfo.Low,
            High: marketInfo.High,
            Volume: marketInfo.Volume,
        })
    }

    marketInfos := make([]*OutputMarketInfoInterval, 0, len(marketInfosMap))
    keys := make([]int, 0, len(marketInfosMap))

    for key, _ := range marketInfosMap {
        keys = append(keys, int(key))
    }
    sort.Ints(keys)
    for _, key := range keys {
        marketInfos = append(marketInfos, marketInfosMap[int64(key)])
    }

	render.JSON(w, r, marketInfos)
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
		r.Get("/", getMarketInfo)
		r.Get("/{market}", getMarketInfo)
	})

	err = http.ListenAndServe(":80", r)

	if err != nil {
		log.Fatal("Failed to start the http server")
	}
}
