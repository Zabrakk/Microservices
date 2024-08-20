package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetMongoUri() (mongoUri string, err error) {
	host, port := os.Getenv("MONGODB_HOST"), os.Getenv("MONGODB_PORT")
	if host == "" || port == "" {
		return "", errors.New("mongodb env variables were not set")
	}
	return fmt.Sprintf("mongodb://%s:%s", host, port), nil
}

func main() {
	log.Println("Converter service starting...")
	// Connect to MongoDB

	uri, err := GetMongoUri()
	if err != nil { log.Fatal(err.Error()) }

	//client
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	// mongodb+srv://<user>:<Password>@<host>:<port>

	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("Connected to mongoDB")


	db_videos := client.Database("videos")
	db_mp3s := client.Database("mp3s")

	_, err = gridfs.NewBucket(db_videos, options.GridFSBucket().SetName("fs_videos"))
	if err != nil {
		log.Fatalf("fs_videos creation failed:\n%s", err.Error())
	}
	_, err = gridfs.NewBucket(db_mp3s, options.GridFSBucket().SetName("fs_mp3s"))
	if err != nil {
		log.Fatalf("fs_mp3s creation failed:\n%s", err.Error())
	}
	log.Println("Created GridFS buckets")


}
