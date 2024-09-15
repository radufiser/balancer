package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	"gitlab.com/kiwicom/search-team/balancer/balancer"
	"gitlab.com/kiwicom/search-team/balancer/client"
	"gitlab.com/kiwicom/search-team/balancer/service"
)

func main() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	rand.Seed(time.Now().UnixNano())

	maxParallel := int32(50 + rand.Intn(150))
	log.Println("maxParallel", maxParallel)
	b := balancer.New(&service.TheExpensiveFragileService{}, maxParallel)

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	nbClients := 3
	log.Printf("Starting %d clients", nbClients)
	for i := 0; i < nbClients; i++ {
		go func() {
			workload := 500 + rand.Intn(1000)
			weight := 1 + rand.Intn(3)

			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
			b.Register(ctx, client.New(workload, weight))
		}()
	}
	<-ctx.Done()
	b.Shutdown()
}
