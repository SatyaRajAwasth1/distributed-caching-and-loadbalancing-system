# distributed-caching-and-loadbalancing-system
A basic implementation and visualization of caching and load balancing system for distributed platform. It's a college project that implements basic implementation of data structures like LinkedLists, Hash Maps and Pointers.

## Project Structure
```
distributed-caching-and-loadbalancing-system/
|-- caching/
|   |-- cache/
|   |   |-- cache.go          // Cache implementation
|   |   |-- cacher.go         // Cache interface
|   |   |-- command.go        // Command processing logic
|   |   |-- persist.go        // AOF persistence logic
|   |-- replication.go       // Replication logic
|-- server/
|   |-- master.go            // Master server implementation
|   |-- slave.go             // Slave server implementation
|-- tmp/                     // Temporary files
|-- config.yml               // Configuration file
|-- main.go                  // Main application logic

```