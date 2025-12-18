# 🤖 Robot Secret Reconstruction Simulation

This project is a **distributed systems simulation** where multiple autonomous robots collaborate to reconstruct a shared secret using a **gossip / anti-entropy protocol**.

Each robot initially holds **only a subset of the secret**, split into indexed words. Through unreliable, asynchronous communication, robots progressively exchange missing information until **one robot eventually reconstructs the full secret and writes it exactly once**.

This project is **not** a chat, **not** an HTTP API, and **not** a CRUD service.
It is a **laboratory for reasoning about distributed systems behavior, failure modes, and observability**.

---

## 🎯 What problem does this project simulate?

The simulation models a **distributed knowledge convergence problem** under realistic conditions:

* No central coordinator
* No guaranteed message delivery
* No ordering guarantees
* No synchronous communication
* No shared memory

Despite these constraints, the system must **eventually converge** toward a correct and complete state.

The project demonstrates how **eventual consistency** can be achieved in the presence of:

* message loss
* message duplication
* message reordering
* partial failures
* concurrent execution

---

## 🧠 Core distributed systems concepts illustrated

This project is intentionally designed to surface *fundamental distributed systems principles*:

### Gossip & Anti-Entropy

* Robots periodically exchange *summaries* of their local state.
* Missing information is requested and propagated incrementally.
* No robot ever sends the full state unless necessary.

### Invariants as Consistency Boundaries

The system enforces strong local invariants:

* **Monotonicity**: secret parts are never removed
* **Uniqueness**: a given index maps to exactly one word
* **Idempotence**: duplicate messages have no effect

Invariant violations are treated as **fatal errors** and intentionally trigger panics to test supervision behavior.

### Eventual Convergence

* No robot knows when the system is "done" globally.
* Completion is detected locally using:

  * completeness checks (no missing indexes)
  * a quiet period (quiescence detection)

---

## 🧵 Concurrency & Execution Model

Each robot is composed of **independent workers**, each with a single responsibility:

* gossip initiation
* summary processing
* secret merging
* quiescence detection
* convergence detection

Workers communicate **exclusively through channels** and are:

* asynchronous
* restartable
* isolated from each other

This mirrors actor-like systems and highlights the cost of coordination.

---

## 🔁 Supervision & Fault Tolerance

All workers are supervised:

* panics are recovered
* failed workers are automatically restarted
* failures are isolated and do not crash the system

This supervision model is **inspired by Erlang/OTP**, adapted to Go:

* workers are intentionally simple
* robustness is achieved through supervision, not defensive coding

---

## 📊 Events, Metrics & Observability

The system emits structured events to describe its internal behavior:

* messages sent / received
* message loss, duplication, reordering
* invariant violations
* worker restarts
* channel capacity pressure
* quiescence signals

Events are processed through a **chain-of-responsibility pipeline**, enabling:

* metrics aggregation
* logging
* anomaly detection

Observability is a *first-class concern*, not an afterthought.

---

## 🧪 Testing Philosophy

The test suite focuses on **behavioral guarantees**, not implementation details:

* eventual convergence under loss and duplication
* idempotence and invariant enforcement
* exactly-once secret writing
* no false positives before quiescence
* correct behavior under timeouts

Many tests rely on **Eventually-style assertions**, which are essential when validating asynchronous, distributed behavior.

---

## 🧩 What this project is *really* about

Beyond the surface simulation, this project explores:

* how to structure concurrent systems
* where to enforce correctness
* how failures should be handled, not hidden
* how to *observe* a live distributed system

It is designed as a **thinking tool** as much as a codebase.

---

## 🚧 What’s next

The next step is **not adding more logic**, but improving **human visibility**:

* introducing a lightweight UI (likely a TUI)
* visualizing convergence, instability, and supervision behavior
* turning logs and metrics into an intuitive, live system view

The goal is to make the system **understandable at runtime**, without changing its behavior.

---

If you are interested in distributed systems, failure handling, or observability-driven design, this repository is meant to be read, explored, and experimented with — not just executed.
