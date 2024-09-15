package balancer

import (
	"context"
	"log"
	"sync"

	"gitlab.com/kiwicom/search-team/balancer/work"
)

// Client of our service.
// It produces requests that need to be processed by the Server.
type Client interface {
	// Weight is unit-less number that determines how much processing capacity can a client be allocated
	// when running in parallel with other clients.
	// The higher the weight, the more capacity the client receives.
	// Weight is greater than or equal to 1.
	Weight() int
	// Workload returns a channel of requests that are meant to be processed by the Server.
	// Client's channel is always filled with at least one request.
	// Each request takes 1 capacity unit while it is being processed by the server.
	Workload(ctx context.Context) chan *work.Request
}

// Server processes requests from clients.
// It is an expensive service that we need to utilise as much as we can, but only up to the limits agreed
// with the owners of the service.
type Server interface {
	// Process takes one requestID and does something with it.
	Process(ctx context.Context, request *work.Request) error
}

// Balancer makes sure the Server is not smashed with incoming requests.
// It limits the number of requests processed by the Server in parallel.
// Imagine there's an SLO defined, and we don't want to make the owners of the expensive service angry.
//
// If implementing more advanced balancer, make sure to correctly assign processing capacity to a client based on other
// clients currently in process.
// To give an example of this, imagine there's a maximum number of parallel requests set to 100 and
// there are two clients registered, both with the same weight.
// When they are both served in parallel, each of them gets to send 50 requests at the same time.
// In the same scenario, if there were two clients with weight 1 and one client with weight 2,
// the first two would be allowed to send 25 requests and the other one would send 50.
// It's likely that the one sending 50 would be served faster, finishing the work early, with only two clients
// remaining.
// At that point, the server's capacity should be divided between the two remaining clients, allowing them to
// send 50 parallel requests each.
type Balancer struct {
	maxParallelRequests int32
	server              Server
	weightedSlice       *WeightedSlice[Client]
	workQueue           chan *work.Request
	clientsChan         map[Client]chan *work.Request
	mu                  sync.Mutex
	wg                  sync.WaitGroup
	cancel              context.CancelFunc
}

// New creates a new Balancer instance.
// The balancer needs the server where it sends requests to and a maximum number of parallel requests that should be
// in flight at any given time.
// to the server.
// THIS IS A HARD REQUIREMENT - THE SERVICE CANNOT PROCESS MORE THAN maxParallelRequests IN PARALLEL.
func New(server Server, maxParallelRequests int32) *Balancer {
	ctx, cancel := context.WithCancel(context.Background())
	b := Balancer{
		server:              server,
		maxParallelRequests: maxParallelRequests,
		weightedSlice:       NewWeightedSlice[Client](),
		workQueue:           make(chan *work.Request, maxParallelRequests),
		clientsChan:         make(map[Client]chan *work.Request),
		cancel:              cancel,
	}

	b.workerPool(int(maxParallelRequests), ctx)
	go b.start(ctx)
	return &b
}

// Register a client to the balancer and start dispatching its requests to the server.
// Assume that one client can register itself multiple times.
func (b *Balancer) Register(ctx context.Context, client Client) {
	b.mu.Lock()
	defer b.mu.Unlock()

	log.Printf("Registering client %+v\n", client)
	b.weightedSlice.AddItem(WeightedItem[Client]{Value: client, Weight: client.Weight()})
	reqChan := client.Workload(ctx)
	b.clientsChan[client] = reqChan
}

// Start dispatching requests to the server in a balanced way.
func (b *Balancer) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Balancer shutting down")
			return
		default:
			if b.weightedSlice.Len() > 0 {
				client := b.weightedSlice.GetRandomItem()
				log.Printf("Client %+v \n", client)
				ch, ok := b.clientsChan[client]
				if ok {
					select {
					case req, more := <-ch:
						if more {
							b.workQueue <- req
						} else {
							// Client channel closed
							b.cleanupClient(client)
						}
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}
}

// Worker processes tasks from the work queue.
func (b *Balancer) worker(i int, ctx context.Context) {
	defer b.wg.Done()

	for {
		select {
		case task := <-b.workQueue:
			log.Printf("Worker %d processing task: %+v \n", i, task)
			err := b.server.Process(ctx, task)
			if err != nil {
				log.Printf("Worker %d encountered error processing task: %v\n", i, err)
			}
		case <-ctx.Done():
			log.Printf("Worker %d shutting down", i)
			return
		}
	}
}

// WorkerPool creates a fixed number of workers.
func (b *Balancer) workerPool(numWorkers int, ctx context.Context) {
	b.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go b.worker(i, ctx)
	}
}

// CleanupClient removes a client from the weighted slice and channel map.
func (b *Balancer) cleanupClient(client Client) {
	b.mu.Lock()
	defer b.mu.Unlock()

	log.Printf("Cleaning up client %+v\n", client)
	delete(b.clientsChan, client)
	b.weightedSlice.RemoveItemByValue(client)
}

// Graceful shutdown of the balancer.
func (b *Balancer) Shutdown() {
	log.Println("Initiating shutdown...")
	b.cancel()
	b.wg.Wait()
	log.Println("Shutdown complete")
}
