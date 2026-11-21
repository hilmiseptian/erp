# Feature Registry — ERP

| Fitur | Status | Cakupan / catatan |
|---|---|---|
| Authentication | Tersedia | Token divalidasi oleh API POS Laravel; otorisasi role ERP masih perlu dirancang. |
| API Documentation | Tersedia | Halaman Swagger UI di `/api-docs`; memakai spesifikasi swaggo dan Try it out. |
| Supplier | Tersedia | CRUD per company; schema supplier digunakan bersama POS. |
| Customer | MVP tersedia lokal | Master customer B2B per company; kode unik per company, data pajak/kontak/alamat, aktif/nonaktif. Migration disimpan di repo POS dan belum dijalankan. |
| Master Item ERP | MVP tersedia | List, search/filter kategori dan status, detail, tambah/edit, aktif/nonaktif. Company-scoped. |
| Saldo item | Tersedia di Stock | Saldo terkini ditampilkan per company, branch, dan item; nilai pembukaan disiapkan nol. |
| Category | Referensi tersedia | Modul Item memakai kategori aktif company atau kategori global yang tersedia di database. |
| Purchase Request (PR) | MVP tersedia lokal | Header/detail sesuai skema yang disetujui; memuat tanggal PR dan tanggal dibutuhkan (`req_date`); supplier dan branch per company, harga/kuantitas per item, total header dihitung dari detail; dapat diedit saat open atau dibatalkan. Migration belum dijalankan. |
| Purchase Order (PO) | MVP lokal | Satu PO bersumber dari satu PR melalui `ref_num`; PR dapat dipenuhi parsial, kuantitas tersisa dihitung per item, harga PO boleh berbeda, due date wajib. Header/detail dan migration POS tersedia; migration belum dijalankan. |
| Inventory ledger / stock card | MVP tersedia | Ringkasan saldo per branch dan rincian pergerakan; GR menambah stok pada HPP PO, POS mengurangi stok saat pembayaran pada moving average. |
| Goods Receipt (GR) | MVP tersedia | Penerimaan parsial dari satu PO; harga mengikuti PO, saat diposting saldo branch dan status penerimaan PO diperbarui. |
| Stock Adjustment | MVP tersedia lokal | Opening balance dan koreksi stok memakai header/detail; draft menyimpan snapshot kuantitas sistem, posting menolak snapshot usang, memperbarui `stock_period` dan mencatat non-zero movement ke `stock_card`; satu opening balance dapat diposting per item pada branch. HPP sementara menggunakan `items.cost_price` untuk selisih masuk/keluar sampai aturan akunting dibahas. Migration tersedia di repo POS dan belum dijalankan. |
| Sales Order | MVP tersedia lokal | Sales order B2B per company/branch/customer; harga unit disalin server-side dari selling price item dan tidak dapat diubah langsung, diskon baris/dokumen nominal; `total_amount` adalah jumlah setelah diskon baris dan `grand_total` setelah diskon dokumen; nama/alamat disimpan sebagai snapshot. Tidak melakukan reservasi atau perubahan saldo stok. Goods Delivery posted menghitung qty terkirim serta memperbarui status SO. Migration Customer/SO POS tersedia dan belum dijalankan. |
| Goods Delivery (GD) | MVP tersedia lokal | Posting langsung dan parsial dari satu Sales Order; satu SO dapat dipenuhi beberapa GD. Header memakai `ref_type`/`ref_num`, `branch_id` adalah branch asal. Posting atomik menolak stok kurang, mengurangi saldo branch dan mencatat HPP rata-rata seperti POS; status SO dihitung kembali. Detail hanya menyimpan item dan qty. Migration POS tersedia dan belum dijalankan. Sumber transfer belum aktif sampai modul transfer dibuat. |
| Invoice | Belum dibuat | Scope pembayaran/accounting belum ditentukan. |
| Advanced item data | Rencana | Unit conversion, barcode, lot/batch, serial, expiry, non-stock/service. |

## Batas paket

POS-only tetap memiliki pengelolaan item sederhana. Paket ERP menggunakan master item ERP yang lebih lengkap. Entitlement paket dan sinkronisasi operasional masih perlu dirancang.
