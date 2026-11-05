# Queuewise

Queuewise adalah MVP sistem appointment dan antrean untuk bisnis berbasis layanan seperti klinik, salon, dan bengkel. Sistem memisahkan pengalaman customer untuk booking dengan area operasional staff untuk mengelola kedatangan dan antrean.

## Status pengembangan

Alur inti sudah dapat dicoba secara lokal:

```text
Customer membuat booking
        ↓
Staff melakukan check-in
        ↓
Customer masuk antrean dan mendapat nomor
        ↓
Staff memanggil → mulai layanan → selesai
        ↓
Layar antrean publik diperbarui real-time
```

Lihat [roadmap pengembangan](docs/ROADMAP.md) untuk fitur berikutnya dan statusnya.

## Fitur saat ini

- Booking appointment publik dengan kode booking unik
- Validasi jadwal staff yang bentrok
- Data demo organisasi, cabang, layanan, dan staff
- Check-in dan pembuatan nomor antrean harian
- Operasional antrean: panggil, mulai layanan, dan selesaikan
- Layar antrean publik real-time memakai Redis Pub/Sub dan SSE
- Dashboard staff lokal untuk melihat appointment dan antrean

## Menjalankan aplikasi

### 1. Mulai PostgreSQL dan Redis

```bash
docker compose up -d
```

### 2. Siapkan backend dan database

```bash
cd backend
cp .env.example .env
cd ..
docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U queuewise -d queuewise < backend/migrations/000001_create_queuewise_schema.sql
```

Migration ini membuat data demo Klinik Sehat Sentosa, satu cabang, dua layanan, dan dua staff.

### 3. Jalankan backend

```bash
cd backend
go run ./cmd/api
```

Backend tersedia di `http://localhost:8080`.

### 4. Jalankan frontend

Pada terminal lain:

```bash
cd frontend
cp .env.local.example .env.local
npm run dev
```

Frontend tersedia di `http://localhost:3000`.

## Rute aplikasi

| URL | Fungsi |
| --- | --- |
| `/` | Landing page |
| `/booking` | Form booking customer |
| `/booking/[kode]` | Konfirmasi booking |
| `/queue/jakarta-selatan` | Layar antrean publik |
| `/admin` | Dashboard operasional demo |
| `/admin/appointments` | Daftar dan check-in appointment |
| `/admin/queue` | Panggil, mulai, dan selesaikan antrean |

## Stack teknologi

| Layer | Teknologi | Peran |
| --- | --- | --- |
| Backend | Go + Gin | REST API dan aturan bisnis booking/antrean |
| Database utama | PostgreSQL | Menyimpan data organisasi, customer, appointment, dan antrean |
| Real-time | Redis Pub/Sub + SSE | Mengabarkan perubahan antrean ke semua layar publik |
| Frontend | Next.js + TypeScript | Area customer dan dashboard staff |
| Local infrastructure | Docker Compose | Menjalankan PostgreSQL dan Redis |

## Catatan MVP

Dashboard admin saat ini sengaja berjalan tanpa autentikasi agar alur lokal dapat dicoba langsung. Jangan gunakan mode ini pada production. Authentication, reminder, pembayaran, dan fitur production lainnya tercatat di [roadmap](docs/ROADMAP.md).
