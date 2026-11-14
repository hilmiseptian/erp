# PRD Supplier

**Versi dokumen:** 1.0.0 (pra-production) · **Status:** Implementasi awal tersedia

## Tujuan dan fungsi

Menyimpan supplier per company sebagai master untuk proses purchasing.

## Data

Model/migrasi POS yang menjadi schema bersama memuat `company_id`, `code`, `name`, `contact_person`, `phone`, `email`, `address`, `is_active`, `created_at`, `updated_at`; `code` unik per company. Detail ada pada migrasi supplier POS.

## Perilaku API ERP saat ini

CRUD supplier: list, detail, create, update, delete. Query dibatasi `company_id`; kode supplier dan nama wajib; code/name maksimal 30/150 karakter. Belum ada filter/search/pagination yang didokumentasikan.

## Pertanyaan lanjutan

Apakah delete harus soft delete, supplier dapat dipakai lintas branch, dan field pajak/termin pembayaran dibutuhkan? Ditentukan sebelum invoice/purchasing diperluas.
