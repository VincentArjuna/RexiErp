-- RexiERP Seed Data
-- Indonesian Business Scenarios for MSMEs

-- Insert Indonesian geographic data
INSERT INTO countries (id, code, name) VALUES
    ('00000000-0000-0000-0000-000000000001', 'ID', 'Indonesia');

-- Indonesian Provinces
INSERT INTO provinces (id, country_id, code, name) VALUES
    ('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', '31', 'DKI Jakarta'),
    ('10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', '32', 'Jawa Barat'),
    ('10000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', '33', 'Jawa Tengah'),
    ('10000000-0000-0000-0000-000000000004', '00000000-0000-0000-0000-000000000001', '34', 'DI Yogyakarta'),
    ('10000000-0000-0000-0000-000000000005', '00000000-0000-0000-0000-000000000001', '35', 'Jawa Timur'),
    ('10000000-0000-0000-0000-000000000006', '00000000-0000-0000-0000-000000000001', '63', 'Bali'),
    ('10000000-0000-0000-0000-000000000007', '00000000-0000-0000-0000-000000000001', '64', 'Nusa Tenggara Barat'),
    ('10000000-0000-0000-0000-000000000008', '00000000-0000-0000-0000-000000000001', '65', 'Nusa Tenggara Timur'),
    ('10000000-0000-0000-0000-000000000009', '00000000-0000-0000-0000-000000000001', '61', 'Kalimantan Barat'),
    ('10000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000001', '62', 'Kalimantan Tengah'),
    ('10000000-0000-0000-0000-000000000011', '00000000-0000-0000-0000-000000000001', '13', 'Sumatera Barat'),
    ('10000000-0000-0000-0000-000000000012', '00000000-0000-0000-0000-000000000001', '14', 'Riau'),
    ('10000000-0000-0000-0000-000000000013', '00000000-0000-0000-0000-000000000001', '15', 'Jambi'),
    ('10000000-0000-0000-0000-000000000014', '00000000-0000-0000-0000-000000000001', '16', 'Sumatera Selatan'),
    ('10000000-0000-0000-0000-000000000015', '00000000-0000-0000-0000-000000000001', '17', 'Bengkulu'),
    ('10000000-0000-0000-0000-000000000016', '00000000-0000-0000-0000-000000000001', '18', 'Lampung'),
    ('10000000-0000-0000-0000-000000000017', '00000000-0000-0000-0000-000000000001', '19', 'Kepulauan Bangka Belitung'),
    ('10000000-0000-0000-0000-000000000018', '00000000-0000-0000-0000-000000000001', '21', 'Kepulauan Riau'),
    ('10000000-0000-0000-0000-000000000019', '00000000-0000-0000-0000-000000000001', '11', 'Aceh'),
    ('10000000-0000-0000-0000-000000000020', '00000000-0000-0000-0000-000000000001', '12', 'Sumatera Utara'),
    ('10000000-0000-0000-0000-000000000021', '00000000-0000-0000-0000-000000000001', '71', 'Sulawesi Utara'),
    ('10000000-0000-0000-0000-000000000022', '00000000-0000-0000-0000-000000000001', '72', 'Sulawesi Tengah'),
    ('10000000-0000-0000-0000-000000000023', '00000000-0000-0000-0000-000000000001', '73', 'Sulawesi Selatan'),
    ('10000000-0000-0000-0000-000000000024', '00000000-0000-0000-0000-000000000001', '74', 'Sulawesi Tenggara'),
    ('10000000-0000-0000-0000-000000000025', '00000000-0000-0000-0000-000000000001', '75', 'Gorontalo'),
    ('10000000-0000-0000-0000-000000000026', '00000000-0000-0000-0000-000000000001', '76', 'Sulawesi Barat'),
    ('10000000-0000-0000-0000-000000000027', '00000000-0000-0000-0000-000000000001', '51', 'Bali'),
    ('10000000-0000-0000-0000-000000000028', '00000000-0000-0000-0000-000000000001', '52', 'Nusa Tenggara Barat'),
    ('10000000-0000-0000-0000-000000000029', '00000000-0000-0000-0000-000000000001', '53', 'Nusa Tenggara Timur'),
    ('10000000-0000-0000-0000-000000000030', '00000000-0000-0000-0000-000000000001', '64', 'Papua Barat'),
    ('10000000-0000-0000-0000-000000000031', '00000000-0000-0000-0000-000000000001', '91', 'Papua'),
    ('10000000-0000-0000-0000-000000000032', '00000000-0000-0000-0000-000000000001', '81', 'Maluku'),
    ('10000000-0000-0000-0000-000000000033', '00000000-0000-0000-0000-000000000001', '82', 'Maluku Utara'),
    ('10000000-0000-0000-0000-000000000034', '00000000-0000-0000-0000-000000000001', '74', 'Sulawesi Tenggara');

