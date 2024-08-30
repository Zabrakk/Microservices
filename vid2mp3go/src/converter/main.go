package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var dbVideos, dbMp3s	*mongo.Database
var fsVideos, fsMp3s	*gridfs.Bucket
var channel				*amqp.Channel

type RabbitMQMessage struct {
	VideoFid	string		`json:"video_fid"`
	Mp3Fid		string		`json:"mp3_fid"`
	Username	string		`json:"username"`
}

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

func ConvertToMp3(body []byte) (err error) {
	var receivedMsg RabbitMQMessage
	json.Unmarshal(body, &receivedMsg)
	log.Printf("VideoFid %s, Mp3Fid %s, Username %s \n", receivedMsg.VideoFid, receivedMsg.Mp3Fid, receivedMsg.Username)

	log.Println("Creating temp files for video and audio")

	tempVideoFile, err := os.CreateTemp(".", "video")
	if err != nil { return err }
	tempAudioFile, err := os.CreateTemp(".", "*.mp3")
	if err != nil { return err }
	defer os.Remove(tempVideoFile.Name())
	defer os.Remove(tempAudioFile.Name())

	log.Println("Getting ID from Hex")
	id, err := primitive.ObjectIDFromHex(receivedMsg.VideoFid)
	if err != nil { return err }

	log.Println("Creating GridFS buckets")
	fsVideos, err = gridfs.NewBucket(dbVideos, options.GridFSBucket())
	if err != nil { return err }
	fsMp3s, err = gridfs.NewBucket(dbMp3s, options.GridFSBucket())
	if err != nil { return err }

	log.Println("Downloading video")
	n, err := fsVideos.DownloadToStream(id, tempVideoFile)
	if err != nil { return err }
	log.Printf("Downloaded %d bytes\n", n)

	log.Println("Extracting audio from video")
	cmd := exec.Command("ffmpeg", "-y", "-i", tempVideoFile.Name(), "-q:a", "0", "-map", "a", tempAudioFile.Name())
	err = cmd.Run()
	if err != nil { return err }

	log.Println("Saving audio to MongoDB")
	audioFid, err := fsMp3s.UploadFromStream(tempAudioFile.Name(), tempAudioFile)
	if err != nil { return err }
	log.Printf("Audio uploaded with FID %s\n", audioFid)

	receivedMsg.Mp3Fid = audioFid.Hex()
	body, err = json.Marshal(receivedMsg)
	if err != nil { return err }

	log.Println("Publishing message to RabbitMQ")
	err = channel.Publish(
		"",						// Exchange
		os.Getenv("MP3_QUEUE"),	// Routing key
		false,					// Mandatory
		false,					// Immediate
		amqp.Publishing{		// Msg
			ContentType: "application/json",
			Body: body,
		},
	)
	if err != nil { return err }

	log.Println("Done")
	return nil
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
	dbVideos = client.Database("videos")
	dbMp3s = client.Database("mp3s")

	// Create gridFS buckets
	/*
	fsVideos, err = gridfs.NewBucket(db_videos, options.GridFSBucket().SetName("fs_videos"))
	FailOnError(err, "fs_videos creation failed")
	fsMp3s, err = gridfs.NewBucket(db_mp3s, options.GridFSBucket().SetName("fs_mp3s"))
	FailOnError(err, "fs_mp3s creation failed")
	log.Println("Created GridFS buckets")
	*/

	// Connect to RabbitMQ
	connection, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672")
	FailOnError(err, "Connecting to RabbitMQ failed")
	defer connection.Close()
	log.Println("Connected to RabbitMQ")

	// Open a channel for message receiving
	channel, err = connection.Channel()
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
		false,   	// Auto-ack
		false,  	// Exclusive
		false,  	// No-local
		false,  	// No-wait
		nil,    	// Args
	)
	FailOnError(err, "Channel consume failed")

	var forever chan struct{}
	// Message handler
	go func() {
		for msg := range msgs {
			log.Printf("Received a message: %s", msg.Body)
			if err := ConvertToMp3(msg.Body); err != nil {
				// Send nack if something goes wrong with convertion
				log.Println("Error occured during video convertion process:\n", err.Error())
				channel.Ack(msg.DeliveryTag, true)
				//channel.Nack(msg.DeliveryTag, false, true)
			} else {
				channel.Ack(msg.DeliveryTag, true)
			}
		}
	}()
	log.Println("Waiting for messages...")
	<-forever
}
