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
	// Publisher
	publisher()
	// Subscriber
	subscriber()
}

func publisher() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	queueName := "main_queue"
	err = ch.ExchangeDeclare(
		queueName, // exchange name
		"direct",  // exchange type
		true,      // durable
		false,     // auto-deleted
		false,     // internal
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	args := amqp.Table{
		"x-dead-letter-exchange":    "dlx_exchange",
		"x-dead-letter-routing-key": "dlq",
	}
	_, err = ch.QueueDeclare(
		queueName, // queue name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		args,      // arguments with DLX configuration
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		queueName, // queue name
		"key",     // routing key
		queueName, // exchange name
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")

	body := "Trx Id #1 Publish Pay Created!"
	err = ch.Publish(
		queueName, // exchange name
		"key",     // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
			Expiration:  "5000", // assume expire time
		},
	)

	failOnError(err, "Failed to publish a message")
	fmt.Printf(" [x] Sent %s\n", body)
}

func subscriber() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	dlxName := "dlx_exchange"
	err = ch.ExchangeDeclare(
		dlxName,  // exchange name
		"direct", // exchange type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare DLX exchange")

	args := amqp.Table{
		"x-dead-letter-exchange":    "dlx_exchange",
		"x-dead-letter-routing-key": "dlq",
	}

	dlqName := "dlq"
	q1, err := ch.QueueDeclare(
		dlqName, // queue name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare DLQ queue")

	err = ch.QueueBind(
		dlqName, // queue name
		"dlq",   // routing key
		dlxName, // exchange name
		false,
		nil,
	)
	failOnError(err, "Failed to bind DLQ queue")

	queueName := "main_queue"
	q, err := ch.QueueDeclare(
		queueName, // queue name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		args,      // arguments
	)
	failOnError(err, "Failed to declare a queue")

	fmt.Println(q)

	//msgs, err := ch.Consume(
	//	q.Name, // queue name
	//	"",     // consumer
	//	false,  // auto-ack
	//	false,  // exclusive
	//	false,  // no-local
	//	false,  // no-wait
	//	nil,    // args
	//)
	//failOnError(err, "Failed to register a consumer")

	msgs1, err := ch.Consume(
		q1.Name, // queue name
		"",      // consumer
		false,   // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)

	forever := make(chan bool)

	//go func() {
	//	for d := range msgs {
	//		fmt.Printf("payment created. main queue Received a message, wont ack. let it expire to go to DLQ: %s\n", d.Body)
	//	}
	//}()

	go func() {
		for d := range msgs1 {
			fmt.Printf("after 5 seconds, dlq queue received a message because message not consumed in mainq ueue. "+
				"payment will be checked with expire status : %s\n", d.Body)
			d.Ack(false)
		}
	}()

	fmt.Println("Waiting for messages. To exit press CTRL+C")
	<-forever
}