-- Major Indonesian Cities
INSERT INTO cities (id, province_id, code, name, type) VALUES
    -- Jakarta
    ('20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '3171', 'Jakarta Pusat', 'kota'),
    ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', '3172', 'Jakarta Utara', 'kota'),
    ('20000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', '3173', 'Jakarta Barat', 'kota'),
    ('20000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000001', '3174', 'Jakarta Selatan', 'kota'),
    ('20000000-0000-0000-0000-000000000005', '10000000-0000-0000-0000-000000000001', '3175', 'Jakarta Timur', 'kota'),

    -- West Java
    ('20000000-0000-0000-0000-000000000006', '10000000-0000-0000-0000-000000000002', '3273', 'Bandung', 'kota'),
    ('20000000-0000-0000-0000-000000000007', '10000000-0000-0000-0000-000000000002', '3272', 'Bogor', 'kota'),
    ('20000000-0000-0000-0000-000000000008', '10000000-0000-0000-0000-000000000002', '3271', 'Bekasi', 'kota'),
    ('20000000-0000-0000-0000-000000000009', '10000000-0000-0000-0000-000000000002', '3578', 'Depok', 'kota'),
    ('20000000-0000-0000-0000-000000000010', '10000000-0000-0000-0000-000000000002', '3274', 'Cimahi', 'kota'),
    ('20000000-0000-0000-0000-000000000011', '10000000-0000-0000-0000-000000000002', '3275', 'Tasikmalaya', 'kota'),

    -- Central Java
    ('20000000-0000-0000-0000-000000000012', '10000000-0000-0000-0000-000000000003', '3374', 'Semarang', 'kota'),
    ('20000000-0000-0000-0000-000000000013', '10000000-0000-0000-0000-000000000003', '3571', 'Surakarta', 'kota'),
    ('20000000-0000-0000-0000-000000000014', '10000000-0000-0000-0000-000000000003', '3375', 'Pekalongan', 'kota'),
    ('20000000-0000-0000-0000-000000000015', '10000000-0000-0000-0000-000000000003', '3376', 'Tegal', 'kota'),
    ('20000000-0000-0000-0000-000000000016', '10000000-0000-0000-0000-000000000003', '3371', 'Magelang', 'kota'),

    -- East Java
    ('20000000-0000-0000-0000-000000000017', '10000000-0000-0000-0000-000000000005', '3578', 'Surabaya', 'kota'),
    ('20000000-0000-0000-0000-000000000018', '10000000-0000-0000-0000-000000000005', '3579', 'Malang', 'kota'),
    ('20000000-0000-0000-0000-000000000019', '10000000-0000-0000-0000-000000000005', '3573', 'Kediri', 'kota'),
    ('20000000-0000-0000-0000-000000000020', '10000000-0000-0000-0000-000000000005', '3575', 'Blitar', 'kota'),
    ('20000000-0000-0000-0000-000000000021', '10000000-0000-0000-0000-000000000005', '3576', 'Madiun', 'kota'),

    -- Bali
    ('20000000-0000-0000-0000-000000000022', '10000000-0000-0000-0000-000000000006', '5171', 'Denpasar', 'kota'),

    -- North Sumatra
    ('20000000-0000-0000-0000-000000000023', '10000000-0000-0000-0000-000000000020', '1271', 'Medan', 'kota');

-- Sample Tenants (Indonesian MSMEs)
INSERT INTO tenants (id, name, domain, subdomain, company_type, business_category, tax_number, tax_status, email, phone, address, province_id, city_id, postal_code, is_active, subscription_plan, max_users) VALUES
    -- Jakarta-based Retail Business
    ('30000000-0000-0000-0000-000000000001', 'Toko Maju Jaya Abadi', 'majujaya.rexi-erp.local', 'majujaya', 'PT', 'dagang', '01.123.456.7-123.000', 'pkp', 'info@majujaya.com', '021-12345678', 'Jl. Sudirman No. 123, Jakarta Pusat', '10000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', '10110', true, 'premium', 50),

    -- Bandung-based Manufacturing Business
    ('30000000-0000-0000-0000-000000000002', 'CV Sentosa Tekstil', 'sentosa.rexi-erp.local', 'sentosa', 'CV', 'manufaktur', '02.234.567.8-456.000', 'pkp', 'admin@sentosatekstil.com', '022-87654321', 'Jl. Gatot Subroto No. 456, Bandung', '10000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000006', '40281', true, 'premium', 30),

    -- Surabaya-based Service Business
    ('30000000-0000-0000-0000-000000000003', 'PT Karya Mandiri Jasa', 'karyamandiri.rexi-erp.local', 'karyamandiri', 'PT', 'jasa', '03.345.678.9-789.000', 'pkp', 'contact@karyamandiri.id', '031-98765432', 'Jl. Tunjungan No. 789, Surabaya', '10000000-0000-0000-0000-000000000005', '20000000-0000-0000-0000-000000000017', '60261', true, 'business', 25),

    -- Semarang-based Trading Business
    ('30000000-0000-0000-0000-000000000004', 'UD Berkah Dagang', 'berkahdagang.rexi-erp.local', 'berkahdagang', 'UD', 'dagang', NULL, 'non_pkp', 'berkahdagang@email.com', '024-55556666', 'Jl. Pemuda No. 321, Semarang', '10000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000012', '50139', true, 'basic', 10),

    -- Denpasar-based Tourism Business
    ('30000000-0000-0000-0000-000000000005', 'PT Bali Sejahtera Wisata', 'balisejahtera.rexi-erp.local', 'balisejahtera', 'PT', 'jasa', '04.456.789.0-012.000', 'pkp', 'info@balisejahtera.com', '0361-77778888', 'Jl. Raya Kuta No. 999, Denpasar', '10000000-0000-0000-0000-000000000006', '20000000-0000-0000-0000-000000000022', '80361', true, 'premium', 40);

-- Sample Users for each tenant
INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, phone, role, is_active, is_email_verified) VALUES
    -- Toko Maju Jaya Abadi Users
    ('40000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', 'admin@majujaya.com', '$2a$10$rQZ8ZHWKQGYHQkIKJ9K7/.pVvJqYYzGk5v8N6L9w8x8Q9x7x6x5x4', 'Ahmad', 'Sudirman', '0812-1234-5678', 'super_admin', true, true),
    ('40000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', 'manager@majujaya.com', '$2a$10$rQZ8ZHWKQGYHQkIKJ9K7/.pVvJqYYzGk5v8N6L9w8x8Q9x7x6x5x4', 'Siti', 'Nurhaliza', '0812-2345-6789', 'manager', true, true),
    ('40000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000001', 'kasir@majujaya.com', '$2a$10$rQZ8ZHWKQGYHQkIKJ9K7/.pVvJqYYzGk5v8N6L9w8x8Q9x7x6x5x4', 'Budi', 'Santoso', '0812-3456-7890', 'employee', true, true),

    -- CV Sentosa Tekstil Users
    ('40000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000002', 'owner@sentosatekstil.com', '$2a$10$rQZ8ZHWKQGYHQkIKJ9K7/.pVvJqYYzGk5v8N6L9w8x8Q9x7x6x5x4', 'Raden', 'Wijaya', '0822-1234-5678', 'super_admin', true, true),
    ('40000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000002', 'akuntan@sentosatekstil.com', '$2a$10$rQZ8ZHWKQGYHQkIKJ9K7/.pVvJqYYzGk5v8N6L9w8x8Q9x7x6x5x4', 'Dewi', 'Kartika', '0822-2345-6789', 'admin', true, true),

    -- PT Karya Mandiri Jasa Users
    ('40000000-0000-0000-0000-000000000006', '30000000-0000-0000-0000-000000000003', 'director@karyamandiri.id', '$2a$10$rQZ8ZHWKQGYHQkIKJ9K7/.pVvJqYYzGk5v8N6L9w8x8Q9x7x6x5x4', 'Hendra', 'Kusuma', '0831-1234-5678', 'super_admin', true, true),
    ('40000000-0000-0000-0000-000000000007', '30000000-0000-0000-0000-000000000003', 'hr@karyamandiri.id', '$2a$10$rQZ8ZHWKQGYHQkIKJ9K7/.pVvJqYYzGk5v8N6L9w8x8Q9x7x6x5x4', 'Rina', 'Wulandari', '0831-2345-6789', 'admin', true, true),

    -- UD Berkah Dagang Users
    ('40000000-0000-0000-0000-000000000008', '30000000-0000-0000-0000-000000000004', 'pemilik@berkahdagang.com', '$2a$10$rQZ8ZHWKQGYHQkIKJ9K7/.pVvJqYYzGk5v8N6L9w8x8Q9x7x6x5x4', 'Muhammad', 'Hidayat', '0824-1234-5678', 'super_admin', true, true),

    -- PT Bali Sejahtera Wisata Users
    ('40000000-0000-0000-0000-000000000009', '30000000-0000-0000-0000-000000000005', 'ceo@balisejahtera.com', '$2a$10$rQZ8ZHWKQGYHQkIKJ9K7/.pVvJqYYzGk5v8N6L9w8x8Q9x7x6x5x4', 'Made', 'Artana', '0836-1234-5678', 'super_admin', true, true),
    ('40000000-0000-0000-0000-000000000010', '30000000-0000-0000-0000-000000000005', 'finance@balisejahtera.com', '$2a$10$rQZ8ZHWKQGYHQkIKJ9K7/.pVvJqYYzGk5v8N6L9w8x8Q9x7x6x5x4', 'Ni Made', 'Suci', '0836-2345-6789', 'admin', true, true);

