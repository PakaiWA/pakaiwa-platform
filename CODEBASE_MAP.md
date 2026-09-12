# Codebase Navigation Map

> **Repository Type**: Go Library / Shared Foundation SDK (Polyrepo support for PakaiWA ecosystem)  
> **Go Version**: Go 1.25.0+  
> **Evidence Classification**: All packages documented below are `CONFIRMED` via source code inspection.

---

## `cache/redis`
- **Responsibility**: Abstraksi koneksi Redis client dengan health check ping otomatis pada saat inisialisasi.
- **Entry / Key Files**:
  - [`client.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/cache/redis/client.go) — Inisialisasi struct `Config` dan fungsi `NewRedisClient`.
  - [`client_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/cache/redis/client_test.go) — Unit & integration test Redis client.
- **Dependencies**: `github.com/redis/go-redis/v9`.
- **Consumers**: Consumer/service PakaiWA yang membutuhkan distributed caching atau state store.
- **External Integrations**: Redis Server (v7+).
- **Key Notes**: Field `Password` diberi tag `json:"-"` untuk mencegah kebocoran kredensial secara tidak sengaja via log/JSON serialization.

---

## `db/postgres`
- **Responsibility**: Manajemen database connection pooling PostgreSQL berbasis `pgxpool` dengan validasi config fail-fast, timeout liveness check, dan context-aware logger injection.
- **Entry / Key Files**:
  - [`config.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/db/postgres/config.go) — Struct konfigurasi pool (`DSN`, `MinConns`, `MaxConns`, dsb.).
  - [`pgsql.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/db/postgres/pgsql.go) — Bootstrap `NewDatabase` dengan integrasi `ctxmeta.LoggerDB` dan setting `application_name`.
  - [`pgsql_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/db/postgres/pgsql_test.go) — Test connection pool dan error path.
- **Dependencies**: `github.com/jackc/pgx/v5/pgxpool`, `github.com/sirupsen/logrus`, `observability/logging/ctxmeta`.
- **Consumers**: Service backend microservices yang terhubung ke PostgreSQL (misalnya auth service, device service, messaging store).
- **External Integrations**: PostgreSQL 12+ (CI menggunakan PostgreSQL 16 Alpine).
- **Key Notes**: Menyuntikkan `application_name` ke dalam PostgreSQL runtime params untuk kemudahan tracing di pg_stat_activity.

---

## `errors`
- **Responsibility**: Helper utilitas error handling berbasis idiomatik Go generics untuk fase bootstrap/inisialisasi (`Must`).
- **Entry / Key Files**:
  - [`panic.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/errors/panic.go) — Generic helper `Must[T any](v T, err error) T`.
  - [`panic_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/errors/panic_test.go) — Unit test untuk memastikan panic terpanggil dengan tepat jika error != nil.
- **Dependencies**: Standard library Go.
- **Consumers**: Digunakan di semua modul atau aplikasi consumer saat memuat config atau resource tak boleh gagal pada startup.
- **External Integrations**: None.
- **Key Notes**: Dikhususkan untuk fail-fast initialization, bukan untuk menggantikan pengecekan error normal pada runtime request path.

---

## `http/client`
- **Responsibility**: Wrapper HTTP client dengan singleton pattern (`sync.Once`), serialization JSON generic (`Post`, `Put`, `Patch`), dan default timeout.
- **Entry / Key Files**:
  - [`client.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/http/client/client.go) — Fungsi `Get`, `Post[T]`, `Put[T]`, `Patch[T]`, dan singleton client internal.
  - [`client_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/http/client/client_test.go) — Test suite client HTTP.
- **Dependencies**: Standard library `net/http`, `sync`, `encoding/json`.
- **Consumers**: Internal service communication atau external REST API calls.
- **External Integrations**: HTTP REST APIs.
- **Key Notes**: Memiliki timeout default 5 detik. Memeriksa validitas payload JSON sebelum dikirim ke remote target.

---

## `http/server/fiber`
- **Responsibility**: Factory helper untuk standarisasi pembuatan web server menggunakan framework Fiber v3.
- **Entry / Key Files**:
  - [`fiber.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/http/server/fiber/fiber.go) — Struct `Options`, fungsi `NewFiber(opts)`, dan `DefaultOptions()`.
  - [`fiber_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/http/server/fiber/fiber_test.go) — Pengujian instansiasi Fiber app dan konfigurasi trust proxy.
- **Dependencies**: `github.com/gofiber/fiber/v3`.
- **Consumers**: Microservice HTTP API di ekosistem PakaiWA.
- **External Integrations**: None.
- **Key Notes**: Secara default mengaktifkan `TrustProxy: true` dan `EnableIPValidation: true` untuk lingkungan cloud/reverse-proxy.

---

## `messaging/event`
- **Responsibility**: Abstraksi domain event interface untuk menjamin keseragaman event contract di pipeline event-driven.
- **Entry / Key Files**:
  - [`event.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/messaging/event/event.go) — Interface `Event` (`EventID()`, `EventName()`, `EventKey()`).
