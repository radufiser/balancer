package balancer

import (
	"context"

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
	// implement me
}

// New creates a new Balancer instance.
// The balancer needs the server where it sends requests to and a maximum number of parallel requests that should be
// in flight at any given time.
// to the server.
// THIS IS A HARD REQUIREMENT - THE SERVICE CANNOT PROCESS MORE THAN maxParallelRequests IN PARALLEL.
func New(server Server, maxParallelRequests int32) *Balancer {
	panic("implement me")
}

// Register a client to the balancer and start dispatching its requests to the server.
// Assume that one client can register itself multiple times.
func (b *Balancer) Register(ctx context.Context, client Client) {
	panic("implement me")
}
