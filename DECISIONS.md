# Architecture Decision Records (ADR)

> **Document Principle**: Append-only. Do not overwrite or mutate past records without marking them superseded.

---

## ADR-001: Posisi Repository Sebagai Shared Core Library Ekosistem PakaiWA

- **Status**: Accepted
- **Date**: 2026-09-12
- **Source**: Developer interview & Codebase evidence
- **Context**:
  PakaiWA memiliki berbagai microservice dan engine (seperti whatsmeow engine, gateway API, webhook worker). Diperlukan repositori terpusat untuk menstandardisasi koneksi database, structured logging, HTTP server framework, dan event streaming tanpa duplikasi kode.
- **Decision**:
  Repository `github.com/PakaiWA/pakaiwa-platform` ditetapkan sebagai **Shared Core Library** yang diimpor via `go.mod` oleh semua downstream service di ekosistem PakaiWA.
- **Consequences**:
  - **Positif**: Konsistensi arsitektur, standardisasi library (Fiber v3, pgx v5, logrus, confluent-kafka-go), dan kemudahan patching kerentanan secara terpusat.
  - **Negatif / Trade-off**: Perubahan breaking change pada interface library mengharuskan update dependensi dan pengujian pada setiap service downstream.

---

## ADR-002: Strategi Penanganan Pesan Gagal Kafka Berbasis Cluster & Topic

- **Status**: Accepted
- **Date**: 2026-09-12
- **Source**: Developer interview
- **Context**:
  Dalam pengiriman dan pemrosesan event WhatsApp, kegagalan sementara jaringan atau broker dapat menyebabkan message drop atau lag.
- **Decision**:
  Strategi fault tolerance dan pengulangan penanganan pesan dipusatkan di tingkat **Kafka Cluster & Topic Pattern** (misalnya pengalihan ke retry topic, Dead Letter Queue / DLQ topic terpisah, serta consumer backoff), bukan dengan buffer queue database di level library platform.
- **Consequences**:
  - **Positif**: Library tetap lean dan berperforma tinggi tanpa overhead persistensi lokal; pemisahan poison messages tidak memblokir partisi utama.
  - **Negatif / Trade-off**: Konfigurasi topik, consumer group routing, dan alerting DLQ harus dipelihara dengan cermat di level infrastruktur Kafka.

---

## ADR-003: Standarisasi Autentikasi Password dengan Bcrypt

- **Status**: Accepted
- **Date**: 2026-09-12
- **Source**: Developer interview & Codebase evidence (`security/password`)
- **Context**:
  Kebutuhan hashing kata sandi untuk autentikasi kredensial pengguna dan service account di platform PakaiWA.
- **Decision**:
  Menggunakan implementasi Bcrypt dari standard extension library Go (`golang.org/x/crypto/bcrypt`) dengan `bcrypt.DefaultCost`. Standard ini ditetapkan memadai untuk kebutuhan keamanan autentikasi saat ini.
- **Consequences**:
  - **Positif**: Teruji, aman dari serangan brute-force, zero external C-dependency, dan performa yang terukur.
  - **Negatif / Trade-off**: Memiliki batasan input maksimum 72 byte khas bcrypt (harus diperhatikan di layer validasi DTO).

---

## ADR-004: Adopsi Observability Stack LGTM / Grafana Ecosystem (Prometheus, Loki, Alloy, Tempo)

- **Status**: Accepted
- **Date**: 2026-09-12
- **Source**: Developer interview
- **Context**:
  Platform membutuhkan visibilitas menyeluruh (Metrics, Logs, Traces) yang terintegrasi secara mulus di seluruh microservice.
- **Decision**:
  Standar telemetry PakaiWA berorientasi penuh pada stack **Grafana Ecosystem**:
  - **Metrics**: Prometheus (via `observability/metrics`)
  - **Logs**: Loki (didukung oleh `OrderedJSONFormatter` logrus yang deterministik)
  - **Telemetry Agent / Collector**: Grafana Alloy
  - **Distributed Tracing**: Grafana Tempo (berkorelasi dengan `trace_id` yang dipropagasi via `ctxmeta`)
- **Consequences**:
  - **Positif**: Korelasi metrik, log, dan trace satu klik di Grafana dashboard; format JSON log yang saat ini dibangun sangat optimal untuk ingestion engine Loki dan Alloy.
  - **Negatif / Trade-off**: Header W3C TraceContext atau OpenTelemetry exporter perlu dijaga kesesuaiannya agar sinkron saat dikirim melalui Grafana Alloy ke Tempo.

---

## ADR-005: Penamaan Tag Rilis Otomatis Berbasis Waktu Jakarta (WIB)

- **Status**: Accepted
- **Date**: 2026-09-12
- **Source**: CI/CD Workflow (`.github/workflows/CI.yaml`)
- **Context**:
  Kebutuhan versi rilis otomatis pada setiap commit signifikan di branch utama yang mencerminkan waktu deployment aktual.
- **Decision**:
  Sistem CI/CD GitHub Actions secara otomatis menghitung dan mempublikasikan git tag format `v0.yy.m-dhhmm` menggunakan zona waktu `Asia/Jakarta` (WIB) apabila terjadi perubahan signifikan pada kode Go.
- **Consequences**:
  - **Positif**: Pelacakan build dan rilis dependensi menjadi sangat mudah dikorelasikan langsung dengan timeline operasional tim.
  - **Negatif / Trade-off**: Format non-SemVer murni mengharuskan consumer menggunakan tag rilis spesifik atau pseudo-versioning Go tooling.
