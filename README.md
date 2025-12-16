# 🤖 Robot Secret Reconstruction Simulation

This project is a **distributed systems simulation** where multiple autonomous robots collaborate to reconstruct a shared secret using a **gossip / anti-entropy protocol**.

Each robot initially holds **only a subset of the secret**, split into indexed words.  
Through unreliable, asynchronous communication, robots progressively exchange missing information until **one robot eventually reconstructs the full secret** and writes it to disk.

This project is **not** a chat, **not** an HTTP API, and **not** a CRUD service.  
It is a **laboratory for reasoning about distributed systems behavior**.

---

## 🎯 What problem does this project simulate?

This simulation models a **distributed knowledge convergence problem** under real-world conditions:

* No central coordinator
* No guaranteed message delivery
* No ordering guarantees
* No synchronous communication
* No shared memory

It demonstrates how **eventual consistency** can be achieved despite:

* message loss
* message duplication
* random delays
* concurrent execution
* partial knowledge

This pattern is commonly found in:

* gossip protocols
* monitoring agents (e.g. Datadog-like agents)
* peer-to-peer systems
* distributed caches
* CRDT-based systems

---

## 🧩 High-Level Architecture

Each robot is an **independent concurrent system** composed of multiple workers.
Robots **never share memory** and only communicate through **asynchronous message passing**.

### Robot internal workers

Each robot runs the following goroutines:

#### 1. Gossip Sender Worker
* Periodically selects another robot at random
* Sends a `GossipSummary` containing:
   * the indexes it already knows
   * its own robot ID
* Messages may be:
   * dropped
   * duplicated
   * rejected due to buffer saturation

#### 2. Summary Processor (Anti-Entropy Worker)
* Receives a `GossipSummary`
* Computes which secret parts the sender is missing
* Replies with a `GossipUpdate`
* Implements **anti-entropy reconciliation**

#### 3. Update Processor Worker
* Receives `GossipUpdate` messages
* Merges secret parts into local state
* Enforces **strong invariants** via a single merge function

#### 4. Supervisor Worker
* Observes robot state evolution
* Detects:
   * full secret reconstruction
   * quiescence (no recent updates)
* Handles panics from invariant violations
* Restarts failed workers when needed

All workers:
* run concurrently
* communicate via buffered channels
* can be stopped via context cancellation

There is **no global coordinator**.

---

## 🧠 Core Design Principles

### 1. Monotonic State Growth (Strong Invariant)

A robot **never forgets** a word once learned.

* Secret parts are immutable
* No deletion
* No overwrite
* No rollback

This guarantees **eventual convergence**.

---

### 2. Idempotent Message Processing

Messages can be:

* duplicated
* received out of order
* replayed

Robots explicitly ignore already-known secret parts **when they are identical**.

Result:

* duplication is harmless
* retries are safe
* the system is resilient by design

---

### 3. Order Independence

Internally, secret parts are stored **unordered**.  
Ordering is applied **only at read time**, based on word indexes.

This cleanly separates:

* internal state representation
* final deterministic output

The final secret is **always reconstructed in the correct order**, regardless of message arrival order.

---

### 4. No Locks, No Global Synchronization

Despite heavy concurrency:

* no mutexes
* no global locks
* no shared writable state

Correctness is ensured **by invariants**, not by synchronization.

This mirrors real-world distributed systems where locks do not scale.

---

## 🔒 Invariant Enforcement & Consistency Boundary

All mutations of a robot’s local state go through a **single consistency boundary**:

**MergeSecretPart()** function guarantees:

* monotonic state growth
* index → word immutability
* idempotent updates

If conflicting data is received for the same index, the program **panics deliberately**.

Such panics are:
* caught by the supervisor
* logged
* followed by a worker restart

This models **fail-fast correctness** inside an unreliable environment.

---

## 📡 Gossip & Anti-Entropy Protocol

The protocol follows a classic anti-entropy pattern:

1. A robot selects a random peer
2. Sends a summary of known indexes
3. The peer computes missing parts
4. Sends back updates
5. The sender merges new information

This exchange is:

* asynchronous
* unordered
* lossy
* idempotent

Yet it converges.

---

## ⚠️ Failure Modes Simulated

| Failure Type        | Description                         |
|--------------------|-------------------------------------|
| Message loss        | Messages may be randomly dropped    |
| Message duplication | Messages may be sent multiple times |
| Buffer saturation   | Channels may reject messages        |
| Unordered delivery  | No ordering guarantees              |
| Concurrency         | Independent goroutines per robot    |

---

## 🕰️ Completion & Stability Detection

A robot is considered a winner when:

* it has reconstructed the **full secret**
* no updates were received for a configurable **quiet period**

Time is used **only as a stability heuristic**, not as a correctness mechanism.

---

## 🏆 Winner Selection

Multiple robots may reach completion.

To avoid race conditions:

* a non-blocking winner channel is used
* only the first robot succeeds
* others stop gracefully

The winner writes the reconstructed secret to disk.

---

## 🚫 Explicitly Out of Scope

* Byzantine behavior
* Permanent partitions
* Security / authentication
* Persistence across restarts
* Clock synchronization guarantees

---

## 🔮 Possible Extensions

* Logical clocks (Lamport / vector)
* Message IDs and explicit deduplication
* Failure suspicion
* Metrics & observability
* CRDT-based state
* Alternative payloads
* Other transports (QUIC, BLE, WebRTC)

---

## 📦 Protobuf Code Generation

Proto files are located in `/proto`.

Generate Go code from the project root:
```bash
docker build -t protoc-image .
```

```bash
docker run --rm -v "$PWD:/defs" protoc-image \
  -I . \
  --go_out=paths=source_relative:. \
  --go-grpc_out=paths=source_relative:. \
  proto/robot.proto