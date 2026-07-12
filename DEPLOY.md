# Deploy DailyGrind — panduan biaya terendah

Dua jalur. Pilih sesuai kebutuhan:

| Jalur | Biaya | Cocok untuk | Catatan |
|-------|-------|-------------|---------|
| **A. Satu VPS + Docker** | **~Rp60–75rb/bln** (€4) | Produksi / selalu-nyala | Semua di satu tempat (DB gratis, self-host), tanpa cold start. **Rekomendasi.** |
| **B. Free tier** | **Rp0** | Demo / portfolio | Ada cold start / sleep; Postgres free bisa expired. |

---

## Jalur A — 1 VPS (~€4/bln, rekomendasi)

Contoh VPS termurah yang andal: **Hetzner** CX22 / CAX11 (€3.79/bln), atau Contabo/DigitalOcean/Vultr/Linode. Butuh 1 domain (mis. dari Cloudflare/Niagahoster).

**1. Provision VPS + domain**
- Buat VPS Ubuntu 24.04 (1 vCPU / 2–4 GB RAM cukup).
- Arahkan **A record** domainmu ke IP VPS (mis. `dailygrind.example.com` → `1.2.3.4`).

**2. Install Docker di VPS**
```bash
ssh root@IP_VPS
curl -fsSL https://get.docker.com | sh
```

**3. Ambil kode + konfigurasi env**
```bash
git clone <repo-url> dailygrind && cd dailygrind
cp .env.prod.example .env
nano .env          # isi SITE_DOMAIN, SITE_URL, password DB, JWT_SECRET
# JWT_SECRET acak: openssl rand -base64 48
```

**4. Jalankan seluruh stack (DB + API + Nuxt + Caddy/HTTPS)**
```bash
docker compose -f docker-compose.prod.yml up -d --build
```
Caddy otomatis mengurus sertifikat HTTPS (Let's Encrypt) untuk domainmu. Buka
`https://SITE_DOMAIN` — API demo sudah ter-seed (login `nadia@email.com` / `password123`).

**5. (Opsional) Impor Alkitab lengkap — sekali saja**
```bash
docker compose -f docker-compose.prod.yml run --rm backend /app/import-bible
```
Data tersimpan di volume `dg_pgdata` dan bertahan selama tidak `down -v`.

**Update ke depannya:**
```bash
git pull && docker compose -f docker-compose.prod.yml up -d --build
```

---

## Jalur B — Free tier (Rp0)

Pisah 3 layanan, semua punya paket gratis:

| Bagian | Layanan gratis | Cara |
|--------|----------------|------|
| PostgreSQL | **Neon** (neon.tech) | Buat project → salin `DATABASE_URL` (sudah `?sslmode=require`) |
| Go API | **Render** / **Koyeb** / **Fly.io** | Deploy dari repo, root `backend/`, pakai Dockerfile |
| Nuxt | **Cloudflare Pages** / **Vercel** | Deploy dari repo, root `frontend/` |

**1. Postgres (Neon)** — buat project, simpan `DATABASE_URL`.

**2. Backend (Render contoh)**
- New → Web Service → connect repo → Root Directory `backend`, Runtime **Docker**.
- Environment variables:
  - `DATABASE_URL` = dari Neon
  - `JWT_SECRET` = string acak ≥32 char
  - `CORS_ORIGIN` = URL frontend (mis. `https://dailygrind.pages.dev`)
  - (`PORT` diisi otomatis oleh platform; app sudah membacanya)
- Health check path: `/api/health`.
- Setelah live, isi Alkitab dari laptop (arahkan ke Neon):
  ```bash
  cd backend
  DATABASE_URL="<url-neon>" go run ./cmd/import-bible
  ```

**3. Frontend (Cloudflare Pages contoh)**
- Root `frontend/`, build command `npm run build`, framework preset **Nuxt**.
- Build env var: `NUXT_PUBLIC_API_BASE` = `https://<backend-url>/api`.
- Setelah dapat URL Pages, kembali ke backend dan set `CORS_ORIGIN` ke URL itu.

> Catatan free tier: Render free "tidur" setelah ~15 menit tanpa trafik (request
> pertama lambat). Neon free tier auto-suspend saat idle. Untuk demo tidak masalah;
> untuk produksi pilih Jalur A.

---

## Ringkasan variabel penting

| Variabel | Di mana | Contoh |
|----------|---------|--------|
| `DATABASE_URL` | backend | `postgres://user:pass@host:5432/db?sslmode=require` |
| `JWT_SECRET` | backend | ≥32 char acak |
| `CORS_ORIGIN` | backend | harus = origin frontend |
| `NUXT_PUBLIC_API_BASE` | frontend | `https://domain/api` |