-- Standard Chart of Accounts (SAK Indonesia compliant) for each tenant
-- This will be created for each tenant, but here's a sample for the first tenant
INSERT INTO chart_of_accounts (id, tenant_id, code, name, description, account_type, is_active) VALUES
    -- Assets
    ('50000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', '1', 'AKTIVA', 'Total aktiva perusahaan', 'asset', true),
    ('50000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', '11', 'AKTIVA LANCAR', 'Aktiva yang dapat diubah menjadi kas dalam satu tahun', 'asset', true),
    ('50000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000001', '110', 'KAS DAN SETARA KAS', 'Kas, bank, dan investasi jangka pendek', 'asset', true),
    ('50000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000001', '1101', 'KAS', 'Kas tunai dan kas di bank', 'asset', true),
    ('50000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000001', '1102', 'BANK', 'Saldo rekening bank', 'asset', true),
    ('50000000-0000-0000-0000-000000000006', '30000000-0000-0000-0000-000000000001', '111', 'PIUTANG USAHA', 'Piutang dari penjualan kredit', 'asset', true),
    ('50000000-0000-0000-0000-000000000007', '30000000-0000-0000-0000-000000000001', '1111', 'PIUTANG USAHA', 'Piutang dari penjualan barang/jasa', 'asset', true),
    ('50000000-0000-0000-0000-000000000008', '30000000-0000-0000-0000-000000000001', '112', 'PERSEDIAAN', 'Persediaan barang dagang dan bahan baku', 'asset', true),
    ('50000000-0000-0000-0000-000000000009', '30000000-0000-0000-0000-000000000001', '1121', 'BARANG DAGANG', 'Persediaan barang dagangan', 'asset', true),
    ('50000000-0000-0000-0000-000000000010', '30000000-0000-0000-0000-000000000001', '1122', 'BAHAN BAKU', 'Persediaan bahan baku produksi', 'asset', true),

    -- Liabilities
    ('50000000-0000-0000-0000-000000000011', '30000000-0000-0000-0000-000000000001', '2', 'KEWAJIBAN', 'Total kewajiban perusahaan', 'liability', true),
    ('50000000-0000-0000-0000-000000000012', '30000000-0000-0000-0000-000000000001', '21', 'KEWAJIBAN LANCAR', 'Kewajiban jatuh tempo dalam satu tahun', 'liability', true),
    ('50000000-0000-0000-0000-000000000013', '30000000-0000-0000-0000-000000000001', '210', 'HUTANG USAHA', 'Hutang dari pembelian kredit', 'liability', true),
    ('50000000-0000-0000-0000-000000000014', '30000000-0000-0000-0000-000000000001', '2101', 'HUTANG USAHA', 'Hutang kepada pemasok', 'liability', true),
    ('50000000-0000-0000-0000-000000000015', '30000000-0000-0000-0000-000000000001', '211', 'HUTANG PAJAK', 'Hutang pajak yang terutang', 'liability', true),
    ('50000000-0000-0000-0000-000000000016', '30000000-0000-0000-0000-000000000001', '2111', 'PPN KELUARAN', 'PPN yang harus disetor', 'liability', true),
    ('50000000-0000-0000-0000-000000000017', '30000000-0000-0000-0000-000000000001', '2112', 'PAJAK PENGHASILAN', 'PPh pasal 21, 23, 25, 29', 'liability', true),

    -- Equity
    ('50000000-0000-0000-0000-000000000018', '30000000-0000-0000-0000-000000000001', '3', 'EKUITAS', 'Modal pemilik dan laba ditahan', 'equity', true),
    ('50000000-0000-0000-0000-000000000019', '30000000-0000-0000-0000-000000000001', '31', 'MODAL SAHAM', 'Modal disetor pemilik', 'equity', true),
    ('50000000-0000-0000-0000-000000000020', '30000000-0000-0000-0000-000000000001', '311', 'MODAL DISETOR', 'Modal yang disetor pemilik', 'equity', true),
    ('50000000-0000-0000-0000-000000000021', '30000000-0000-0000-0000-000000000001', '32', 'LABA DITAHAN', 'Laba yang belum dibagi', 'equity', true),

    -- Revenue
    ('50000000-0000-0000-0000-000000000022', '30000000-0000-0000-0000-000000000001', '4', 'PENDAPATAN', 'Total pendapatan usaha', 'revenue', true),
    ('50000000-0000-0000-0000-000000000023', '30000000-0000-0000-0000-000000000001', '41', 'PENDAPATAN USAHA', 'Pendapatan dari kegiatan utama', 'revenue', true),
    ('50000000-0000-0000-0000-000000000024', '30000000-0000-0000-0000-000000000001', '410', 'PENJUALAN', 'Pendapatan penjualan barang/jasa', 'revenue', true),
    ('50000000-0000-0000-0000-000000000025', '30000000-0000-0000-0000-000000000001', '4101', 'PENJUALAN BARANG', 'Pendapatan penjualan barang dagang', 'revenue', true),
    ('50000000-0000-0000-0000-000000000026', '30000000-0000-0000-0000-000000000001', '4102', 'PENJUALAN JASA', 'Pendapatan penjualan jasa', 'revenue', true),
    ('50000000-0000-0000-0000-000000000027', '30000000-0000-0000-0000-000000000001', '42', 'PENDAPATAN LAIN-LAIN', 'Pendapatan di luar kegiatan utama', 'revenue', true),

    -- Expenses
    ('50000000-0000-0000-0000-000000000028', '30000000-0000-0000-0000-000000000001', '5', 'BEBAN', 'Total beban usaha', 'expense', true),
    ('50000000-0000-0000-0000-000000000029', '30000000-0000-0000-0000-000000000001', '51', 'HARGA POKOK PENJUALAN', 'Beban langsung terkait penjualan', 'expense', true),
    ('50000000-0000-0000-0000-000000000030', '30000000-0000-0000-0000-000000000001', '511', 'HARGA POKOK PENJUALAN', 'HPP barang terjual', 'expense', true),
    ('50000000-0000-0000-0000-000000000031', '30000000-0000-0000-0000-000000000001', '52', 'BEBAN USAHA', 'Beban operasional perusahaan', 'expense', true),
    ('50000000-0000-0000-0000-000000000032', '30000000-0000-0000-0000-000000000001', '521', 'GAJI DAN UPAH', 'Gaji karyawan dan upah buruh', 'expense', true),
    ('50000000-0000-0000-0000-000000000033', '30000000-0000-0000-0000-000000000001', '522', 'SEWA', 'Beban sewa gedung/kantor', 'expense', true),
    ('50000000-0000-0000-0000-000000000034', '30000000-0000-0000-0000-000000000001', '523', 'LISTRIK DAN AIR', 'Beban utilitas', 'expense', true),
    ('50000000-0000-0000-0000-000000000035', '30000000-0000-0000-0000-000000000001', '524', 'TELEPON DAN INTERNET', 'Beban komunikasi', 'expense', true),
    ('50000000-0000-0000-0000-000000000036', '30000000-0000-0000-0000-000000000001', '525', 'PEMASARAN DAN IKLAN', 'Beban promosi dan iklan', 'expense', true),
    ('50000000-0000-0000-0000-000000000037', '30000000-0000-0000-0000-000000000001', '526', 'DEPRESIASI', 'Beban depresiasi aset tetap', 'expense', true);

