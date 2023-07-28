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

	// Declare a queue with a generated name
	q1, err := ch.QueueDeclare(
		"queue1",
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	// Declare a queue with a generated name
	q2, err := ch.QueueDeclare(
		"event.donation_verified.to.suramadu",
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,
	)
	failOnError(err, "Failed to declare a queue")

	// Bind the queue to the exchange with different routing keys
	err = ch.QueueBind(
		q1.Name,
		"routing.event.donation_verified",
		exchangeName,
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	err = ch.QueueBind(
		q2.Name,
		"routing.event.donation_verified",
		exchangeName,
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	// Consume messages from the queue
	msgsQ1, err := ch.Consume(
		q1.Name,
		"",
		true,  // auto-acknowledge
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	// Consume messages from the queue
	msgsQ2, err := ch.Consume(
		q2.Name,
		"",
		true,  // auto-acknowledge
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	failOnError(err, "Failed to register a consumer")

	// Start consuming messages
	forever := make(chan bool)
	go func() {
		for d := range msgsQ1 {
			message := string(d.Body)
			routingKey := d.RoutingKey
			fmt.Printf("Received a message Q1 with routing key '%s': %s\n", routingKey, message)
		}
	}()

	go func() {
		for d := range msgsQ2 {
			message := string(d.Body)
			routingKey := d.RoutingKey
			fmt.Printf("Received a message Q2 with routing key '%s': %s\n", routingKey, message)
		}
	}()

	fmt.Println("Consumer started. To exit, press CTRL+C")
	<-forever
}
