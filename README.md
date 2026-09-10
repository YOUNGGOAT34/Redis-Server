# CacheDB

A **Redis-compatible in-memory database written from scratch in Go** implementing the **RESP protocol over raw TCP**, with persistence, replication, transactions, access control, and multiple data structures.

This project is designed as a **systems-level deep dive** into how Redis works internally — from the wire protocol to replication semantics and data structure design.

---

## Features

* **RESP protocol** parser and encoder implemented from scratch over raw TCP
* **Master-replica replication** — full PING/REPLCONF/PSYNC handshake, RDB file transfer, and offset tracking
* **Persistence** — AOF (Append Only File) and RDB snapshot support
* **Transactions** — `MULTI`/`EXEC` with `WATCH`/dirty flag semantics for optimistic locking
* **ACL system** — 64-bit permission bitmap for per-user command access control
* **Sorted sets** — skiplist + hashmap architecture, matching Redis internals
* **Lists** — doubly linked list with `LPUSH`, `RPUSH`, `LPOP`, `BLPOP`
* **Streams** — `XADD`, `XRANGE`, `XREAD` with binary search on entry IDs
* **Pub/Sub** — channel-based message broadcasting
* **Key expiry** — passive expiry on access
* **WAIT command** — replica acknowledgement polling with configurable timeout

---

## Architecture Overview

* **TCP Listener**

  * Accepts client connections
  * Spawns a goroutine per connection

* **RESP Parser**

  * Incrementally parses incoming bytes
  * Handles pipelining — multiple commands per read

* **Command Dispatcher**

  * Routes parsed commands to handlers
  * Enforces ACL permissions per user per command

* **Replication Layer**

  * Master propagates write commands to replicas after successful execution
  * Replica connects to master, performs handshake, receives RDB snapshot, then streams commands
  * Offset tracking via atomics for `WAIT` acknowledgement

* **Persistence Layer**

  * AOF — write commands appended to file, replayed on startup
  * RDB — binary snapshot loaded into memory on startup

---

## Supported Commands

### Strings
`GET` `SET` `DEL` `INCR`

### Lists
`LPUSH` `RPUSH` `LPOP` `LRANGE` `LLEN` `BLPOP`

### Sorted Sets
`ZADD` `ZREM` `ZRANGE` `ZSCORE` `ZRANK` `ZCARD`

### Streams
`XADD` `XRANGE` `XREAD`

### Pub/Sub
`SUBSCRIBE` `PUBLISH`

### Transactions
`MULTI` `EXEC` `DISCARD` `WATCH` `UNWATCH`

### Access Control
`ACL SETUSER` `ACL GETUSER` `AUTH`

### Replication
`REPLCONF` `PSYNC` `WAIT` `INFO replication`

### Generic
`KEYS` `TYPE` `SAVE` `PING` `ECHO`

---

## Requirements

* Go 1.21 or higher
* Linux or macOS

---
## Testing
### Build all packages:

```bash
go build ./...
```
### Run the full test suite:

```bash
go test ./... -count=1
```

### Run the integration tester suite with verbose output:

```bash
go test ./app/tester -count=1 -v
```

### Run the tester suite with Go's race detector:

```bash
go test -race ./app/tester -count=1
```


The -count=1 flag disables test-result caching, ensuring the tests are executed fresh.

## Build & Run

```bash
make build
```

### Run as standalone server

```bash
./cachedb --port 6379
```

### Run as replica

```bash
./cachedb --port 6380 --replicaof "localhost 6379"
```

### With persistence

```bash
./cachedb --port 6379 \
  --appendonly yes \
  --appendfilename appendonly.aof \
  --dbfilename rdbfile.db
```

---

## Running with Docker

A multi-stage `Dockerfile` builds a small, non-root, statically-linked image
(no CGO, no shell, no OS package manager - `gcr.io/distroless/static-debian12:nonroot`).

```bash
docker build -t cachedb .

# standalone, with a persistent volume mounted at /data:
docker run -d --name cachedb \
  -p 6379:6379 \
  -v cachedb-data:/data \
  cachedb
```

By default the image runs `--dir /data --appendonly yes`, so both the RDB
file and the AOF directory land directly under the mounted volume - data
survives `docker stop`/`docker rm`/container recreation as long as the same
volume is reused. `docker stop` sends SIGTERM, which triggers CacheDB's
existing graceful shutdown (stop accepting new clients, drain in-flight
ones, close the AOF file) rather than a hard kill.

### Master/replica with Docker Compose

