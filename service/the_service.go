package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"gitlab.com/kiwicom/search-team/balancer/work"
)

// TheExpensiveFragileService is a service that we need to utilise as much as we can, since it's expensive to run,
// but on the other hand is very fragile, so we can't just flood it with thousands of requests per second.
// It implements balancer.Server interface.
type TheExpensiveFragileService struct{}

// Process a single request and return an error if occurred.
func (TheExpensiveFragileService) Process(_ context.Context, request *work.Request) error {
	if rand.Intn(1000) < 2 {
		return fmt.Errorf("request %d failed", request.ID)
	}
	// do not implement me, just imagine there's huge, complex, almost extra-terrestrial logic here which takes
	// arbitrary number of time to complete.
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
	return nil
}
