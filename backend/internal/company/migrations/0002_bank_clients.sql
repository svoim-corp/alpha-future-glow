-- +goose Up
-- Все пользователи — клиенты банка, признак не нужен (API_SPEC.md, «Пользователи — действующие клиенты банка»).
ALTER TABLE company.companies DROP COLUMN is_bank_client;

-- Компании под каждую ветку «Схемы решения» — backend/SCENARIOS.md.
INSERT INTO company.companies (id, inn, name, type) VALUES
    ('c-2',  '7704211530',   'ООО «Вектор Логистик»',           'LLC'),
    ('c-3',  '772512345601', 'ИП Орлова Мария Игоревна',        'IE'),
    ('c-4',  '7728334410',   'ООО «СтройТехМонтаж»',            'LLC'),
    ('c-5',  '7719456720',   'ООО «Ромашка»',                   'LLC'),
    ('c-6',  '7733567830',   'ООО «ТехноСервис»',               'LLC'),
    ('c-7',  '7716678940',   'ООО «Зелёный сад»',               'LLC'),
    ('c-8',  '7707789050',   'ООО «Альфа-Трейд»',               'LLC'),
    ('c-9',  '7724890160',   'ООО «Мебельная фабрика «Дуб»»',   'LLC'),
    ('c-10', '7731901270',   'ООО «Полюс»',                     'LLC'),
    ('c-11', '773498765432', 'ИП Кузнецов Андрей Павлович',     'IE');

-- +goose Down
DELETE FROM company.companies WHERE id IN ('c-2', 'c-3', 'c-4', 'c-5', 'c-6', 'c-7', 'c-8', 'c-9', 'c-10', 'c-11');
ALTER TABLE company.companies ADD COLUMN is_bank_client boolean NOT NULL DEFAULT false;
