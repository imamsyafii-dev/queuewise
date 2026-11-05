# Queuewise

Queuewise adalah MVP sistem appointment dan antrean untuk bisnis berbasis layanan seperti klinik, salon, dan bengkel.

## Fitur MVP

- Booking appointment publik
- Validasi bentrok jadwal staff
- Kode booking unik
- Check-in dan nomor antrean harian
- Operasional antrean: panggil, mulai layanan, dan selesaikan
- Layar antrean publik real-time melalui Redis Pub/Sub dan SSE
- Dashboard operasional lokal

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

Migration ini juga membuat data demo Klinik Sehat Sentosa, dua layanan, dan dua staff.

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

## Catatan MVP

Dashboard admin saat ini sengaja berjalan tanpa autentikasi agar alur lokal dapat dicoba langsung. Sebelum deployment, tambahkan authentication/authorization, rate limiting untuk endpoint publik, pembayaran, dan worker Redis untuk reminder.
