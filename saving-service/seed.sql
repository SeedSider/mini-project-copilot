-- Seed data for BankEase saving-service
-- Run: psql -d saving -f seed.sql

-- Exchange rates seed
INSERT INTO exchange_rate (id, country, currency, country_code, buy, sell) VALUES
  ('a1b2c3d4-0001-4000-8000-000000000001', 'Vietnam',   'VND', 'VN', 1.403,  1.746),
  ('a1b2c3d4-0001-4000-8000-000000000002', 'Nicaragua', 'NIO', 'NI', 9.123,  12.09),
  ('a1b2c3d4-0001-4000-8000-000000000003', 'Korea',     'KRW', 'KR', 3.704,  5.151),
  ('a1b2c3d4-0001-4000-8000-000000000004', 'China',     'CNY', 'CN', 1.725,  2.234)
ON CONFLICT (id) DO NOTHING;

-- Interest rates seed
INSERT INTO interest_rate (id, kind, deposit, rate) VALUES
  ('b2c3d4e5-0002-4000-8000-000000000001', 'individual', '1m',  4.5),
  ('b2c3d4e5-0002-4000-8000-000000000002', 'corporate',  '2m',  5.5),
  ('b2c3d4e5-0002-4000-8000-000000000004', 'corporate',  '6m',  2.5),
  ('b2c3d4e5-0002-4000-8000-000000000011', 'individual', '12m', 5.9)
ON CONFLICT (id) DO NOTHING;

-- Branches seed
INSERT INTO branch (id, name, distance, latitude, longitude) VALUES
  ('c3d4e5f6-0003-4000-8000-000000000001', 'Bank 1656 Union Street',   '50m',    -6.2,   106.816),
  ('c3d4e5f6-0003-4000-8000-000000000002', 'Bank Secaucus',             '1,2 km', -6.205, 106.82),
  ('c3d4e5f6-0003-4000-8000-000000000003', 'Bank 1657 Riverside Drive', '5,3 km', -6.195, 106.825),
  ('c3d4e5f6-0003-4000-8000-000000000004', 'Bank Rutherford',           '70m',    -6.21,  106.812),
  ('c3d4e5f6-0003-4000-8000-000000000005', 'Bank 1656 Union Street',   '30m',    -6.208, 106.814)
ON CONFLICT (id) DO NOTHING;
