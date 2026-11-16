# Instruksi Pengembangan ERP

## Konfirmasi skema database sebelum modul baru

- Sebelum mulai mengimplementasikan modul baru, ajukan dan konfirmasikan rancangan skema database dengan pemilik produk. Jangan membuat kode, migration, atau perubahan database untuk modul tersebut sebelum skemanya disetujui.
- Rancangan yang diajukan harus menjelaskan tabel, kolom dan tipe data, primary/foreign key, relasi, unique constraint/index, nullability/default, serta status dan aturan siklus hidup transaksi yang relevan.
- Database ERP dipakai bersama POS. Periksa schema/migration POS dan jelaskan dampak kompatibilitas serta repo yang akan menjadi lokasi migration. Jangan mengasumsikan kolom atau relasi yang belum disepakati.
- Setelah menyampaikan rancangan, tunggu konfirmasi eksplisit sebelum meneruskan implementasi yang bergantung pada rancangan itu.

## Pola tabel transaksi

- Setiap transaksi yang memiliki baris item menggunakan tabel header dan detail yang terpisah.
- Gunakan nama tabel berpasangan dengan pola `<transaction>_header` dan `<transaction>_detail`, misalnya `purchase_request_header` dan `purchase_request_detail`.
- Header menyimpan identitas dokumen dan atribut tingkat dokumen, seperti nomor, company, branch, tanggal, status, pemohon, dan catatan.
- Detail menyimpan baris transaksi, termasuk foreign key ke header, item, kuantitas, satuan, serta catatan baris bila diperlukan.
- Tentukan foreign key, cascade/restrict behavior, indeks, dan kebutuhan snapshot data detail dalam rancangan skema sebelum implementasi.

## Database dan migration

- Jangan menjalankan migration terhadap database aktif tanpa instruksi eksplisit.
- Karena schema database digunakan bersama POS, letakkan migration pada repo yang disepakati sebagai pemilik migration dan sinkronkan dokumentasinya.

## Bahasa

- bahasa dalam seluruh konten di dalam aplikasi adalah bahasa inggris

## Layer frontend

- `web/src/views/` berisi halaman yang berfokus pada template dan komposisi komponen.
- `web/src/components/` berisi komponen UI kecil yang dapat digunakan ulang.
- `web/src/composables/` menampung logic fitur dan state reaktif.
- `web/src/services/` menjadi satu-satunya layer untuk pemanggilan API.
- `web/src/stores/` menampung state global Pinia.
- `web/src/utils/` berisi helper murni tanpa state atau akses API.
- Hindari menaruh logic bisnis atau request API langsung di view.
