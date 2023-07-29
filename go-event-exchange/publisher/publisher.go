package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// Connect to RabbitMQ server
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Declare a topic exchange
	exchangeName := "event_exchange"
	err = ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,
	)
	failOnError(err, "Failed to declare an exchange")

	// Sample routing keys
	routingKeys := []string{"routing.event.donation_verified.to.santet.wa", "routing.event.donation_verified.to.suramadu.fbpixel"}

	// Publish messages with different routing keys
	for _, key := range routingKeys {
		message := fmt.Sprintf("Message with routing key '%s'", key)
		err = ch.Publish(
			exchangeName,
			key,
			false,
			false,
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(message),
			},
		)
		failOnError(err, "Failed to publish a message")
		fmt.Printf("Sent: %s\n", message)
	}
}
