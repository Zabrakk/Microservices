package main

import (
	"encoding/json"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var channel	*amqp.Channel

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

// Parses the JSON encoded body of a message received from the RabbitMQ mp3 queue.
// The message's contents is then used to print out a notification.
// If JSON unmarshalling fails, an error is returned.
func Notify(body []byte) (err error) {
	var receivedMsg RabbitMQMessage
	err = json.Unmarshal(body, &receivedMsg)
	if err != nil { return err }
	log.Printf("Attention user %s! Your mp3 is ready for download.\nfid: %s\n", receivedMsg.Username, receivedMsg.Mp3Fid)
	return nil
}

func main() {
	log.Println("Notification service starting...")

	connection, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672")
	FailOnError(err, "Connecting to RabbitMQ failed")
	defer connection.Close()
	log.Println("Connected to RabbitMQ")

	channel, err = connection.Channel()
	FailOnError(err, "Opening a RabbitMQ channel failed")
	defer channel.Close()
	log.Printf("RabbitMQ channel opened")

	queue, err := channel.QueueDeclare(
		os.Getenv("MP3_QUEUE"),	// Name
		true,						// Durable
		false,						// Delete when unused
		false,						// Exclusive
		false,						// No-wait
		nil,						// Args
	)
	FailOnError(err, "RabbitMQ queue creation failed")

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
	go func() {
		for msg := range msgs {
			log.Printf("Received a message: %s", msg.Body)
			if err := Notify(msg.Body); err != nil {
				// Send nack if something goes wrong with convertion
				log.Println("Error occured while creating notification:\n", err.Error())
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