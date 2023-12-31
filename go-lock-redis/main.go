package main

import (
	"context"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/jackc/pgx/v4/pgxpool"
	goredislib "github.com/redis/go-redis/v9"
	"log"
	"sync"
	"time"
)

var (
	ctx       = context.Background()
	redisAddr = "localhost:6379" // Replace with your Redis server address
	redisPwd  = ""               // Replace with your Redis server password, if any
	redisDB   = 0                // Replace with your Redis database number
	pgURL     = "postgres://root:pass@localhost:5492/test?sslmode=disable"
)

func main() {
	// Create a pool with go-redis (or redigo) which is the pool redisync will
	// use while communicating with Redis. This can also be any pool that
	// implements the `redis.Pool` interface.
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     redisAddr,
		Password: redisPwd,
		DB:       redisDB,
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)

	// Create an instance of redisync to be used to obtain a mutual exclusion
	// lock.
	rs := redsync.New(pool)

	pgPool, err := pgxpool.Connect(ctx, pgURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pgPool.Close()

	// Initialize the ticket count in PostgreSQL
	_, err = pgPool.Exec(ctx, "CREATE TABLE IF NOT EXISTS tickets (count INT)")
	if err != nil {
		log.Fatal(err)
	}

	_, err = pgPool.Exec(ctx, "INSERT INTO tickets (count) VALUES ($1) ON CONFLICT DO NOTHING", 15)
	if err != nil {
		log.Fatal(err)
	}

	// Simulate concurrent ticket reservation
	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		defer func() {
			wg.Done()
		}()

		// Create a Redsync mutex
		mutex := rs.NewMutex("ticket_lock")

		// Configure retry parameters
		maxRetries := 5
		retryInterval := time.Second

		for j := 0; j < maxRetries; j++ {
			// Attempt to acquire the lock
			if err := mutex.Lock(); err != nil {
				log.Printf("Error acquiring lock (attempt %d): %v\n", j+1, err)
				time.Sleep(retryInterval)
				continue
			}

			// Lock acquired, proceed with ticket reservation
			defer func() {
				// Release the lock
				if _, err := mutex.Unlock(); err != nil {
					log.Printf("Error releasing lock: %v\n", err)
				}
			}()

			// Get current ticket count from PostgreSQL
			var count int
			err := pgPool.QueryRow(ctx, "SELECT count FROM tickets FOR UPDATE").Scan(&count)
			if err != nil {
				log.Printf("Error getting ticket count: %v\n", err)
				return
			}

			// Simulate some processing time
			time.Sleep(100 * time.Millisecond)

			// Reserve ticket
			if count > 0 {
				count--
				_, err := pgPool.Exec(ctx, "UPDATE tickets SET count = $1", count)
				if err != nil {
					log.Printf("Error updating ticket count: %v\n", err)
					return
				}
				fmt.Printf("Ticket reserved by %d after %d attempts\n", i, j+1)
				return
			} else {
				fmt.Printf("No more tickets available for %d after %d attempts\n", i, j+1)
				return
			}
		}

		// Retry limit reached
		log.Printf("Max retries reached. Unable to acquire lock for %d\n", i)
	}

}
