# Architecture & Concurrency Report

## 1. Concurrency Model: Fan-Out / Fan-In
This data processing pipeline employs a dynamic Fan-Out/Fan-In concurrency architecture utilizing Go channels and `sync.WaitGroup` to maximize throughput across I/O and CPU boundaries.

- **Ingestion (Fan-Out):** HTTP boundaries are highly latent. For every `SourceDef` requested by the client, a dedicated Goroutine is spawned to download and parse the data concurrently. This ensures a slow CSV download does not block a fast JSON API download.
- **Processing (Worker Pools):** The pipeline spins up dynamically configurable worker pools (e.g., `validation_workers: 5`) that actively listen to the `recordsCh`. This allows CPU-intensive tasks (like regex matching or type-casting) to be distributed across CPU cores, preventing a single slow record from bottlenecking the pipeline.
- **Aggregation (Fan-In):** To ensure mathematical accuracy (preventing race conditions on sums and averages), all validated and transformed records are funneled into a single `aggregate` goroutine. This guarantees deterministic mathematical output without relying on expensive Mutex locks.

## 2. Observability & Thread Safety
A core challenge of high-concurrency pipelines is extracting metrics (Progress % and Error Counts) without introducing Data Races or locking up the workers. 
This was solved using **Side-Channels**. The worker pools push metrics into a `progressCh` and `errorCh`. A dedicated, isolated `observabilityTracker` goroutine uses a non-blocking `select` loop to continuously drain these channels, calculate the totals, and periodically flush them to the PostgreSQL database every 2 seconds via a `time.Ticker`.

## 3. Resilience and Context Cancellation
The engine uses Go's `context.Context` to manage graceful shutdowns. If a client triggers the `PATCH /cancel` endpoint, the orchestrator cancels the context. Every single worker pool and ingestion goroutine actively listens to `ctx.Done()` and immediately aborts its work, preventing CPU/Memory leaks.

## 4. Data Export & Streaming
To avoid memory bottlenecks when returning enormous datasets, the pipeline implements an independent **Export Worker**. It listens to an `exportCh` and immediately persists processed records to a PostgreSQL `JSONB` column. The REST API exposes `GET /export/json` and `GET /export/csv` streaming endpoints that pipe database rows directly into the HTTP response stream chunk-by-chunk, bypassing memory buffers entirely.

## 5. Real-Time Metrics & Latency Profiling
To monitor pipeline health, we implemented lock-free performance counters. Using Go's `sync/atomic` package, the engine dynamically tracks microsecond-level processing times (`time.Since`) for every single record across ingestion, validation, transformation, and export. The API returns real-time processing throughput (`records_per_second`) dynamically, and upon completion, surfaces average `stage_latencies`.

## 6. Trade-offs
1. **Memory Buffering vs Disk Streaming:** Currently, the pipeline passes `map[string]any` records through channels in memory. While blazing fast, a massive 50GB CSV file could trigger an Out-Of-Memory (OOM) killer. A future trade-off would involve writing intermediate states to disk or a Kafka queue.
2. **Channel Sizes:** The channels (`recordsCh`, `validatedCh`) are currently buffered at size 100. This provides a minor buffer to absorb bursts, but still applies strong back-pressure.
3. **Database Throttling:** Saving every single record to PostgreSQL individually during the `exportWorker` could overwhelm the DB during a massive data pipeline. The system could be upgraded to batch-insert exported records.