```bash
docker compose up --build
```

This starts a master (`cachedb-master`, published on host port 6379) and a
replica (`cachedb-replica`, published on host port 6380) on the same Compose
network, with the replica's `--replicaof` pointing at the master by its
Compose service name (Docker's embedded DNS resolves this automatically).
The replica's own retry/backoff already tolerates the master not being
ready yet; the Compose file also gates the replica's startup on the
master's `HEALTHCHECK` passing, purely to avoid a few seconds of retry
noise on a normal `docker compose up`.

Each service has its own named volume (`master-data`, `replica-data`), so
persisted state for each role survives container recreation independently.

---

## Compatibility

CacheDB speaks RESP — it works with `redis-cli` and any Redis client library out of the box:

```bash
redis-cli -p 6379 SET foo bar
redis-cli -p 6379 GET foo
```

---

## Replication

Start a master and one or more replicas:

```bash
# master
./cachedb --port 6379

# replica
./cachedb --port 6380 --replicaof "localhost 6379"
```

The replica performs the full handshake automatically:
1. Sends `PING` to verify master is alive
2. Sends `REPLCONF listening-port` to register itself
3. Sends `REPLCONF capa psync2` to negotiate capabilities
4. Sends `PSYNC ? -1` to request a full resync
5. Receives an RDB snapshot and loads it into memory
6. Streams all subsequent write commands from the master

---

## Sorted Sets — Design Decision

Sorted sets use a **skiplist for ordered traversal** and a **hashmap for O(1) member lookup** — the same dual-structure architecture Redis uses internally.

The skiplist uses **probabilistic level selection** (p=0.25, max 32 levels) with **span tracking** at each level for efficient rank queries without full traversal.

---

## Transactions

`WATCH` marks keys for optimistic locking. If any watched key is modified before `EXEC`, the transaction is aborted — the client's dirty flag is set and the queued commands are discarded.

---

## ACL System

Permissions are encoded as a **64-bit bitmap** — one bit per command. Users are granted or revoked permissions by OR-ing or AND-ing permission masks:

```text
GET | SET | DEL  →  read/write string access
@READ            →  all read commands
@WRITE           →  all write commands
@ADMIN           →  replication and user management commands
```
---

## Concurrency & Thread Safety

* Per-data-structure mutexes — `ZSMutex`, `ListMutex`, `StreamMutex`
* `UsersMutex` for ACL user table
* `WatchedKeysMutex` for transaction key tracking
* `ReplicasMutex` for replica list
* Atomic offset tracking via `sync/atomic` for lock-free reads

---

## Limitations

* Single-node only — no clustering or consistent hashing
* Streams use a sorted slice with binary search — not a radix tree
* No TLS support
* Expiry is passive only — no active expiration background task

---


## Benchmarks

Benchmarked against Redis 7.x on equivalent hardware.
No pipelining — representative of real application workloads.

`redis-benchmark -c 50 -n 1,000,000 -t <cmd> -q`

| Operation  | CacheDB      | Redis        | Ratio |
|------------|--------------|--------------|-------|
| SET        | 44,630 ops/s | 57,780 ops/s | 77%   |
| GET        | 44,607 ops/s | 56,170 ops/s | 79%   |
| LPUSH      | 44,678 ops/s | 56,439 ops/s | 79%   |
| LRANGE 100 | 20,885 ops/s | 38,729 ops/s | 54%   |
| LRANGE 600 | 5,084 ops/s  | 13,572 ops/s | 37%   |
| ZADD       | 40,576 ops/s | 50,725 ops/s | 80%   |
| INCR       | 41,034 ops/s | 48,787 ops/s | 84%   |
| PING       | 40,928 ops/s | 50,296 ops/s | 81%   |

**Single value operations achieve 77-84% of Redis throughput.**

The LRANGE gap grows with range size due to pointer chasing in
the doubly linked list implementation — each node is a separate
heap allocation causing CPU cache misses at scale. Redis uses
a listpack encoding for small lists — a contiguous memory block
eliminating cache misses entirely. Implementing listpack is a
planned improvement.

## Educational Value

This project demonstrates:

* Wire protocol implementation — RESP parsing and encoding from scratch
* Distributed systems fundamentals — replication, consistency, offset tracking
* Data structure design — skiplist with span tracking, generic sets
* Concurrency in Go — goroutines, mutexes, atomics, channels
* Persistence patterns — AOF replay and RDB snapshot loading
* Access control design — bitmap-based permission systems

---