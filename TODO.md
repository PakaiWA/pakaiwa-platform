# Project TODO & Technical Debt

> **Audit Date**: 2026-09-12  
> **Source Scan**: Full codebase regex inspection (`TODO`, `FIXME`, `HACK`, `XXX`, `BUG`, `DEPRECATED`)

---

## 1. Immediate Tasks
_Tugas atau perbaikan yang berkaitan langsung dengan keandalan operasional atau pengujian._

- [x] [`http/client/client.go:68-69`](file:///Users/i/work/PakaiWA/pakaiwa-platform/http/client/client.go#L68-L69): Hapus debug logging `fmt.Println` (`Raw bytes:` dan `As string:`) serta hilangkan variable shadowing pada unmarshal check.
- [ ] [`messaging/kafka/producer.go:37`](file:///Users/i/work/PakaiWA/pakaiwa-platform/messaging/kafka/producer.go#L37): Pada `NewKafkaProducer`, saat error fungsi mengembalikan `nil` tanpa nilai error. Pertimbangkan refactor signature menjadi `(producer.MessageProducer, error)` agar caller dapat mendeteksi kegagalan bootstrap Kafka broker secara idiomatik.

---

## 2. Existing Code Annotations (TODO / FIXME)
_Daftar TODO/FIXME yang tercantum langsung di dalam source code._

*Tidak ditemukan anotasi `TODO`, `FIXME`, `HACK`, atau `BUG` yang tertinggal di dalam source code saat ini (0 occurrences).*

---

## 3. Technical Debt & Structural Improvements
_Pekerjaan arsitektural dan refactoring jangka menengah untuk maintainability dan ekstensi platform._

- [ ] **HTTP Client Configurable Pool & Timeout**:
  - [`http/client/client.go:85`](file:///Users/i/work/PakaiWA/pakaiwa-platform/http/client/client.go#L85): Saat ini timeout HTTP client di-hardcode sebesar 5 detik dan menggunakan default transport global. Tambahkan dukungan konfigurasi `Transport` (MaxIdleConns, MaxIdleConnsPerHost, IdleConnTimeout) dan dynamic timeout.
- [ ] **OpenTelemetry / W3C Distributed Tracing**:
  - [`observability/logging/ctxmeta/trace.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/observability/logging/ctxmeta/trace.go): TraceID saat ini dikelola via string manual di context. Dapat dievolusikan untuk kompatibel dengan standard W3C TraceContext atau OpenTelemetry Span Context.
- [ ] **Graceful Drain on Fiber Shutdown**:
  - [`http/server/fiber/fiber.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/http/server/fiber/fiber.go): Buat helper terintegrasi yang menghubungkan `runtime/shutdown` dengan `app.ShutdownWithContext(...)` dari Fiber v3 untuk menyederhanakan boiler-plate shutdown di consumer service.
- [ ] **Circuit Breaker / Retry Policy on Messaging Producer**:
  - [`messaging/kafka/producer.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/messaging/kafka/producer.go) & [`messaging/http/producer.go`](file:///Users/i/work/PakaiWA/pakaiwa-platform/messaging/http/producer.go): Tambahkan built-in retry backoff policy yang dapat dikonfigurasi saat menghadapi `ErrQueueFull` di Kafka atau transient HTTP 5xx pada webhook.
