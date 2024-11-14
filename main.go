package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	// "math"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Location struct {
	Province   string               `json:"province" bson:"province"`
	District   string               `json:"district" bson:"district"`
	SubDistrict string              `json:"sub_district" bson:"sub_district"`
	Village    string               `json:"village" bson:"village"`
	Border     []primitive.A        `json:"border" bson:"border"`
}

type GeoQuery struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

var client *mongo.Client
var collection *mongo.Collection

func connectMongo() {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("osm").Collection("local.Indo")
	fmt.Println("Connected to MongoDB!")
}

func findNearestLocation(w http.ResponseWriter, r *http.Request) {
	var geoQuery GeoQuery
	err := json.NewDecoder(r.Body).Decode(&geoQuery)
	if err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Perform MongoDB geospatial query using $nearSphere on `border` field
	filter := bson.M{
		"border": bson.M{
			"$nearSphere": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{geoQuery.Longitude, geoQuery.Latitude},
				},
				"$maxDistance": 10000, // Adjust the maximum search distance in meters as needed
			},
		},
	}

	var result Location
	err = collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		http.Error(w, "No nearby location found", http.StatusNotFound)
		return
	}

	// Send response with found location
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	connectMongo()
	http.HandleFunc("/api/nearest-path", findNearestLocation)
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
