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

---

## ADR-006: Optimasi Komprehensif Kode, Memori, dan Pipeline Linting/Testing

- **Status**: Accepted
- **Date**: 2026-09-12
- **Source**: Repository Optimization & Code Quality Audit
- **Context**:
  Pemeriksaan menyeluruh pada source code dan tooling CI/CD menemukan beberapa area yang perlu dioptimalkan:
  1. Adanya stdout debug printing `fmt.Println` di jalur HTTP client produksi yang menimbulkan overhead I/O dan alokasi memori.
  2. Alokasi heap berulang pada fungsi `Get40Space()` di layer validation logging melalui `strings.Repeat(" ", 40)`.
  3. Padding memori struct tidak optimal (*fieldalignment*) pada struct domain/konfigurasi (`Options`, `ConsumerConfig`, dan struct pengujian).
  4. Variable shadowing dan unhandled error return di beberapa file test dan implementasi producer.
  5. Konfigurasi linter `golangci-lint` yang memicu puluhan warning tidak relevan untuk pustaka internal, serta ketidaksinkronan target CI antar-tool.
- **Decision**:
  Menerapkan serangkaian optimasi kode dan proses berikut:
  1. **HTTP Client Cleanliness**: Menghapus `fmt.Println` debug logging di `http/client/client.go` dan membersihkan variable shadowing pada error unmarshaling.
  2. **Zero-Allocation Log Padding**: Mengganti alokasi dinamis `strings.Repeat` di `validation/validator_logging.go` dengan string konstanta statis `fortySpaces`.
  3. **Struct Memory Alignment**: Mengoptimalkan urutan field struct pada `http/server/fiber.Options`, `messaging/kafka.ConsumerConfig`, dan mock structs untuk meminimalkan padding byte di arsitektur 64-bit.
  4. **Error Handling & Type Safety**: Memperbaiki pengecekan type assertion di `mustNewHttpProducer` dan menambahkan anotasi penanganan error (`nolint:errcheck`) pada pembersihan defer soket Kafka/HTTP test.
  5. **Linter & Pipeline Harmonization**: Menyelaraskan aturan `revive` pada `.golangci.yml` sehingga seluruh lint check (`govet`, `errcheck`, `staticcheck`, `revive`) dan hook `pre-commit` serta `pre-push` lulus 100% tanpa noise.
- **Consequences**:
  - **Positif**: Throughput kode meningkat (zero-overhead logging & zero-alloc padding), efisiensi cache CPU lebih tinggi dari struct alignment yang optimal, serta linting pipeline `make lint` dan `pre-commit`/`pre-push` bersih dan berstatus 0 issues.
  - **Negatif / Trade-off**: Urutan field di struct berubah secara internal (namun backwards compatible karena menggunakan nama field eksplisit saat inisialisasi).
