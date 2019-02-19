package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
)

const bittrexUrl string = "https://bittrex.com/api/v1.1/public/getmarketsummaries"

var marketsMap map[string]bool
var mongoCollection *mongo.Collection

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

	markets := strings.Split(os.Getenv("MARKETS"), ",")
	// to speed-up market look-up, create a map of markets
	marketsMap = make(map[string]bool)
	for _, market := range markets {
		marketsMap[market] = true
	}

	ticker := time.NewTicker(10500 * time.Millisecond)

	go func() {
		for t := range ticker.C {
			collectData()
			log.Println("Tick at", t)
		}
	}()

	// wait forever
	select {}
}

func collectData() {
	updateOpt := options.Update()
	updateOpt.SetUpsert(true)

	result, err := doGetRequest(bittrexUrl)
	if err != nil {
		log.Printf("failed to load markets data, cannot proceed!")
		return
	}

	resultParsed, _ := gabs.ParseJSON(result)

	if !resultParsed.Path("success").Data().(bool) {
		log.Printf(
			"markets api returned error: %s!",
			resultParsed.Path("success").Data().(string),
		)
		return
	}

	children, _ := resultParsed.S("result").Children()
	for _, market := range children {
		name := market.Path("MarketName").Data().(string)
		_, found := marketsMap[name]
		if !found {
			continue
		}
		high := market.Path("High").Data().(float64)
		low := market.Path("Low").Data().(float64)
		volume := market.Path("Volume").Data().(float64)
		log.Printf("%s: [%0.6f - %0.6f] x %0.6f\n", name, low, high, volume)

		newDoc := bson.D{
			{"low", low},
			{"high", high},
			{"volume", volume},
		}
		_, err := mongoCollection.UpdateOne(
			context.Background(),
			bson.D{{"_id", name}},
			bson.D{{"$set", newDoc}},
			updateOpt,
		)

		if err != nil {
			log.Fatal("insert failed %+v", err)
		}

	}
}

func doGetRequest(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
