# PRD Awal: Purchasing dan Inventory ERP

**Status:** Draft untuk diskusi
**Tanggal:** 13 November 2026
**Cakupan:** Perencanaan produk; belum merupakan spesifikasi implementasi.

## 1. Tujuan

Menyediakan alur pengadaan, pemenuhan pesanan, dan pencatatan persediaan ERP yang dapat bekerja bersama POS. Dokumen ini mencatat keadaan repo yang diperiksa, usulan kemampuan tambahan, serta keputusan yang masih perlu dikonfirmasi. Item yang disebut “perlu” di bawah adalah rekomendasi analisis, bukan keputusan final.

## 2. Temuan dari repo saat ini

### ERP (`C:\Apache24\htdocs\erp`)

- API Go saat ini menyediakan autentikasi dan CRUD supplier. Query supplier dipisahkan menurut `company_id`.
- Web memiliki tampilan supplier. Menu Inventory dan Purchasing di sidebar masih nonaktif/“Soon”.
- Belum ditemukan model, route, atau repository ERP untuk PR, PO, SO, goods receipt/delivery, invoice, stock list, atau stock card.
- Repo ERP menghubungkan token ke API POS Laravel (`/api/user`) dan mengambil `company_id`.

### POS (`C:\Apache24\htdocs\laract-api`, `C:\Apache24\htdocs\laract-fe`)

- API POS memiliki modul Item, Order/POS, Shift, dan Reports; frontend POS memiliki area items, orders, dashboard, dan laporan.
- Model item POS memiliki `stock`, `min_stock`, `cost_price`, `selling_price`, dan `unit`.
- POS membuat penjualan melalui `sales_head`, `sales_detail`, dan `sales_payment`; order mendukung item, pembayaran, dan pembatalan.
- Migrasi supplier POS mendefinisikan tabel yang dipakai ERP. Ini adalah bukti kontrak schema pada repo, tetapi belum membuktikan kedua aplikasi memakai database aktif yang sama.
- Tidak ditemukan ledger pergerakan stok/stock card dalam daftar model/migrasi POS yang diperiksa. Keberadaan field `items.stock` saja tidak membuktikan mekanisme pengurangan stok; perilaku aktual harus dikonfirmasi dari alur order POS dan database deployment.
- POS mengelola company dan branch; migrasi penjualan menambahkan `company_id` dan `branch_id` ke header sales.

## 3. Modul yang diminta

Daftar awal dari pemilik produk:

1. Purchase Request (PR)
2. Purchase Order (PO)
3. Sales Order (SO)
4. Goods Receipt (penerimaan barang)
5. Goods Delivery (pengiriman barang)
6. Purchase Invoice
7. Sales Invoice
8. Stock list / saldo persediaan
9. Stock card / ringkasan dan riwayat stok per item
10. Pemotongan stok dari transaksi POS

## 4. Rekomendasi kemampuan tambahan

### Inti agar alur operasional lengkap

- **Supplier dan customer:** Supplier sudah mulai tersedia di ERP. Master customer belum tampak di ERP; POS memiliki customer belum diketahui dari bukti yang diperiksa. SO, delivery, dan sales invoice perlu pihak customer atau aturan lain yang eksplisit.
- **Master gudang/lokasi stok:** Branch POS bukan otomatis gudang. Tentukan apakah saldo dikelola per company, per branch, per gudang, atau kombinasi.
- **Stock adjustment dan stock opname:** Dibutuhkan untuk koreksi stok dengan alasan, otorisasi, dan jejak audit; jangan mengubah saldo tanpa transaksi yang menjelaskan sebabnya.
- **Transfer antar gudang/branch:** Diperlukan jika stok dapat berpindah lokasi. Umumnya perlu status dikirim dan diterima agar barang dalam perjalanan dapat dilacak.
- **Retur pembelian dan retur penjualan:** Menangani barang kembali dan koreksi stok/dokumen asal. Bisa menjadi fase setelah alur normal, tetapi perlu diputuskan sebelum desain status final.
- **Pembayaran / hutang / piutang:** Invoice saja tidak mengelola pelunasan. Tentukan apakah ERP juga mencatat pembayaran, termin, saldo jatuh tempo, dan pembatalan/kredit memo atau hanya menyimpan referensi invoice.
- **Otorisasi dan status dokumen:** Draft, diajukan, disetujui/ditolak, diproses sebagian, selesai, dibatalkan; aturan siapa boleh melakukan transisi harus ditentukan.
- **Nomor dokumen dan audit trail:** Nomor unik per perusahaan (dan mungkin branch/periode), pencatat waktu/pengguna, perubahan status, alasan, dan referensi silang dokumen.

### Pengendalian stok yang disarankan

- PR **tidak** menambah stok fisik. PO yang disetujui juga tidak menambah stok fisik; bila diperlukan, tampilkan sebagai stok on-order/expected.
- Goods receipt menambah stok fisik sesuai kuantitas yang benar-benar diterima, dapat parsial, dan merujuk ke PO/baris PO.
- SO umumnya mengalokasikan/reservasi stok, bukan langsung mengurangi stok fisik. Tentukan kebijakan reservasi dan pelepasannya saat dibatalkan.
- Goods delivery mengurangi stok fisik sesuai kuantitas yang dikirim, dapat parsial, dan merujuk ke SO/baris SO.
- Transaksi POS yang telah berhasil/final mengurangi stok di lokasi POS yang dipilih tepat satu kali. Tentukan dampak pembatalan, refund, dan retur.
- Stock card idealnya menjadi tampilan ledger yang dapat ditelusuri ke transaksi sumber; saldo stok adalah hasil konsisten dari saldo awal ditambah/kurang seluruh pergerakan yang sah.

