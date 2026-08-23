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

## Build & Run

```bash
go build -o cachedb ./app
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

## Educational Value

This project demonstrates:

* Wire protocol implementation — RESP parsing and encoding from scratch
* Distributed systems fundamentals — replication, consistency, offset tracking
* Data structure design — skiplist with span tracking, generic sets
* Concurrency in Go — goroutines, mutexes, atomics, channels
* Persistence patterns — AOF replay and RDB snapshot loading
* Access control design — bitmap-based permission systems

---