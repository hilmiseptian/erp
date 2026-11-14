# PRD Authentication

**Versi dokumen:** 1.0.0 (pra-production) · **Status:** Deskripsi implementasi saat ini

## Kondisi saat ini

Web ERP memakai alur login Laravel POS untuk memperoleh token. API ERP memvalidasi token ke endpoint POS `/api/user`, lalu membaca identitas dan `company_id`. Endpoint ERP `/api/me` mengembalikan pengguna terautentikasi. Middleware supplier mensyaratkan pengguna terhubung ke company.

## Kebutuhan

- Akses API ERP memerlukan token yang valid.
- Data perusahaan dibatasi menurut `company_id`.
- Role/permission ERP masih keputusan terbuka; pemilik produk cenderung ingin pengelolaan role di ERP.
- Aturan akses lintas branch, logout, expiry/revocation token, dan paket fitur perlu disepakati.

## Belum ditentukan

Sumber otoritatif role/permission, kebijakan refresh/revocation, session handling, dan batas izin branch. Lihat [API Auth](../api/API-AUTH.md).
