package work

// Request represents work that needs to be done.
type Request struct {
	// ID identifies the work that needs to be done to the server.
	// It is the only field here for the sake of simplicity.
	// The same ID can be requested by different clients.
	ID int
}
