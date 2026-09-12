# Project Context: PakaiWA Platform

> **Status**: CONFIRMED & GROUNDED  
> **Classification**: Shared Foundation Framework & Architecture Baseline

---

## 1. Project Purpose
`pakaiwa-platform` adalah fondasi pustaka inti (*shared core library*) untuk platform **PakaiWA** (ekosistem gateway WhatsApp, broadcast messaging, device session management, dan webhook processing). Repositori ini menyediakan paket-paket dasar yang siap pakai untuk produksi, meminimalisir duplikasi kode antarmicroservice, serta menegakkan standar keandalan tinggi (resilience), keamanan, dan observabilitas.

---

## 2. System Boundary
- **Dalam Lingkup Repositori (`pakaiwa-platform`)**:
  - Abstraksi klien database (PostgreSQL connection pool via `pgxpool`).
  - Abstraksi klien cache (Redis via `go-redis/v9`).
  - Abstraksi publisher/consumer event (Kafka berbasis Confluent librdkafka dan Webhook HTTP).
  - Web engine bootstrap factory berbasis Fiber v3.
  - Telemetri terpusat: Prometheus metrics handler, Logrus Ordered JSON Formatter terstruktur, domain-scoped loggers (`app`, `db`, `http`, `kafka`, `wa`), dan propagasi Trace ID.
  - Utilitas penunjang: validasi struct (`validator/v10`), hashing password (`bcrypt`), fail-fast bootstrap helper (`errors.Must`), serta sinyal graceful shutdown.
- **Di Luar Lingkup (Dikelola oleh Downstream Services / Repo Lain)**:
  - Logika bisnis protokol WhatsApp (misal Whatsmeow integration, multi-device encryption).
  - Business domain endpoints, otorisasi token JWT, skema database aplikasi, dan migrations.
  - Orchestration pipeline worker dan consumer Kafka handler konkret.

---

## 3. Main Actors & Consumers
- **PakaiWA Backend Services**:
  - **API Gateway Service**: Mengonsumsi `http/server/fiber`, `observability`, dan `validation`.
  - **WhatsApp Worker / Device Engine**: Mengonsumsi `messaging/kafka`, `db/postgres`, `cache/redis`, dan `observability/logging/ctxmeta` (khususnya `LoggerWA`).
  - **Webhook Dispatcher Service**: Mengonsumsi `messaging/http` dan `messaging/event`.
- **Infrastructure Operators / SRE**: Memantau metrik via Prometheus dan log via Grafana Stack (Loki / Alloy / Tempo).

---

## 4. Important Domain Concepts & Glossary
- **Client JID (`device_id`)**: Identifier unik WhatsApp account/device (Jabber ID) yang menjadi kunci partisi dan header routing (`X-Device-ID` / Kafka header) di seluruh pesan.
- **Domain-Scoped Loggers**: Logger yang dipisah secara eksplisit per domain (`LoggerApp`, `LoggerDB`, `LoggerHTTP`, `LoggerKafka`, `LoggerWA`) untuk kemudahan filter di Grafana Loki.
- **Ordered JSON Logging**: Format serialisasi log dengan urutan kunci terstandarisasi (`time`, `level`, `trace_id`, `msg`, `caller`, data keys, `module`).
- **Delivery Report Poll Loop**: Background loop berbasis goroutine yang mengonsumsi event status penerimaan pesan dari Kafka broker.

---

## 5. External Systems & Infrastructure
- **PostgreSQL 12+** (Pengujian CI menggunakan PostgreSQL 16 Alpine).
- **Redis 7+** (In-memory store).
- **Apache Kafka Cluster** (Event streaming message broker).
- **Grafana Observability Stack**:
  - **Prometheus**: Metrics collection via `/metrics`.
  - **Grafana Loki**: Log aggregation dari stdout Ordered JSON.
  - **Grafana Alloy**: OpenTelemetry / telemetry collector.
  - **Grafana Tempo**: Distributed tracing engine.

---

## 6. Runtime Environment & Constraints
- **Language / Runtime**: Go 1.25.0+ (Tested pada matrix Go 1.25.7 dan 1.26).
- **Dependencies Native**: Membutuhkan toolchain CGO untuk kompilasi `github.com/confluentinc/confluent-kafka-go/v2` (librdkafka).
- **Graceful Shutdown**: Mengantisipasi `SIGINT` dan `SIGTERM` (standar siklus hidup container/Kubernetes pod).
- **Network Boundaries**: Proteksi reverse proxy bawaan (`TrustProxy` di Fiber) dan validasi skema HTTP/HTTPS ketat pada Webhook producer untuk mitigasi SSRF.

---

## 7. Coding Conventions & Standards
- **Testing**: Menargetkan test coverage tinggi (~90%+); race condition detector diaktifkan di CI (`-race`).
- **Code Quality**: Diverifikasi oleh `golangci-lint`, `govulncheck`, `staticcheck`, dan `gosec`.
- **Pre-commit**: Menggunakan tool `pre-commit` untuk verifikasi formatting dan linting lokal sebelum push.
- **Release Automation**: Tagging otomatis format `v0.yy.m-dhhmm` berbasis waktu Jakarta (`Asia/Jakarta`).

---

## 8. Known Limitations
- `http/client` saat ini menggunakan singleton global dengan konfigurasi timeout default 5 detik tanpa konfigurasi connection pool kustom per-target URL.
- Error penanganan pada inisialisasi `NewKafkaProducer` saat ini mengembalikan `nil` tanpa error objek terpisah (lihat `TODO.md`).
- Pengujian penuh `db/postgres` pada unit test mock terbatas karena beberapa branch `pgxpool.NewWithConfig` membutuhkan instance PostgreSQL aktif (diuji di integration test CI).
