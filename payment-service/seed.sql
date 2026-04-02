-- Providers (6 items)
INSERT INTO provider (id, name) VALUES
  ('1', 'Biznet'),
  ('2', 'Indihome'),
  ('3', 'MyRepublic'),
  ('4', 'XL Home'),
  ('5', 'CBN'),
  ('6', 'First Media')
ON CONFLICT (id) DO NOTHING;

-- Sample Internet Bill (linked to test user_id)
INSERT INTO internet_bill (user_id, customer_id, name, address, phone_number, code, bill_from, bill_to, internet_fee, tax, total) VALUES
  ('00000000-0000-0000-0000-000000000001', '#2345641ASS', 'Jackson Maine', '403 East 4th Street, Santa Ana', '+8424599721', '#2345641', '01/09/2019', '01/10/2019', '$50', '$0', '$50');

-- Currencies (10 items)
INSERT INTO currency (code, label, rate) VALUES
  ('AUD', 'AUD (Australian Dollar)',        1.53),
  ('CNY', 'CNY (Chinese Yuan)',             7.24),
  ('EUR', 'EUR (Euro)',                     0.92),
  ('GBP', 'GBP (British Pound Sterling)',   0.79),
  ('IDR', 'IDR (Indonesian Rupiah)',        16350),
  ('JPY', 'JPY (Japanese Yen)',             149.5),
  ('MYR', 'MYR (Malaysian Ringgit)',        4.72),
  ('SAR', 'SAR (Saudi Riyal)',              3.75),
  ('SGD', 'SGD (Singapore Dollar)',         1.34),
  ('USD', 'USD (United States Dollar)',     1);

-- Beneficiaries (sample contacts for test account)
INSERT INTO beneficiary (account_id, name, phone, avatar) VALUES
  ('acc-001', 'John Doe', '081234567890', ''),
  ('acc-001', 'Jane Smith', '087654321012', ''),
  ('acc-001', 'Bob Wilson', '082112345678', '')
ON CONFLICT DO NOTHING;

-- Payment Cards (sample cards for test account)
INSERT INTO payment_card (id, account_id, holder_name, card_label, masked_number, balance, currency, brand, gradient_colors) VALUES
  ('card-001', 'acc-001', 'John Doe', 'Primary Card', '**** **** **** 1234', 500000, 'USD', 'VISA', '{"#1a2980","#26d0ce"}'),
  ('card-002', 'acc-001', 'John Doe', 'Business Card', '**** **** **** 5678', 1000000, 'USD', 'MASTERCARD', '{"#eb3349","#f45c43"}')
ON CONFLICT DO NOTHING;
