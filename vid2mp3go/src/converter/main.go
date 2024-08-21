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

// TODO: func failOnError

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

	// Create MongoDB client / connection
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

	if err != nil { log.Fatal(err.Error()) }
	log.Println("Connected to mongoDB")

	// Create DB handles for the video and mp3 storage databases
	db_videos := client.Database("videos")
	db_mp3s := client.Database("mp3s")

	// Create gridFS buckets
	_, err = gridfs.NewBucket(db_videos, options.GridFSBucket().SetName("fs_videos"))
	if err != nil {
		log.Fatalf("fs_videos creation failed:\n%s", err.Error())
	}
	_, err = gridfs.NewBucket(db_mp3s, options.GridFSBucket().SetName("fs_mp3s"))
	if err != nil {
		log.Fatalf("fs_mp3s creation failed:\n%s", err.Error())
	}
	log.Println("Created GridFS buckets")

	// Connect to RabbitMQ
	connection, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672")
	if err != nil { log.Fatal(err.Error()) }
	defer connection.Close()
	log.Println("Connected to RabbitMQ")

	// Open a channel for message receiving
	channel, err := connection.Channel()
	if err != nil { log.Fatal(err.Error()) }
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
	if err != nil { log.Fatal(err.Error()) }

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
	if err != nil { log.Fatal(err.Error()) }

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
