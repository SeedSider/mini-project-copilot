-- Seed data for BankEase user-profile-service
-- Run: psql -d bankease_db -f seed.sql

-- Profile: 1 default user
INSERT INTO profile (id, bank, branch, name, card_number, card_provider, balance, currency, account_type)
VALUES (
    'da08ecfe-de3b-42b1-b1ce-018e144198f5',
    'Citibank',
    'Tangerang',
    'Jane Doe',
    '12355478990',
    'Mastercard Platinum',
    5000000,
    'IDR',
    'REGULAR'
) ON CONFLICT (id) DO NOTHING;

-- Menu: 9 homepage menu items (mix of REGULAR and PREMIUM)
INSERT INTO menu (id, "index", type, title, icon_url, is_active) VALUES
('menu_001', 1, 'REGULAR', 'Account and Card',
 'https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=id_card',
 TRUE),
('menu_002', 2, 'PREMIUM', 'Transfer',
 'https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=send_money',
 TRUE),
('menu_003', 3, 'REGULAR', 'Payment',
 'https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=payment',
 TRUE),
('menu_004', 4, 'REGULAR', 'Top Up',
 'https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=add_card',
 TRUE),
('menu_005', 5, 'PREMIUM', 'Investment',
 'https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=trending_up',
 TRUE),
('menu_006', 6, 'REGULAR', 'History',
 'https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=history',
 TRUE),
('menu_007', 7, 'PREMIUM', 'Wealth Management',
 'https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=account_balance',
 TRUE),
('menu_008', 8, 'REGULAR', 'Settings',
 'https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=settings',
 TRUE),
('menu_009', 9, 'PREMIUM', 'Priority Services',
 'https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200&icon_names=star',
 TRUE)
ON CONFLICT (id) DO NOTHING;
