package main

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

const batchSize = 3

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"myQueue", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		panic(err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		panic(err)
	}

	batch := make([]amqp.Delivery, 0, batchSize)
	//lastMsgReceived := time.Now()
	for {
		timeout := time.After(3 * time.Second)
		// Wait for the specified amount of time
		select {
		case msg := <-msgs:
			batch = append(batch, msg)
			if len(batch) >= batchSize {
				fmt.Printf("masuk ketika batch lebih besar dari %d", batchSize)
				processBatch(batch)
				batch = batch[:0]
				msg.Ack(false)
			}
			//lastMsgReceived = time.Now()
		case <-timeout:
			fmt.Println("masuk ketika ga ada message selama 3 detik")
			processBatchAck(batch)
			batch = batch[:0]
		}
	}
}

type TestAja struct {
	CampaignId int
	Amount     int
	DonationId int
}

func processBatch(batch []amqp.Delivery) {
	// process messages in batch
	//grouping the structs by unique ID
	grouped := make(map[int]int)
	var testArr []TestAja
	var test TestAja

	fmt.Printf("Received batch of %d messages:\n", len(batch))
	for _, msg := range batch {
		json.Unmarshal(msg.Body, &test)
		testArr = append(testArr, test)
	}

	for _, v := range testArr {
		_, ok := grouped[v.CampaignId]
		if !ok {
			grouped[v.CampaignId] = 0
		}
		grouped[v.CampaignId] += v.Amount
	}

	//iterating over the map and print the sum for each id
	for id, sum := range grouped {
		fmt.Println("Sum of Amount for Campaign Id : ", id, "is", sum)
	}
}

func processBatchAck(batch []amqp.Delivery) {
	// process messages in batch
	//grouping the structs by unique ID
	grouped := make(map[int]int)
	var testArr []TestAja
	var test TestAja

	fmt.Printf("Received batch of %d messages:\n", len(batch))
	for _, msg := range batch {
		json.Unmarshal(msg.Body, &test)
		testArr = append(testArr, test)
		msg.Ack(false)
	}

	for _, v := range testArr {
		_, ok := grouped[v.CampaignId]
		if !ok {
			grouped[v.CampaignId] = 0
		}
		grouped[v.CampaignId] += v.Amount
	}

	//iterating over the map and print the sum for each id
	for id, sum := range grouped {
		fmt.Println("Sum of Amount for Campaign Id : ", id, "is", sum)
	}
}
