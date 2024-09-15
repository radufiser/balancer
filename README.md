# Task description

Imagine a standard client-server relationship, only in our case, the server is very fragile (and very expensive).
We can't let an arbitrary number of clients smashing the server with arbitrary number of requests,
but as it's so expensive to run, we'd like to utilize it as much as possible. And that's why we need a balancer.

A balancer is a gateway between the clients and the server. A client registers themselves in balancer and the balancer
is responsible for distributing the service capacity between registered clients.

The balancer must ensure that at any given time, the number of parallel requests sent to the server does not
exceed the provided limit.

Based on your skill, you can choose to implement a very simple balancer or production-ready balancer.

The implementation of a simple balancer can, for example, be one of:

* A balancer that serves one client at a time, enqueuing others.
* A balancer that serves registered clients in a round-robin fashion.
* A balancer that is processing batches of registered clients and distributing the capacity among them, while enqueuing
  incoming clients and once one batch is done, process the next batch.

The ultimate, production balancer would run the requests from multiple clients in parallel according to their weights
and process the work fairly from all the clients.
A client with twice the weight of another client would be allowed to run twice more parallel requests at any given time.

In any case, **the balancer must ensure that the number of requests in process at any given time equals the provided
limit of the server**, aka we never can over-utilise the server, but we also don't want to under-utilise it.


# Balancer Implementation Description

## Implemented Components
### Balancer
- **Description**: The core component responsible for managing and balancing client requests to the server.
- **Properties**:
  - `maxParallelRequests`: The maximum number of concurrent requests allowed.
  - `server`: The server that processes the requests.
  - `weightedSlice`: A collection of clients where each client's capacity is determined by its weight.
  - `workQueue`: A buffered channel for storing requests waiting to be processed.
  - `clientsChan`: A map storing each client and its corresponding workload channel.
  - `mu`: A mutex to ensure safe concurrent access to shared resources.
  - `wg`: A `sync.WaitGroup` to coordinate and wait for the worker goroutines.
  - `cancel`: A cancel function used to gracefully shut down the balancer.

### WeightedSlice
- **Description**: A specialized data structure used to store and manage clients, taking into account their weights for request processing.
- **Properties**:
  - `elements`: A slice that holds weighted clients (or other types of elements).
  - `indexMap`: A map that tracks the positions of the elements in the slice, allowing for efficient removal operations.
  - `mu`: A mutex that ensures thread-safe operations when modifying the `WeightedSlice`.
  
- **Key Methods**:
  - **AddItem(item)**: Adds an item to the `WeightedSlice` according to its weight, ensuring the item appears multiple times based on its weight.
  - **GetRandomItem()**: Returns a random item from the slice, with higher-weight items having a higher probability of being selected.
  - **RemoveItemByValue(value)**: Removes all instances of an item with the specified value from the `WeightedSlice`, using the `indexMap` for efficient removal.
  - **Len()**: Returns the current number of elements in the `WeightedSlice`.

- **Usage in Balancer**:
  - The `Balancer` uses `WeightedSlice` to fairly allocate processing capacity to clients. The `GetRandomItem()` method is used to select clients based on their weight, ensuring clients with higher weights get more requests processed.

## 3. Key Methods

### New(server, maxParallelRequests)
- **Purpose**: Creates and initializes a new instance of the `Balancer`.
- **Process**:
  - Sets up the server, maxParallelRequests, and starts the worker pool.
  - Initializes the `workQueue` and other internal properties.
  - Starts balancing requests by running the `start()` method in a separate goroutine.

### Register(ctx, client)
- **Purpose**: Registers a client with the balancer.
- **Process**:
  - Adds the client to the `weightedSlice` based on its weight.
  - Retrieves the client's workload channel and stores it in `clientsChan`.

### start(ctx)
- **Purpose**: Main loop to balance client requests.
- **Process**:
  - Continuously retrieves a client from `weightedSlice` based on its weight.
  - Reads requests from the client's channel and places them into the `workQueue`.
  - If a client's workload channel closes, the client is removed from the `weightedSlice`.

### worker(i, ctx)
- **Purpose**: Handles processing tasks from the `workQueue`.
- **Process**:
  - Each worker reads tasks from the `workQueue` and sends them to the server's `Process()` method.
  - The worker checks for context cancellation and shuts down gracefully when needed.

### workerPool(numWorkers, ctx)
- **Purpose**: Initializes a pool of worker goroutines.
- **Process**:
  - Spawns a fixed number of workers, each responsible for processing tasks in parallel.


## 4. Flow

### Request Distribution
1. **Client Registration**: Clients register with the balancer, providing a workload channel and weight.
2. **Weighted Balancing**: The `Balancer` selects clients from the `weightedSlice` based on their weight, ensuring that clients with higher weight receive more requests to process.
3. **Task Dispatch**: Requests from the clients are placed in the `workQueue`.
4. **Worker Processing**: Worker goroutines consume requests from the `workQueue` and send them to the server for processing.

