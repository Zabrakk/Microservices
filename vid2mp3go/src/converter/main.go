package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Connect to MongoDB
	uri := fmt.Sprintf("mongodb://%s:%s", os.Getenv("MONGODB_HOST"), os.Getenv("MONGODB_PORT"))
	//client
	_, err := mongo.Connect(context.TODO(), options.Client().
					ApplyURI(uri))
	// mongodb+srv://<user>:<Password>@<host>:<port>

	if err != nil {
		log.Fatal(err.Error())
	}

}