-- Sample Product Categories for each tenant
INSERT INTO product_categories (id, tenant_id, name, description, is_active) VALUES
    -- Categories for Toko Maju Jaya Abadi (Retail)
    ('60000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', 'Elektronik', 'Produk elektronik rumah tangga', true),
    ('60000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', 'Pakaian', 'Pakaian pria, wanita, dan anak-anak', true),
    ('60000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000001', 'Makanan', 'Makanan dan minuman kemasan', true),
    ('60000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000001', 'Peralatan Rumah Tangga', 'Peralatan dapur dan kebersihan', true),

    -- Categories for CV Sentosa Tekstil (Manufacturing)
    ('60000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000002', 'Kain Batik', 'Produk kain batik tradisional', true),
    ('60000000-0000-0000-0000-000000000006', '30000000-0000-0000-0000-000000000002', 'Kain Modern', 'Kain motif modern', true),
    ('60000000-0000-0000-0000-000000000007', '30000000-0000-0000-0000-000000000002', 'Bahan Baku', 'Benang dan bahan baku tekstil', true);

-- Sample Products for each tenant
INSERT INTO products (id, tenant_id, sku, name, description, category_id, unit, purchase_price, selling_price, tax_rate, min_stock, is_active) VALUES
    -- Products for Toko Maju Jaya Abadi
    ('70000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', 'ELK-001', 'Televisi LED 32"', 'TV LED 32 inch smart TV', '60000000-0000-0000-0000-000000000001', 'Unit', 2500000, 2999000, 0.11, 5, true),
    ('70000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', 'ELK-002', 'Kulkas 2 Pintu', 'Kulkas 2 pintu kapasitas 200L', '60000000-0000-0000-0000-000000000001', 'Unit', 3500000, 3999000, 0.11, 3, true),
    ('70000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000001', 'PAK-001', 'Kemeja Pria L', 'Kemeja formal pria size L', '60000000-0000-0000-0000-000000000002', 'Unit', 150000, 199000, 0.11, 20, true),
    ('70000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000001', 'MAK-001', 'Mi Instan', 'Mi instan rasa ayam bawang', '60000000-0000-0000-0000-000000000003', 'Dus', 2500, 3500, 0.11, 100, true),

    -- Products for CV Sentosa Tekstil
    ('70000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000002', 'TEX-001', 'Kain Batik Cap', 'Kain batik cap motif klasik', '60000000-0000-0000-0000-000000000005', 'Meter', 85000, 120000, 0.11, 50, true),
    ('70000000-0000-0000-0000-000000000006', '30000000-0000-0000-0000-000000000002', 'TEX-002', 'Kain Batik Tulis', 'Kain batik tulis premium', '60000000-0000-0000-0000-000000000005', 'Meter', 150000, 250000, 0.11, 20, true),
    ('70000000-0000-0000-0000-000000000007', '30000000-0000-0000-0000-000000000002', 'BAH-001', 'Benang Polyester', 'Benang polyester warna putih', '60000000-0000-0000-0000-000000000007', 'Roll', 45000, 65000, 0.11, 100, true);

