# 🚀 Robot Secret Reconstruction Simulation

## 📘 Description

This project simulates the collaboration of **6 to N robots** to reconstruct a secret distributed across multiple words.  
Each robot receives **a random subset of the secret**, and no robot knows the full phrase initially.

Robots communicate using a single method:  
➡️ **message exchanges via Go channels**.

The goal is for **one robot to reconstruct the entire secret**, then write it to a final output file once all words have been received and at least one word contains the end-of-secret character (`END_OF_SECRET`).

The simulation includes realistic conditions:

- Random message loss and duplication
- Concurrent execution via goroutines
- Progressive propagation of words between robots
- Proper shutdown once a robot reconstructs the secret

---

## ✨ Features

- Random distribution of secret words among robots
- Asynchronous communication using goroutines and buffered channels
- Configurable handling of lost and duplicated messages
- Secret reconstruction once all words have been collected
- Writing the reconstructed secret to a single output file (`OUTPUT_FILE`)
- Quiet period (`QUIET_PERIOD`) to ensure all messages have propagated
- Maximum attempt limit (`MAX_ATTEMPTS`) for message exchanges
- Structured logging configurable via `LOG_LEVEL` (`DEBUG`, `INFO`, `WARN`, `ERROR`)
- Proper shutdown of robot goroutines after reconstruction

---

## 🧠 Assumptions

1. No robot permanently fails: all goroutines remain active during the simulation.
2. Messages may be lost or duplicated but are not permanently blocked.
3. Secret reconstruction is **independent of word order**: it succeeds once all expected words are present.
4. The `END_OF_SECRET` character determines the logical end of the secret.
5. The output file (`OUTPUT_FILE`) is the single canonical destination.
6. All configuration is provided via **environment variables** or the Makefile.

---

## 🔧 Environment Variables

| Variable                     | Description |
|------------------------------|-------------|
| `SECRET`                     | The secret phrase to reconstruct. |
| `NBR_OF_ROBOTS`              | Number of robots in the simulation. |
| `BUFFER_SIZE`                | Buffer size for each robot’s channel. |
| `END_OF_SECRET`              | Character marking the end of the secret (e.g., `"."`). |
| `OUTPUT_FILE`                | File where the reconstructed secret will be written. |
| `PERCENTAGE_OF_LOST`         | Percentage of messages randomly lost. |
| `PERCENTAGE_OF_DUPLICATED`   | Percentage of messages randomly duplicated. |
| `DUPLICATED_NUMBER`          | Number of times a message is duplicated if duplication occurs. |
| `MAX_ATTEMPTS`               | Maximum number of attempts for sending messages between robots. |
| `TIMEOUT`                    | Overall timeout for the simulation (e.g., `10s`). |
| `QUIET_PERIOD`               | Time to wait after the last word is received before declaring the secret complete. |
| `LOG_LEVEL`                  | Log level (`DEBUG`, `INFO`, `WARN`, `ERROR`). |

---

## ▶️ Running via Makefile (recommended)

The Makefile handles compilation, execution, and environment variable setup:

```bash
make run
