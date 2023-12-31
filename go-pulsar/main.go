package main

import (
	"context"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"log"
)

func main() {
	// Pulsar client configuration
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://localhost:6650",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Topic to produce and consume messages
	topic := "my-topic"

	// Create a producer
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	// Create a consumer
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: "my-subscription",
		Type:             pulsar.Shared,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	// Send a message
	msg := "Hello, Pulsar!"
	_, err = producer.Send(context.Background(), &pulsar.ProducerMessage{
		Payload: []byte(msg),
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Message sent:", msg)

	// Consume messages
	for {
		msg, err := consumer.Receive(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Received message: %s\n", string(msg.Payload()))
		consumer.Ack(msg)
	}
}