Aturan ini adalah rekomendasi konsep. Metode costing, reservasi, dan waktu posting belum diputuskan.

## 5. Alur tingkat tinggi yang diusulkan

```text
PR -> approval -> PO -> penerimaan parsial/penuh (Goods Receipt) -> Purchase Invoice -> pembayaran (opsional)
SO -> reservasi (opsional) -> Goods Delivery -> Sales Invoice -> pembayaran (opsional)
POS sale final -> stock movement keluar (lokasi POS)
Adjustment / opname / transfer / retur -> stock movement dengan referensi dan audit
```

Dokumen-dokumen tidak harus selalu berantai penuh. Contohnya, kebijakan dapat mengizinkan receipt tanpa PO atau sales invoice langsung dari POS. Kebijakan tersebut perlu diputuskan.

## 6. Prinsip produk awal

- Semua data bisnis dibatasi menurut `company_id`, konsisten dengan pola repo.
- Kuantitas dan nilai dokumen harus mendukung pemrosesan parsial bila proses nyata memerlukannya.
- Pergerakan stok harus dapat ditelusuri ke sumber, lokasi, item, kuantitas, waktu, dan pengguna.
- Dokumen yang sudah diposting tidak diedit diam-diam; koreksi menggunakan pembatalan/reversal atau dokumen koreksi yang dapat diaudit.
- Stock list harus membedakan stok fisik, reserved, available, on-order, dan incoming jika kebutuhan dan proses disepakati.
- Integrasi ERP–POS membutuhkan kontrak bersama untuk item/SKU, company, branch/lokasi, status transaksi POS, dan event pengurangan stok.

## 7. Pembagian tanggung jawab ERP dan POS: usulan untuk diskusi

- **POS:** transaksi kasir cepat, shift, pembayaran saat checkout, dan pencatatan penjualan pada branch.
- **ERP:** pengadaan, persetujuan, penerimaan, pemenuhan pesanan, invoice operasional, pengelolaan lokasi stok, penyesuaian, dan analisis stock ledger.
- **Bersama:** identitas company, item/SKU, branch/lokasi, serta kontrak integrasi penjualan-ke-stok.

Ini usulan pembagian, belum keputusan. Hindari dua aplikasi sama-sama mengubah saldo stok tanpa satu mekanisme posting yang disepakati dan idempotent.

## 8. Pertanyaan terbuka untuk keputusan bersama

1. Apakah ERP dan POS memang berbagi satu database MySQL pada deployment, atau hanya bertukar API? Aplikasi mana yang menjadi sumber utama untuk master item dan saldo stok?
2. Apakah inventory dicatat per branch, per gudang, atau per lokasi stok yang dapat berbeda dari branch POS? Apakah satu branch boleh memiliki beberapa gudang?
3. Apakah POS saat ini benar-benar mengurangi `items.stock` saat pembayaran/finalisasi? Bagaimana pembatalan, refund, dan retur diproses?
4. Apakah POS menggunakan `sales_head` sebagai Sales Order juga, atau ERP memerlukan SO terpisah? Apakah SO hanya untuk penjualan non-POS?
5. Apakah PR memerlukan approval? Siapa pembuat dan approver, dan apakah PO boleh dibuat dari beberapa PR?
6. Apakah receipt tanpa PO dan delivery tanpa SO diizinkan? Apakah over-receipt/over-delivery boleh dengan toleransi?
7. Apakah invoice hanya dokumen referensi, atau ERP juga mengelola pembayaran, hutang/piutang, pajak, diskon, mata uang, dan credit note?
8. Metode nilai persediaan apa yang diperlukan (misalnya average/FIFO), dan apakah biaya tambahan pembelian masuk ke nilai stok?
9. Apakah item mendukung unit konversi, lot/batch, serial number, tanggal kedaluwarsa, barang non-stok, dan jasa?
10. Apakah retur, transfer, opname, adjustment, dan approval masuk rilis pertama atau fase lanjutan?
11. Siapa yang boleh melihat/mengubah data lintas branch dan dokumen lintas company? Bagaimana role/permission ERP mengikuti POS?
12. Apakah item master yang POS pakai akan tetap dikelola di POS, atau pindah menjadi master ERP dengan sinkronisasi ke POS?

## 9. Batasan dan langkah berikutnya

Dokumen ini belum menetapkan desain tabel, endpoint, UI, acceptance criteria final, atau urutan implementasi. Sebelum spesifikasi teknis, jawab pertanyaan terbuka terutama sumber kebenaran stok, lokasi, integrasi transaksi POS, dan cakupan invoice/pembayaran. Setelah itu PRD dapat dipecah menjadi fase, requirement per modul, aturan status, dan kriteria penerimaan.

Pemeriksaan repo dilakukan secara read-only. Tidak ada kode, migrasi, atau database yang dibuat/diubah dalam penyusunan PRD ini.
