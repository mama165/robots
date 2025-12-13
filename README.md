# 🤖 Robot Secret Reconstruction Simulation

This project is a **distributed systems simulation** where multiple autonomous robots collaborate to reconstruct a shared secret using a **gossip / anti-entropy protocol**.

Each robot initially holds **only a subset of the secret**, split into indexed words.
Through unreliable, asynchronous communication, robots progressively exchange missing information until **one robot eventually reconstructs the full secret** and writes it to disk.

This project is not a chat, not an HTTP API, and not a CRUD service.
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

This kind of pattern is commonly found in:

* gossip protocols
* monitoring agents (e.g. Datadog-like agents)
* peer-to-peer systems
* distributed caches
* CRDT-based systems

---

## 🧠 Core Design Principles

### 1. Monotonic State Growth (Strong Invariant)

A robot **never forgets** a word once learned.

* Secret parts are immutable
* No deletion
* No overwrite
* No rollback

This guarantees **convergence**.

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

## 🔒 Invariant Enforcement

Robots explicitly enforce core consistency invariants at runtime.

All mutations of a robot’s local state go through a single merge operation which guarantees:

* Monotonic state growth (secret parts are never removed)
* Index → word immutability
* Idempotent message handling

Receiving conflicting data for the same index is considered a **programming error** and triggers a panic.
Such panics are caught by the supervisor, logged, and cause the affected worker to be restarted.

This ensures the system **tolerates external unreliability** while **refusing internal corruption**.

---

## 📡 Gossip & Anti-Entropy Mechanism

The communication protocol is intentionally simple:

1. A robot periodically selects another robot at random
2. It sends a **GossipSummary** containing:

   * the indexes it already knows
   * its own sender ID
3. The receiver computes the **missing parts**
4. It replies with a **GossipUpdate**
5. The sender merges new information (if any)

This is a classic **anti-entropy exchange**.

---

## ⚠️ Failure Modes Simulated

The system explicitly simulates real-world failures:

| Failure Type        | Description                         |
| ------------------- | ----------------------------------- |
| Message loss        | Messages may be randomly dropped    |
| Message duplication | Messages may be sent multiple times |
| Buffer saturation   | Channels may reject messages        |
| Unordered delivery  | No delivery ordering guarantees     |
| Concurrency         | All robots run independently        |

Despite this, the system still converges.

---

## 🕰️ Completion Detection Strategy

A robot is considered a winner when:

* It has reconstructed the **full secret**
* No new updates have been received for a configurable **quiet period**

This avoids false positives and models **eventual quiescence** instead of instant completion.

---

## 🏆 Winner Selection

Multiple robots *may* reach a complete state.

To avoid race conditions:

* a non-blocking winner channel is used
* only the first robot successfully publishes itself
* others gracefully stop

The winner writes the reconstructed secret to a file.

---

## 🚫 What this project deliberately does NOT solve

* Byzantine behavior
* Network partitions with permanent isolation
* Security / authentication
* Persistent storage across restarts
* Clock synchronization (physical time)

These are intentionally excluded to keep the focus sharp.

---

## 🔮 Possible Extensions

This project is designed to be extended:

* Lamport logical clocks
* Message IDs and explicit deduplication
* Heartbeat & failure suspicion
* Metrics & observability
* CRDT-based state representation
* Alternative payloads (chat messages, media chunks)
* Different transport layers (WebRTC, QUIC, BLE)

The core engine remains the same.

---

## 📦 Protobuf Code Generation

The `.proto` files are located in the `/proto` directory.

To generate Go code on **Windows, Linux or macOS**, run the following command
from the **project root**:

> **IMPORTANT:** The command must be executed from the project root.

```bash
docker run --rm -v "${PWD}/proto:/defs" namely/protoc-go ls /defs/proto
```
