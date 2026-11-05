# Queuewise Roadmap

Dokumen ini menjadi panduan pengembangan Queuewise setelah MVP booking dan antrean.

## Sudah tersedia

- Booking appointment customer
- Kode booking unik
- Validasi jadwal staff agar tidak bentrok
- Check-in dan nomor antrean per cabang per hari
- Status antrean: `BOOKED` → `IN_QUEUE` → `CALLED` → `IN_SERVICE` → `COMPLETED`
- Dashboard staff lokal
- Layar antrean publik real-time melalui Redis Pub/Sub dan SSE

## Prioritas 1 — Siap dipakai staff

### Authentication dan role

- Login dan logout staff
- Role `owner`, `admin`, `receptionist`, dan `staff`
- Proteksi halaman serta API admin
- Audit log untuk perubahan appointment dan antrean

### CRUD operasional

- Kelola organisasi dan cabang
- Kelola layanan: nama, durasi, harga, dan status aktif
- Kelola staff dan layanan yang dapat ditangani
- Jam operasional, jeda, hari libur, dan slot booking yang valid

### Customer booking

- Reschedule dan cancel menggunakan kode booking
- Halaman detail booking yang dapat dibuka kembali
- Kapasitas antrean dan estimasi waktu tunggu

## Prioritas 2 — Notifikasi dan pembayaran

### Reminder

- Worker terpisah untuk reminder H-1 dan H-2 jam
- Redis queue untuk pekerjaan pengiriman dan retry
- Integrasi WhatsApp atau email provider
- Template pesan dan riwayat reminder
- Dead-letter queue untuk reminder yang terus gagal

### Pembayaran

- Invoice untuk appointment
- Deposit atau pembayaran penuh
- Integrasi payment gateway
- Webhook pembayaran yang idempotent
- Pembatalan dan refund

## Prioritas 3 — Pengalaman antrean

- Tampilan TV fullscreen dengan branding cabang
- Notifikasi suara saat nomor dipanggil
- Estimasi waktu tunggu berdasarkan durasi layanan
- Banyak loket atau ruang layanan dalam satu cabang
- Analitik antrean: waktu tunggu rata-rata, no-show, dan throughput staff

## Prioritas 4 — Production readiness

- Unit test, integration test, dan end-to-end test
- Rate limiting endpoint booking publik
- Validasi input, structured logging, request ID, metrics, dan tracing
- Dockerfile untuk API, worker, dan frontend
- CI/CD, environment production, backup PostgreSQL, dan migration runner
- Dokumentasi API melalui OpenAPI

## Prinsip implementasi

- PostgreSQL tetap menjadi sumber data permanen.
- Redis digunakan untuk kebutuhan cepat dan asynchronous: real-time event, cache, rate limit, serta job reminder.
- Setiap fitur baru diselesaikan bersama test, dokumentasi, dan perubahan schema yang dapat dimigrasikan.
