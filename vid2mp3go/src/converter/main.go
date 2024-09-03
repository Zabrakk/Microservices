package main

import (
	"encoding/json"
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

func ConvertToMp3(body []byte) (err error) {
	// Get JSON data from message body
	var receivedMsg RabbitMQMessage
	json.Unmarshal(body, &receivedMsg)

	log.Printf("VideoFid %s, Mp3Fid %s, Username %s \n", receivedMsg.VideoFid, receivedMsg.Mp3Fid, receivedMsg.Username)
	log.Println("Creating temp files for video and audio")

	// Create temp files for video and audio storage
	tempVideoFile, err := os.CreateTemp(".", "video")
	if err != nil { return err }
	tempAudioFile, err := os.CreateTemp(".", "*.mp3")
	if err != nil { return err }
	// Delete temp files when this function finishes
	defer os.Remove(tempVideoFile.Name())
	defer os.Remove(tempAudioFile.Name())

	log.Println("Getting ID from Hex")
	id, err := primitive.ObjectIDFromHex(receivedMsg.VideoFid)
	if err != nil { return err }

	// Creates GridFS buckets for file downloads and uploads
	log.Println("Creating GridFS buckets")
	fsVideos, err = gridfs.NewBucket(dbVideos, options.GridFSBucket())
	if err != nil { return err }
	fsMp3s, err = gridfs.NewBucket(dbMp3s, options.GridFSBucket())
	if err != nil { return err }

	// Get video based on id from MongoDB
	log.Println("Downloading video")
	n, err := fsVideos.DownloadToStream(id, tempVideoFile)
	if err != nil { return err }
	log.Printf("Downloaded %d bytes\n", n)

	// Use ffmpeg to extract audio from tempVideoFile and save it into tempAudioFile
	log.Println("Extracting audio from video")
	cmd := exec.Command("ffmpeg", "-y", "-i", tempVideoFile.Name(), "-q:a", "0", "-map", "a", tempAudioFile.Name())
	err = cmd.Run()
	if err != nil { return err }

	// Upload the extracted audio to MongoDB and get its FID.
	log.Println("Saving audio to MongoDB")
	audioFid, err := fsMp3s.UploadFromStream(tempAudioFile.Name(), tempAudioFile)
	if err != nil { return err }
	log.Printf("Audio uploaded with FID %s\n", audioFid)

	// Add the audio's FID to the JSON
	receivedMsg.Mp3Fid = audioFid.Hex()
	body, err = json.Marshal(receivedMsg)
	if err != nil { return err }

	// Send the JSON to MP3 queue, informing the notification service of the audio's creation
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

	uri, err := GetMongoUri()
	FailOnError(err, "MongoUri creation failed")

	client := ConnectToMongoDB(uri)

	// Create DB handles for the video and mp3 storage databases
	dbVideos = client.Database("videos")
	dbMp3s = client.Database("mp3s")

	connection := ConnectToRabbitMQ()
	defer connection.Close()

	channel = OpenChannel(connection)
	defer channel.Close()

	queue := CreateQueue(channel, os.Getenv("VIDEO_QUEUE"))

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
