-- Seed data for BankEase user-profile-service
-- Run: psql -d bankease_db -f seed.sql

-- Profile: 1 default user
INSERT INTO profile (id, bank, branch, name, card_number, card_provider, balance, currency, account_type, image)
VALUES (
    'da08ecfe-de3b-42b1-b1ce-018e144198f5',
    'Citibank',
    'Tangerang',
    'Jane Doe',
    '12355478990',
    'Mastercard Platinum',
    5000000,
    'IDR',
    'REGULAR',
    'https://plnsa.blob.core.windows.net/images/309639b1a516241527289b081036c93c.png?sv=2025-11-05&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2026-04-11T14:56:36Z&st=2026-03-26T06:41:36Z&spr=https&sig=uqbKbqKFLDkoZX%2FxLosVhIardyUGotShuDGtHWv9yjE%3D'
) ON CONFLICT (id) DO NOTHING;

-- Menu: 9 homepage menu items (mix of REGULAR and PREMIUM)
INSERT INTO menu (id, "index", type, title, icon_url, is_active) VALUES
('menu_001', 1, 'REGULAR', 'Account and Card',
 'https://plnsa.blob.core.windows.net/images/23308e7b3129d24e3384d33c8b37e196.svg?sv=2025-11-05&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2026-04-11T14:56:36Z&st=2026-03-26T06:41:36Z&spr=https&sig=uqbKbqKFLDkoZX%2FxLosVhIardyUGotShuDGtHWv9yjE%3',
 TRUE),
('menu_002', 2, 'REGULAR', 'Transfer',
 'https://plnsa.blob.core.windows.net/images/d85fd46b6a83f281b4bd99ec428fd2f2.svg?sv=2025-11-05&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2026-04-11T14:56:36Z&st=2026-03-26T06:41:36Z&spr=https&sig=uqbKbqKFLDkoZX%2FxLosVhIardyUGotShuDGtHWv9yjE%3D',
 TRUE),
('menu_003', 3, 'REGULAR', 'Withdraw',
 'https://plnsa.blob.core.windows.net/images/5a24431e56c0ed1daddf3f3187319d5e.svg?sv=2025-11-05&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2026-04-11T14:56:36Z&st=2026-03-26T06:41:36Z&spr=https&sig=uqbKbqKFLDkoZX%2FxLosVhIardyUGotShuDGtHWv9yjE%3D',
 TRUE),
('menu_004', 4, 'REGULAR', 'Mobile Prepaid',
 'https://plnsa.blob.core.windows.net/images/3d6f6d6f9fe8b73bed28db90c5d4e7d0.svg?sv=2025-11-05&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2026-04-11T14:56:36Z&st=2026-03-26T06:41:36Z&spr=https&sig=uqbKbqKFLDkoZX%2FxLosVhIardyUGotShuDGtHWv9yjE%3D',
 TRUE),
('menu_005', 5, 'REGULAR', 'Pay the Bill',
 'https://plnsa.blob.core.windows.net/images/ae42f9d33cf00c0d844f3d278d04cc63.svg?sv=2025-11-05&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2026-04-11T14:56:36Z&st=2026-03-26T06:41:36Z&spr=https&sig=uqbKbqKFLDkoZX%2FxLosVhIardyUGotShuDGtHWv9yjE%3D',
 TRUE),
('menu_006', 6, 'PREMIUM', 'Save online',
 'https://plnsa.blob.core.windows.net/images/bfb59834d141d0b5a4c1565b747d69b1.svg?sv=2025-11-05&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2026-04-11T14:56:36Z&st=2026-03-26T06:41:36Z&spr=https&sig=uqbKbqKFLDkoZX%2FxLosVhIardyUGotShuDGtHWv9yjE%3D',
 TRUE),
('menu_007', 7, 'PREMIUM', 'Credit card',
 'https://plnsa.blob.core.windows.net/images/e42f2d90c5509846e243daa0185c3484.svg?sv=2025-11-05&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2026-04-11T14:56:36Z&st=2026-03-26T06:41:36Z&spr=https&sig=uqbKbqKFLDkoZX%2FxLosVhIardyUGotShuDGtHWv9yjE%3D',
 TRUE),
('menu_008', 8, 'REGULAR', 'Transaction report',
 'https://plnsa.blob.core.windows.net/images/c46d8c3708e5b7cc90e3f4551ac6a924.svg?sv=2025-11-05&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2026-04-11T14:56:36Z&st=2026-03-26T06:41:36Z&spr=https&sig=uqbKbqKFLDkoZX%2FxLosVhIardyUGotShuDGtHWv9yjE%3D',
 TRUE),
('menu_009', 9, 'PREMIUM', 'Beneficiary',
 'https://plnsa.blob.core.windows.net/images/7a158efb85c46668ba04c8c6bb7f6e5a.svg?sv=2025-11-05&ss=bfqt&srt=sco&sp=rwdlacupiytfx&se=2026-04-11T14:56:36Z&st=2026-03-26T06:41:36Z&spr=https&sig=uqbKbqKFLDkoZX%2FxLosVhIardyUGotShuDGtHWv9yjE%3D',
 TRUE)
ON CONFLICT (id) DO NOTHING;
