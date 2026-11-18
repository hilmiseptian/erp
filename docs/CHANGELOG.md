# Changelog — ERP

## Versi 1.0.0 (pra-production)

Dokumen ini diperbarui terus selama manual testing dan pengembangan. Jangan buat entri versi 2 atau versi baru sebelum produk production.

- Menambahkan modul Supplier ERP.
- Menambahkan MVP Master Item ERP: list, pencarian/filter, detail, create/update, dan aktif/nonaktif.
- Menambahkan MVP Purchase Request sebagai child dari Inventory; approval belum diterapkan.
- Menyesuaikan PR ke tabel `purchase_request_header` dan `purchase_request_detail`, termasuk supplier serta total kuantitas/nilai.
- Menambahkan `req_date` pada header PR sebagai tanggal barang dibutuhkan.
- Menambahkan MVP Purchase Order yang dapat dibuat parsial dari satu PR; sisa kuantitas PR dihitung dari PO aktif.
- Item ERP membaca saldo yang ada; saldo tidak dapat diedit dari master item.
- Menambahkan dokumentasi Swagger awal dari anotasi Go.
- Menambahkan halaman Swagger UI pada web ERP (`/api-docs`) dengan Try it out.

Catat fitur yang selesai, perubahan perilaku, dan batasan yang relevan di bawah entri versi ini.
