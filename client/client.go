package client

import (
	"context"

	"gitlab.com/kiwicom/search-team/balancer/work"
)

// Client implements balancer.Client interface.
type Client struct {
	requestCount int
	weight       int
}

// New creates a new client that will send requestCount requests to the server.
// weight is the relative weight of this client compared to other clients.
func New(requestCount, weight int) *Client {
	return &Client{
		requestCount: requestCount,
		weight:       weight,
	}
}

// Weight of this client, i.e. how "important" they are.
// This might affect how much computation capacity the client should be allocated by Balancer.
func (c *Client) Weight() int {
	return c.weight
}

// Workload feeds requests that need to be processed by this client through the returned channel
// until there is no more work to be fed or ctx was Done.
func (c *Client) Workload(ctx context.Context) chan *work.Request {
	workload := make(chan *work.Request)
	go func() {
		defer close(workload)

		remaining := c.requestCount
		for remaining > 0 {
			select {
			case <-ctx.Done():
				return
			case workload <- &work.Request{ID: remaining}:
				remaining -= 1
			}
		}
	}()
	return workload
}
