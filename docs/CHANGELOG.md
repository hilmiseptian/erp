# Changelog — ERP

## Versi 1.0.0 (pra-production)

Dokumen ini diperbarui terus selama manual testing dan pengembangan. Jangan buat entri versi 2 atau versi baru sebelum produk production.

- Menambahkan modul Supplier ERP.
- Menambahkan MVP Master Item ERP: list, pencarian/filter, detail, create/update, dan aktif/nonaktif.
- Menambahkan MVP Purchase Request sebagai child dari Inventory; approval belum diterapkan.
- Menyesuaikan PR ke tabel `purchase_request_header` dan `purchase_request_detail`, termasuk supplier serta total kuantitas/nilai.
- Menambahkan `req_date` pada header PR sebagai tanggal barang dibutuhkan.
- Menambahkan MVP Purchase Order yang dapat dibuat parsial dari satu PR; sisa kuantitas PR dihitung dari PO aktif.
- Mengubah kuantitas PR, PO, dan penerimaan menjadi bilangan bulat.
- Menambahkan Goods Receipt parsial dari PO; harga mengikuti harga PO dan posting langsung menambah saldo branch.
- Menambahkan modul Stock dengan saldo per company/branch/item, pembukaan saldo nol, dan rincian pergerakan stock card; HPP masuk mengikuti harga PO.
- Menambahkan MVP Stock Adjustment dengan opening balance dan koreksi stok, snapshot system quantity, posting atomik ke saldo branch/stock card, dan validasi satu opening balance per item per branch. HPP sementara memakai item cost price.
- Menghubungkan pembayaran POS dengan pengurangan saldo branch, pencatatan stock card, dan HPP keluar berbasis moving average.
- Memindahkan tampilan saldo dari Master Item ke modul Stock.
- Menambahkan dokumentasi Swagger awal dari anotasi Go.
- Menambahkan halaman Swagger UI pada web ERP (`/api-docs`) dengan Try it out.
- Menambahkan master Customer B2B per company dengan kode unik, alamat billing/shipping, status aktif, dan arsip nonaktif.
- Menambahkan MVP Sales Order B2B dengan harga snapshot dari `items.selling_price`, diskon per baris dan dokumen, serta status open/cancelled dan status pengiriman untuk integrasi Goods Delivery berikutnya.
- Menambahkan perubahan Sales Order hanya saat status open; pembatalan tidak mengubah stok. Goods Delivery belum menjadi bagian implementasi ini.

Catat fitur yang selesai, perubahan perilaku, dan batasan yang relevan di bawah entri versi ini.
