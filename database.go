package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB_CONTEXT_TIMEOUT = 90 * time.Second
var mongoClient *mongo.Client

var DatabaseName = "anomaly_detection"
var CollectionName = "sensor_data"
var pageSize int = 250

type PaginatedResponse struct {
	Items        []AnomalyDataVM
	TotalPages   int
	PageNumber   int
	TotalRecords int
}

// AnomalyData represents the structure of the data to be displayed
type AnomalyData struct {
	ID                string    `bson:"ID"`
	AccX              float64   `bson:"Accel_X"`
	AccY              float64   `bson:"Accel_Y"`
	AccZ              float64   `bson:"Accel_Z"`
	GyrX              float64   `bson:"Gyro_X"`
	GyrY              float64   `bson:"Gyro_Y"`
	GyrZ              float64   `bson:"Gyro_Z"`
	Latitude          float64   `bson:"Latitude"`
	Longitude         float64   `bson:"Longitude"`
	DateTime          time.Time `bson:"Time"`
	Speed             float64   `bson:"Speed"`
	VibrationDetected int8      `bson:"Vibration"`
	Temperature       float64   `bson:"Temperature"`
	CreatedAt         time.Time `bson:"CreatedAt"`
	Anomaly           string    `bson:"Anomaly"`
}

type AnomalyDataVM struct {
	ID          string
	AccelX      float64
	AccelY      float64
	AccelZ      float64
	GyroX       float64
	GyroY       float64
	GyroZ       float64
	Latitude    float64
	Longitude   float64
	Time        string
	Speed       float64
	Vibration   int8
	Temperature float64
	Anomaly     string
}

func ConnectDB() *mongo.Client {
	tlsConfig := &tls.Config{}
	tlsConfig.InsecureSkipVerify = true
	uri := strings.TrimSpace(os.Getenv("MONGODB_URL"))

	if uri == "" {
		log.Fatalln("No database URL was specified")
	}

	connectOptions := options.Client().ApplyURI(uri).SetTLSConfig(tlsConfig)
	// Create a new client and connect to the server

	ctx, cancel := context.WithTimeout(context.Background(), DB_CONTEXT_TIMEOUT)
	defer cancel()
	client, err := mongo.Connect(ctx, connectOptions)

	if err != nil {
		log.Fatalln("could not connect to database url, err - ", err)
	}
	log.Println("db connected at ConnectDB ✓")

	mongoClient = client

	return client
}

// Function to get data (replace with real API call to your cloud server)
func getAnomalyData(pageNumber int) (PaginatedResponse, error) {
	var anomalyData = make([]AnomalyData, 0)

	ctx := context.Background()
	coll := mongoClient.Database(DatabaseName).Collection(CollectionName)

	var skip int
	if pageNumber > 0 {
		skip = (pageNumber - 1) * pageSize
	} else {
		skip = 0
	}

	// Counting the number of records in the collection.
	totalRecords, err := coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		totalRecords = 0
	}

	totalPages := int(totalRecords) / pageSize
	if (totalRecords % 100) != 0 {
		totalPages++
	}

	sortingStage := bson.D{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: 1}}}}
	skipStage := bson.D{{Key: "$skip", Value: skip}}
	limitStage := bson.D{{Key: "$limit", Value: pageSize}}
	pipeline := mongo.Pipeline{sortingStage, skipStage, limitStage}

	listOptions := options.Aggregate().SetAllowDiskUse(true)
	cursor, err := coll.Aggregate(ctx, pipeline, listOptions)
	if err != nil {
		return PaginatedResponse{}, err
	}

	if err := cursor.All(ctx, &anomalyData); err != nil {
		return PaginatedResponse{}, err
	}

	var responseData = make([]AnomalyDataVM, 0, len(anomalyData))
	for _, v := range anomalyData {
		d := AnomalyDataVM{
			ID:          v.ID,
			AccelX:      v.AccX,
			AccelY:      v.AccY,
			AccelZ:      v.AccZ,
			GyroX:       v.GyrX,
			GyroY:       v.GyrY,
			GyroZ:       v.GyrZ,
			Latitude:    v.Latitude,
			Longitude:   v.Longitude,
			Time:        v.DateTime.Format("2006-01-02 15:04:05"),
			Speed:       v.Speed,
			Vibration:   v.VibrationDetected,
			Temperature: v.Temperature,
			Anomaly:     v.Anomaly,
		}

		responseData = append(responseData, d)
	}

	return PaginatedResponse{
		Items:        responseData,
		TotalPages:   totalPages,
		PageNumber:   pageNumber,
		TotalRecords: int(totalRecords),
	}, nil
}
