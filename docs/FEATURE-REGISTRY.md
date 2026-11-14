# Feature Registry — ERP

| Fitur | Status | Cakupan / catatan |
|---|---|---|
| Authentication | Tersedia | Token divalidasi oleh API POS Laravel; otorisasi role ERP masih perlu dirancang. |
| Supplier | Tersedia | CRUD per company; schema supplier digunakan bersama POS. |
| Master Item ERP | MVP tersedia | List, search/filter kategori dan status, detail, tambah/edit, aktif/nonaktif. Company-scoped. |
| Saldo item | Read-only di modul Item | Dibaca dari schema POS; tidak diedit dari master item. Pergerakan stok menunggu modul inventory. |
| Category | Referensi tersedia | Modul Item memakai kategori aktif company atau kategori global yang tersedia di database. |
| Inventory ledger / stock card | Belum dibuat | Fondasi untuk pergerakan stok per branch dan sinkronisasi settlement POS. |
| Purchasing (PR/PO/GR) | Belum dibuat | PO dapat menggabungkan beberapa PR; approval fase berikutnya. |
| Sales Order / Goods Delivery | Belum dibuat | SO ERP untuk penjualan non-POS. |
| Invoice | Belum dibuat | Scope pembayaran/accounting belum ditentukan. |
| Advanced item data | Rencana | Unit conversion, barcode, lot/batch, serial, expiry, non-stock/service. |

## Batas paket

POS-only tetap memiliki pengelolaan item sederhana. Paket ERP menggunakan master item ERP yang lebih lengkap. Entitlement paket dan sinkronisasi operasional masih perlu dirancang.
