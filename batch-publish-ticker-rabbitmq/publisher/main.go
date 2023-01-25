package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/streadway/amqp"
)

const (
	batchSize = 5
	batchWait = 3 * time.Second
)

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
		false,     // delete when usused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		panic(err)
	}

	// buffer channel to temporarily store messages
	buffer := make(chan amqp.Publishing, batchSize)

	var test time.Duration
	var maxD = 3
	var minD = 1
	go func() {
		for msg := range buffer {
			test = time.Duration(rand.Intn(maxD-minD+1)+minD) * 500 * time.Millisecond
			time.Sleep(test)
			ch.Publish(
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				msg,
			)
			fmt.Printf("Sent message: %s\n", msg.Body)
		}
	}()

	type TestAja struct {
		CampaignId int
		Amount     int
		DonationId int
	}

	min := 100000
	max := 300000

	ticker := time.NewTicker(batchWait)
	defer ticker.Stop()
	for {
		var testAja = TestAja{
			CampaignId: 1,
			Amount:     rand.Intn(max-min) + min,
			DonationId: 1,
		}
		res, _ := json.Marshal(testAja)

		select {
		case <-ticker.C:
			// flush buffer
			for len(buffer) > 0 {
				<-buffer
			}
		case buffer <- amqp.Publishing{Body: res}:
			// message added to buffer
		}
	}
}
