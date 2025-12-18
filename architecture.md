# Architecture Overview

## Supervisor & Event Flow

```text
+--------------------+        +------------------+
|    Supervisor      |        |   Event Channel  |
| (manages workers)  |<------>|  events.Event    |
+--------------------+        +------------------+
       |                                  ^
       | starts workers                   |
       v                                  |
+---------------------+   observes    +------------------+
| MergeSecretWorker   |-------------> |      Robot       |
+---------------------+               +------------------+
| ProcessSummaryWorker|
| StartGossipWorker   |
| MetricWorker        |
| SnapshotWorker      |
+---------------------+



This design ensures:

1. **No cyclic dependencies**: Robot is independent of events/workers.  
2. **Event-driven architecture**: All metrics and updates flow through a central channel.  
3. **Resilient and modular**: Supervisor restarts workers automatically. New collectors can be added easily.  

---

If you want, I can also produce a **more detailed "Datadog-style agent architecture"** diagram including **OS metrics collection, custom plugins, and transport layer** in Markdown. It would be closer to a production-ready agent design.  

Do you want me to do that?
