# 🔴 My-Redis – A Redis Clone in Go

A fully functional Redis clone written in Go, supporting core Redis commands, RESP protocol, RDB persistence, and master-slave replication. This project was built from scratch to deeply understand how Redis works under the hood.

---

## 🚀 Features

- ⚡ In-memory key-value store with TTL support
- 💬 RESP (REdis Serialization Protocol) parser and encoder
- 💾 RDB persistence with `SAVE` and `BGSAVE`
- 🔄 Master-slave replication using `REPLICAOF`, `PSYNC`, `REPLCONF`, and `WAIT`
- 🧠 Core Redis commands: `PING`, `ECHO`, `SET`, `GET`, `DEL`, `KEYS`, `INFO`, etc.
- 🔌 Custom TCP server using Go's `net` package
- 🔒 Role-based connection handling (client/master/slave)
- 🧪 Clean modular codebase with extensibility in mind

---

## 📁 Directory Structure

```
My-Redis/
├── main.go              # Entry point of the Redis server
├── resp/                # RESP protocol parser and encoder
├── rdb/                 # RDB file handling (load/save/stream)
├── store/               # In-memory key-value store with TTL
├── testdata/            # Sample data directory (RDB files)
```

---

## 🛠️ Getting Started

### Prerequisites

- Go 1.20+
- (Optional) `redis-cli` for interaction

### Build

```bash
git clone https://github.com/varunarora1606/My-Redis.git
cd My-Redis
go build -o myredis
```

### Run

```bash
./myredis --dir=./testdata --filename=dump.rdb --port=8000
```

---

## ⚙️ Usage

You can use `redis-cli` or `telnet` to interact with the server:

```bash
redis-cli -p 8000
```

### Basic Commands

```bash
SET name varun
GET name
DEL name
KEYS *
SAVE
BGSAVE
INFO replication
```

---

## 🔁 Replication

### Make a node a replica

```bash
REPLICAOF 127.0.0.1 8000
```

### Stop replication (become master)

```bash
REPLICAOF NO ONE
```

### Consistency

Ensure `WAIT <replicas> <timeout-ms>` is used to wait for ACKs from slaves:

```bash
WAIT 1 1000
```

---

## 💾 Persistence

- `SAVE`: Synchronously saves the in-memory store to an RDB file
- `BGSAVE`: Asynchronously saves in the background
- Data is auto-loaded from the RDB file on server start

---

## 🧠 Supported Commands

| Command     | Description                             |
|-------------|-----------------------------------------|
| `PING`      | Health check                            |
| `ECHO msg`  | Echo the message                        |
| `SET key val [PX ms]` | Set key with optional expiry  |
| `GET key`   | Get value by key                        |
| `DEL key`   | Delete a key                            |
| `KEYS *`    | Get all keys                            |
| `SAVE`      | Save to RDB synchronously               |
| `BGSAVE`    | Save to RDB asynchronously              |
| `INFO replication` | Get replication details          |
| `REPLICAOF host port` | Become replica of a master    |
| `WAIT n t`  | Wait for `n` replicas or `t` ms         |
| `PSYNC`     | Used during replication handshake       |
| `REPLCONF`  | Used during replication handshake       |

---

## 🔧 Internals

- Custom TCP connection handling
- RESP parsing with proper protocol handling
- Replication handshake protocol (`PING`, `REPLCONF`, `PSYNC`)
- RDB file streaming during full sync
- Offsets and replication IDs for syncing state
- Periodic ACKs to maintain replication health

---

## 🧪 Example

Run master on port `8000`:

```bash
./myredis --port=8000
```

Run slave on port `8001`:

```bash
./myredis --port=8001
```

Connect to the slave:

```bash
redis-cli -p 8001
REPLICAOF 127.0.0.1 8000
```

Set key from master:

```bash
redis-cli -p 8000
SET hello world
```

Get key from slave:

```bash
redis-cli -p 8001
GET hello
# Output: "world"
```

---

## 📈 Roadmap

- [ ] Add AOF persistence
- [ ] Support more data types (`list`, `set`, etc.)
- [ ] Add LRU eviction policy
- [ ] Add Pub/Sub functionality
- [ ] Redis Cluster support

---

## 👨‍💻 Author

**Varun Arora**  
🔗 [@VarunArora80243](https://x.com/VarunArora80243)

---

## 📝 License

MIT License. See [LICENSE](./LICENSE) for details.

---

## 💡 Acknowledgements

- Inspired by [Redis](https://redis.io/)
- RESP protocol: [https://redis.io/docs/reference/protocol-spec/](https://redis.io/docs/reference/protocol-spec/)
- Built as part of a Redis deep dive learning project

---
