# Radish - My own Redis-like Key-Value Cache

## Overview
Radish is a high-performance in-memory key-value store designed for low-latency caching and fast data retrieval. It supports TCP RESP (Redis Serialization Protocol) with provisions for HTTP too for easy integration with various applications.

## Features
- **Efficient LRU-based caching** with memory-aware eviction.
- **Consistent Hashing** for load distribution across shards.
- **Batch Processing and Write Queue** to improve performance.
- **Multi-threaded Worker Pool** for concurrent write operations.
- **HTTP and TCP RESP Protocol Support** for API interactions.
- **Python SDK** for easy integration with applications.
- **Dockerized Deployment** for easy container-based usage.

## Tech Stack
- **Language:** Golang
- **Networking:** net/http, net/tcp
- **Concurrency:** Goroutines, Channels, sync package
- **Data Storage:** In-memory caching with LRU eviction
- **Containerization:** Docker
- **SDK:** Python

## Installation
### Clone the Repository
```sh
git clone https://github.com/your/repo.git
cd radish
```

### Running via Docker
```sh
docker pull yourdockerhub/radish
docker run -p 7171:7171 yourdockerhub/radish
```

### Running Manually 
This is important and must be done on every host that you want to run Radish on. 
```sh
chmod +x install.sh
./install.sh
./server
```

### Pulling Docker Image from DockerHub
```sh
docker pull arthurw1935/radish
docker run -p 7171:7171 arthurw1935/radish
```

## API Format
### TCP RESP Protocol
Radish follows the RESP protocol for TCP-based communication:

- **PUT Command:**
  ```
  *3\r\n$3PUT$<key_length><key>$<value_length>\r\n<value>
  ```
  Response: `+OK`

- **GET Command:**
  ```
  *2\r\n$3\r\nGET\r\n$<key_length>\r\n<key>
  ```
  
  Response:
  - If key exists: `$<length> <value>`
  - If key does not exist: `$-1\r\n`

## SDK Usage (Python)
Radish provides a Python SDK for easy interaction:

```python
from sdk import CacheClient

client = CacheClient("localhost")
client.put("foo", "bar")
print(client.get("foo"))
```

## Design Goals
- **High Availability:** Designed to handle concurrent requests efficiently.
- **Scalability:** Supports consistent hashing for distributed caching.
- **Performance Optimization:** Uses multi-threaded workers and batch processing.

## System Design Choices
### Cache Management
- Implements **LRU eviction** based on memory constraints.
- Uses **write-through caching** to ensure data consistency.

### Sharding and Hashing
- **Consistent Hashing** distributes keys across multiple shards.
- Virtual nodes help with load balancing.

### Concurrency Model
- **Worker Goroutines** process write tasks in batches.
- Uses **sync.Mutex** and **sync.RWMutex** for thread safety.

### Thought Process
- First, created a basic in-memory key-value store using Golang. Chose Golang for its concurrency model and performance. RPS was around 300-400 requests/sec.
- Implemented LRU caching to manage memory efficiently. But also needed to handle eviction based on memory constraints. (>=70)
- Sharded the cache using consistent hashing to allow for horizontal scaling. This improved performance to around 1200 requests/sec.
- Added a TCP RESP protocol for compatibility with Redis clients. Improved the RPS to around 6000-7000 requests/sec in my machine and 18k req/sec from a separate machine.
- Developed a Python SDK for easy integration with applications.
- Finally, containerized the application using Docker for easy deployment.