-- Sample Customers for each tenant
INSERT INTO customers (id, tenant_id, code, name, email, phone, address, city_id, postal_code, tax_number, credit_limit, is_active) VALUES
    -- Customers for Toko Maju Jaya Abadi
    ('80000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', 'CUS-001', 'PT Teknologi Maju', 'order@teknologimaju.com', '021-55554444', 'Jl. Gatot Subroto No. 100, Jakarta Selatan', '20000000-0000-0000-0000-000000000004', '12345', '01.999.888.7-777.000', 50000000, true),
    ('80000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', 'CUS-002', 'CV Anugerah Sentosa', 'info@anugerahsentosa.com', '021-66667777', 'Jl. Thamrin No. 200, Jakarta Pusat', '20000000-0000-0000-0000-000000000001', '10110', NULL, 25000000, true),
    ('80000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000001', 'CUS-003', 'Budi Santoso', 'budi.santoso@email.com', '0813-1111-2222', 'Jl. Sudirman No. 50, Jakarta Pusat', '20000000-0000-0000-0000-000000000001', '10110', NULL, 10000000, true),

-- Sample Suppliers for each tenant
INSERT INTO suppliers (id, tenant_id, code, name, email, phone, address, city_id, postal_code, tax_number, payment_terms, is_active) VALUES
    -- Suppliers for Toko Maju Jaya Abadi
    ('90000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', 'SUP-001', 'PT Elektronik Indonesia', 'sales@elektronikindo.com', '021-88889999', 'Jl. Industri Raya No. 25, Jakarta Utara', '20000000-0000-0000-0000-000000000002', '14140', '01.111.222.3-444.000', 30, true),
    ('90000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', 'SUP-002', 'UD Garmen Jaya', 'order@garmenjaya.com', '022-77776666', 'Jl. Kiaracondong No. 150, Bandung', '20000000-0000-0000-0000-000000000006', '40281', NULL, 14, true),
    ('90000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000001', 'SUP-003', 'PT Food Indonesia', 'marketing@foodindo.com', '021-33334444', 'Jl. Pabrik No. 10, Bekasi', '20000000-0000-0000-0000-000000000008', '17111', '01.555.666.7-888.000', 21, true);

