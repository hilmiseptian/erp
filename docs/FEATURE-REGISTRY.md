# Feature Registry — ERP

| Fitur | Status | Cakupan / catatan |
|---|---|---|
| Authentication | Tersedia | Token divalidasi oleh API POS Laravel; otorisasi role ERP masih perlu dirancang. |
| API Documentation | Tersedia | Halaman Swagger UI di `/api-docs`; memakai spesifikasi swaggo dan Try it out. |
| Supplier | Tersedia | CRUD per company; schema supplier digunakan bersama POS. |
| Master Item ERP | MVP tersedia | List, search/filter kategori dan status, detail, tambah/edit, aktif/nonaktif. Company-scoped. |
| Saldo item | Read-only di modul Item | Dibaca dari schema POS; tidak diedit dari master item. Pergerakan stok menunggu modul inventory. |
| Category | Referensi tersedia | Modul Item memakai kategori aktif company atau kategori global yang tersedia di database. |
| Purchase Request (PR) | MVP tersedia lokal | Header/detail sesuai skema yang disetujui; memuat tanggal PR dan tanggal dibutuhkan (`req_date`); supplier dan branch per company, harga/kuantitas per item, total header dihitung dari detail; dapat diedit saat open atau dibatalkan. Migration belum dijalankan. |
| Purchase Order (PO) | MVP lokal | Satu PO bersumber dari satu PR melalui `ref_num`; PR dapat dipenuhi parsial, kuantitas tersisa dihitung per item, harga PO boleh berbeda, due date wajib. Header/detail dan migration POS tersedia; migration belum dijalankan. |
| Inventory ledger / stock card | Belum dibuat | Fondasi untuk pergerakan stok per branch dan sinkronisasi settlement POS. |
| Goods Receipt (GR) | Belum dibuat | Penerimaan barang akan memperbarui status penerimaan PO dan saldo inventory. |
| Sales Order / Goods Delivery | Belum dibuat | SO ERP untuk penjualan non-POS. |
| Invoice | Belum dibuat | Scope pembayaran/accounting belum ditentukan. |
| Advanced item data | Rencana | Unit conversion, barcode, lot/batch, serial, expiry, non-stock/service. |

## Batas paket

POS-only tetap memiliki pengelolaan item sederhana. Paket ERP menggunakan master item ERP yang lebih lengkap. Entitlement paket dan sinkronisasi operasional masih perlu dirancang.
