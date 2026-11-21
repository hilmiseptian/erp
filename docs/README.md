# ERP Documentation

Dokumentasi kerja ERP selama tahap pengembangan dan pengujian manual. Semua perubahan pra-production memperbarui dokumentasi versi 1; belum membuat versi dokumentasi baru.

## Isi

- [Changelog](CHANGELOG.md): ringkasan perubahan fitur terakhir.
- [Feature Registry](FEATURE-REGISTRY.md): status dan cakupan fitur ERP/POS.
- [PRD Supplier](prd/PRD-SUPPLIER.md) dan [PRD Auth](prd/PRD-AUTH.md).
- [Swagger API](api/swagger.yaml): spesifikasi OpenAPI yang digunakan halaman Swagger UI; selaraskan dengan anotasi handler Go saat kontrak API berubah.
- Halaman Swagger UI tersedia pada route web ERP `/api-docs` setelah login.
- PR tersedia dari Inventory di sidebar dan memakai item aktif serta branch milik company.
- Stock dan Good Receipts tersedia sebagai child Inventory; saldo dibuka nol per branch dan pergerakan penjualan POS mengurangi saldo saat pembayaran.
- Customer dan Sales Orders B2B tersedia di grup Sales pada sidebar. SO belum memotong atau memesan stok; Goods Delivery akan menjadi transaksi yang memproses pengurangan stok pada tahap berikutnya.
- Kuantitas transaksi PR, PO, GR, dan POS menggunakan bilangan bulat. Migration disimpan di repo POS dan belum dijalankan.

Perbarui CHANGELOG dan FEATURE-REGISTRY saat fitur selesai. Perbarui PRD/API ketika aturan atau kontrak berubah.