-- Sample Warehouses for each tenant
INSERT INTO warehouses (id, tenant_id, code, name, address, is_active) VALUES
    ('A0000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', 'WH-001', 'Gudang Utama', 'Jl. Gudang No. 1, Jakarta Pusat', true),
    ('A0000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', 'WH-002', 'Gudang Cabang Bandung', 'Jl. Cabang No. 5, Bandung', true),
    ('A0000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000002', 'WH-001', 'Gudang Pusat', 'Jl. Pabrik No. 25, Bandung', true);

-- Initial stock levels
INSERT INTO inventory_stocks (id, tenant_id, product_id, warehouse_id, quantity) VALUES
    -- Stock for Toko Maju Jaya Abadi
    ('B0000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000001', 'A0000000-0000-0000-0000-000000000001', 15),
    ('B0000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000002', 'A0000000-0000-0000-0000-000000000001', 8),
    ('B0000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000003', 'A0000000-0000-0000-0000-000000000001', 50),
    ('B0000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000001', '70000000-0000-0000-0000-000000000004', 'A0000000-0000-0000-0000-000000000001', 200),

    -- Stock for CV Sentosa Tekstil
    ('B0000000-0000-0000-0000-000000000005', '30000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000005', 'A0000000-0000-0000-0000-000000000003', 150),
    ('B0000000-0000-0000-0000-000000000006', '30000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000006', 'A0000000-0000-0000-0000-000000000003', 75),
    ('B0000000-0000-0000-0000-000000000007', '30000000-0000-0000-0000-000000000002', '70000000-0000-0000-0000-000000000007', 'A0000000-0000-0000-0000-000000000003', 200);

-- Create indexes for performance if they don't exist
CREATE INDEX IF NOT EXISTS idx_chart_of_accounts_tenant_code ON chart_of_accounts(tenant_id, code);
CREATE INDEX IF NOT EXISTS idx_product_categories_tenant_name ON product_categories(tenant_id, name);
CREATE INDEX IF NOT EXISTS idx_products_tenant_sku_active ON products(tenant_id, sku, is_active);
CREATE INDEX IF NOT EXISTS idx_customers_tenant_code_active ON customers(tenant_id, code, is_active);
CREATE INDEX IF NOT EXISTS idx_suppliers_tenant_code_active ON suppliers(tenant_id, code, is_active);
CREATE INDEX IF NOT EXISTS idx_warehouses_tenant_code ON warehouses(tenant_id, code);

-- Insert some sample transactions for testing
-- This would typically be done through the application but included here for testing purposes
COMMIT;