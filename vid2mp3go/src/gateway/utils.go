package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
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

// This function reads the MONGODB_HOST and MONGODB_PORT env variables
// and uses them to create the mongoUri. If the env variables have not been
// set, this function returns an error.
func GetMongoUri() (mongoUri string, err error) {
	host, port := os.Getenv("MONGODB_HOST"), os.Getenv("MONGODB_PORT")
	if host == "" || port == "" {
		return "", errors.New("mongodb env variables were not set")
	}
	return fmt.Sprintf("mongodb://%s:%s", host, port), nil
}

// Connects to MongoDB and returns a *mongo.client.
// The URI for the DB is to be given with the mongoUri param.
// Uses FailOnError to handle errors.
func ConnectToMongoDB(mongoUri string) (client *mongo.Client) {
	log.Println("Connecting to MongoDB")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoUri))
	FailOnError(err, "Creation of MongoDB connection failed")
	return client
}

// Connects to RabbitMQ and returns a *amqp.Connection.
// Uses FailOnError to handle errors.
func ConnectToRabbitMQ() (connection *amqp.Connection) {
	log.Println("Connecting to RabbitMQ")
	connection, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672")
	FailOnError(err, "Connecting to RabbitMQ failed")
	return connection
}

// Opens a channel for receiving messages from RabbitMQ.
// Returns *amqp.Channel. Uses FailOnError to handle errors.
func OpenChannel(connection *amqp.Connection) (channel *amqp.Channel) {
	log.Println("Opening RabbitMQ channel")
	channel, err := connection.Channel()
	FailOnError(err, "Opening RabbitMQ channel failed")
	return channel
}

// Creates and returns a amqp.Queue. The parameter name will be used as the
// queue's name. Uses FailOnError to handle errors.
func CreateQueue(channel *amqp.Channel, name string) (queue amqp.Queue) {
	log.Println("Creating RabbitMQ queue")
	queue, err := channel.QueueDeclare(
		name,		// Name
		true,		// Durable
		false,		// Delete when unused
		false,		// Exclusive
		false,		// No-wait
		nil,		// Args
	)
	FailOnError(err, "RabbitMQ queue creation failed")
	return queue
}