- **Dependencies**: Standard library Go.
- **Consumers**: Producer dan consumer Kafka/HTTP di ekosistem PakaiWA.
- **External Integrations**: None.
- **Key Notes**: Menjadi kontrak dasar payload yang akan dipublish ke Kafka atau webhook HTTP.

---

## `messaging/producer`
- **Responsibility**: Core interface untuk message publisher agnostik terhadap protokol transport.
- **Entry / Key Files**:
  - [`producer.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/messaging/producer/producer.go) — Interface `MessageProducer` (`Send`, `Flush`, `Close`).
- **Dependencies**: Standard library `context`.
- **Consumers**: `messaging/kafka` dan `messaging/http`.
- **External Integrations**: None.
- **Key Notes**: Mengabstraksi transport message sehingga handler bisnis tidak terikat langsung ke library Kafka atau HTTP webhook.

---

## `messaging/kafka`
- **Responsibility**: Implementasi high-throughput Kafka Producer & Consumer berbasis Confluent librdkafka, dilengkapi background delivery report poll loop dan generic typed event producer.
- **Entry / Key Files**:
  - [`producer.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/messaging/kafka/producer.go) — Struct `KafkaProducer`, `Producer[T event.Event]`, dan `StartProducerPollLoop`.
  - [`consumer.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/messaging/kafka/consumer.go) — Factory `NewKafkaConsumer` dengan bootstrap servers & group ID.
  - [`kafka_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/messaging/kafka/kafka_test.go), [`producer_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/messaging/kafka/producer_test.go) — Unit & mock testing producer Kafka.
- **Dependencies**: `github.com/confluentinc/confluent-kafka-go/v2/kafka`, `github.com/sirupsen/logrus`, `observability/logging/ctxmeta`.
- **Consumers**: Gateway, worker, dan backend services PakaiWA untuk event streaming WhatsApp message & status.
- **External Integrations**: Apache Kafka broker.
- **Key Notes**: Menyisipkan header `device_id` (`clientJID`) di setiap message Kafka untuk routing spesifik session WhatsApp. Menggunakan background goroutine untuk konsumsi delivery report event channel.

---

## `messaging/http`
- **Responsibility**: Implementasi `producer.MessageProducer` berbasis HTTP POST (Webhook) dengan proteksi skema URL (SSRF mitigation), context logging, dan status header.
- **Entry / Key Files**:
  - [`producer.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/messaging/http/producer.go) — Implementasi `HttpProducer` dengan header `X-PakaiWA-Topic`, `X-PakaiWA-Key`, dan `X-Device-ID`.
  - [`producer_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/messaging/http/producer_test.go) — Pengujian webhook delivery, header, dan error status handling.
- **Dependencies**: `messaging/producer`, `observability/logging/ctxmeta`, `github.com/sirupsen/logrus`.
- **Consumers**: Webhook dispatcher service untuk meneruskan event PakaiWA ke endpoint klien eksternal.
- **External Integrations**: External HTTP Webhooks.
- **Key Notes**: Menolak URL dengan skema non-http/https pada inisialisasi; melakukan pembacaan dan pembuangan body (`io.Copy(io.Discard, ...)`) agar connection reuse HTTP tetap optimal.

---

## `observability/logging`
- **Responsibility**: Pengelolaan logging terstruktur terstandarisasi dengan domain-scoped logger (`app`, `db`, `http`, `kafka`, `wa`), Ordered JSON Formatter (urutan field deterministik untuk parsing log/Loki/ELK), serta propagasi context metadata (Trace ID & Logger instance).
- **Entry / Key Files**:
  - [`ctxmeta/key.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta/key.go) — Context key constants.
  - [`ctxmeta/logger.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta/logger.go) — Getter & setter context logger domain-scoped (`LoggerApp`, `LoggerDB`, `LoggerHTTP`, `LoggerKafka`, `LoggerWA`).
  - [`ctxmeta/trace.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta/trace.go) — `WithTraceID` & `TraceID`.
  - [`logrus/logger.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/observability/logging/logrus/logger.go) — `OrderedJSONFormatter`, `NewLoggers`, dan caller resolver (`resolveCaller`).
  - [`logrus/logger_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/observability/logging/logrus/logger_test.go) — Verifikasi formatting, caller detection, dan output deterministik.
