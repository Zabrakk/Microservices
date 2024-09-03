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

	connection := ConnectToRabbitMQ()
	defer connection.Close()

	channel = OpenChannel(connection)
	defer channel.Close()

	queue := CreateQueue(channel, os.Getenv("MP3_QUEUE"))

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