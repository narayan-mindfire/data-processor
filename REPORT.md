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

## 4. Trade-offs
1. **Memory Buffering vs Disk Streaming:** Currently, the pipeline passes `map[string]any` records through channels in memory. While blazing fast, a massive 50GB CSV file could trigger an Out-Of-Memory (OOM) killer. A future trade-off would involve writing intermediate states to disk or a Kafka queue.
2. **Channel Sizes:** The channels (`recordsCh`, `validatedCh`) are currently unbuffered (size 0). This creates strong back-pressure (workers don't ingest faster than they can process), but it trades off peak burst capacity.
3. **Database Throttling:** Saving every single error individually could overwhelm PostgreSQL during a massive data failure. The system mitigates this by batch-updating the Job Progress every 2 seconds, but individual error inserts could eventually be upgraded to batch inserts.