- **Dependencies**: `github.com/sirupsen/logrus`.
- **Consumers**: Semua modul dalam platform dan semua microservice PakaiWA downstream.
- **External Integrations**: Standard Out (stdout) -> Log Shipper (FluentBit/Promtail/Datadog).
- **Key Notes**: Field log diatur dalam urutan tetap: `time`, `level`, `trace_id`, `msg`, `caller`, diikuti data terurut alfabetis, dan diakhiri `module`.

---

## `observability/metrics`
- **Responsibility**: Inisialisasi Prometheus metric counters & histograms serta Fiber handler adapter untuk endpoint scraping `/metrics`.
- **Entry / Key Files**:
  - [`prometheus.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/observability/metrics/prometheus.go) — Metrik `HttpRequests` (`http_requests_total`), `HttpDuration` (`http_request_duration_seconds`), dan `PrometheusHandler()`.
  - [`prometheus_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/observability/metrics/prometheus_test.go) — Test handler scraping Fiber metrics.
- **Dependencies**: `github.com/prometheus/client_golang`, `github.com/gofiber/fiber/v3`.
- **Consumers**: Microservice HTTP server di PakaiWA.
- **External Integrations**: Prometheus scraper.
- **Key Notes**: Menggunakan `httptest.ResponseRecorder` untuk menjembatani standard `promhttp.Handler()` ke Fiber v3 context handler.

---

## `runtime/shutdown`
- **Responsibility**: Manajemen graceful shutdown dengan mendengarkan sinyal OS (`SIGINT`, `SIGTERM`) atau context cancellation.
- **Entry / Key Files**:
  - [`signal.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/runtime/shutdown/signal.go) — Fungsi `WaitForSignal()` dan `Wait(ctx, sigs...)`.
  - [`signal_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/runtime/shutdown/signal_test.go) — Unit test sinyal dan context cancellation.
- **Dependencies**: Standard library `os/signal`, `syscall`, `context`.
- **Consumers**: Fungsi `main()` pada setiap service aplikasi PakaiWA.
- **External Integrations**: OS kernel process signals.
- **Key Notes**: Menghentikan blocking listener secara aman saat pod Kubernetes dihentikan (`SIGTERM`).

---

## `security/password`
- **Responsibility**: Enkripsi hashing dan verifikasi kata sandi menggunakan algoritma bcrypt.
- **Entry / Key Files**:
  - [`password.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/security/password/password.go) — Fungsi `Hash(plain)` dan `Compare(hashed, plain)`.
  - [`password_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/security/password/password_test.go) — Pengujian kecocokan hash dan invalid comparison.
- **Dependencies**: `golang.org/x/crypto/bcrypt`.
- **Consumers**: Layanan autentikasi akun & user management.
- **External Integrations**: None.
- **Key Notes**: Menggunakan `bcrypt.DefaultCost` (cost 10).

---

## `validation`
- **Responsibility**: Validasi struct menggunakan `validator/v10` dengan integrasi tag `json` untuk nama field error serta utilitas context logging terstandarisasi.
- **Entry / Key Files**:
  - [`validator.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/validation/validator.go) — Factory `NewValidator` dengan `RegisterTagNameFunc` berbasis tag JSON.
  - [`validator_logging.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/validation/validator_logging.go) — Helper `LogValidationErrors` dan `TraceIDFromContext`.
  - [`validator_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/validation/validator_test.go), [`validator_logging_test.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/validation/validator_logging_test.go) — Test validasi skenario success/failure.
- **Dependencies**: `github.com/go-playground/validator/v10`, `observability/logging/ctxmeta`, `github.com/sirupsen/logrus`.
- **Consumers**: HTTP Handler dan Message Consumer saat memproses incoming DTO payload.
- **External Integrations**: None.
- **Key Notes**: Secara otomatis memetakan field struct Go ke nama atribut JSON di pesan error log validation.
