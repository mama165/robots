# Robot Secret Reconstruction Simulation

## Description

This project simulates 6 to N robots collaborating to reconstruct a secret distributed as individual words. Each word of the secret is randomly assigned to a robot, so no robot initially has the complete secret. Robots communicate **only via messages** (Go channels) to propagate words.

The goal is to reconstruct the complete secret and write it to a single, well-known file as soon as one robot has rebuilt the entire message.

The simulation includes realistic conditions:

- Messages can be randomly lost or duplicated.
- Robots communicate concurrently using goroutines and buffered channels.
- The secret is considered complete once all words are received **and at least one word ends with the specified `END_OF_SECRET` character**.

---

## Features

- Random distribution of secret words among robots.
- Asynchronous robot communication via Go channels.
- Handling of message duplication and loss.
- Secret reconstruction by any robot once all words are received.
- Writing the reconstructed secret to a single output file.
- Safe concurrent operations with goroutines and buffered channels.
- Configurable quiet period to ensure all messages have propagated before completion.
- Configurable maximum attempts for message exchanges.

---

## Assumptions

1. Robots **do not permanently fail** in this version (all goroutines remain active).
2. Messages may be **lost or duplicated**, but a robot eventually receives all required words unless the network is completely blocked.
3. The **secret reconstruction is independent of word order**. Reconstruction succeeds once all words are present and at least one word ends with `END_OF_SECRET`.
4. The output file (`OUTPUT_FILE`) is considered the final and unique destination.
5. Robot goroutines are properly stopped after the secret is written.
6. Configuration values are passed via **environment variables**.

---

## Environment Variables

| Variable                   | Description |
|----------------------------|------------|
| `SECRET`                   | The secret phrase to reconstruct. |
| `NBR_OF_ROBOTS`            | Number of robots to simulate. |
| `BUFFER_SIZE`              | Channel buffer size for each robot. |
| `END_OF_SECRET`            | Character indicating the end of the secret (e.g., "."). |
| `OUTPUT_FILE`              | Output file where the secret will be written. |
| `PERCENTAGE_OF_LOST`       | Percentage of messages randomly lost. |
| `PERCENTAGE_OF_DUPLICATED` | Percentage of messages randomly duplicated. |
| `DUPLICATED_NUMBER`        | Number of times a message is duplicated if duplication occurs. |
| `MAX_ATTEMPTS`             | Maximum number of attempts for sending messages between robots. |
| `TIMEOUT`                  | Overall timeout for the simulation (e.g., `10s`). |
| `QUIET_PERIOD`             | Time to wait after the last word received before declaring the secret complete (e.g., `5s`). |

---

## Example Usage

```bash
export SECRET="Hidden beneath the old oak tree, golden coins patiently await discovery."
export NBR_OF_ROBOTS=10
export BUFFER_SIZE=100
export END_OF_SECRET="."
export OUTPUT_FILE="secret.txt"
export PERCENTAGE_OF_LOST=10
export PERCENTAGE_OF_DUPLICATED=10
export DUPLICATED_NUMBER=14
export MAX_ATTEMPTS=50
export TIMEOUT=10s
export QUIET_PERIOD=5s

go run main.go
