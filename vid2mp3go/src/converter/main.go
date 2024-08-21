package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// This function is for error checking. If the given err is not nil,
// then the msg will be logged along with the error string.
// Finally os.Exit(1) occurs. However, if err is nil, nothing happens.
func FailOnError(err error, msg string) {
	if err != nil {
		log.Println(msg)
		log.Fatal(err.Error())
	}
}

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
	FailOnError(err, "MongoUri creation failed")

	// Create MongoDB client / connection
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	FailOnError(err, "Creation of MongoDB connection failed")
	log.Println("Connected to mongoDB")

	// Create DB handles for the video and mp3 storage databases
	db_videos := client.Database("videos")
	db_mp3s := client.Database("mp3s")

	// Create gridFS buckets
	_, err = gridfs.NewBucket(db_videos, options.GridFSBucket().SetName("fs_videos"))
	FailOnError(err, "fs_videos creation failed")
	_, err = gridfs.NewBucket(db_mp3s, options.GridFSBucket().SetName("fs_mp3s"))
	FailOnError(err, "fs_mp3s creation failed")
	log.Println("Created GridFS buckets")

	// Connect to RabbitMQ
	connection, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672")
	FailOnError(err, "Connecting to RabbitMQ failed")
	defer connection.Close()
	log.Println("Connected to RabbitMQ")

	// Open a channel for message receiving
	channel, err := connection.Channel()
	FailOnError(err, "Opening a RabbitMQ channel failed")
	defer channel.Close()
	log.Printf("RabbitMQ channel opened")

	// Create a Queue for videos
	queue, err := channel.QueueDeclare(
		os.Getenv("VIDEO_QUEUE"),	// Name
		true,						// Durable
		false,						// Delete when unused
		false,						// Exclusive
		false,						// No-wait
		nil,						// Args
	)
	FailOnError(err, "RabbitMQ queue creation failed")

	// Start consuming messages from the queue
	msgs, err := channel.Consume(
		queue.Name, // Queue
		"",     	// Consumer
		true,   	// Auto-ack, TODO Change to false
		false,  	// Exclusive
		false,  	// No-local
		false,  	// No-wait
		nil,    	// Args
	)
	FailOnError(err, "Channel consume failed")

	var forever chan struct{}

	// Message handler
	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()
	log.Println("Waiting for messages")
	<-forever
}
