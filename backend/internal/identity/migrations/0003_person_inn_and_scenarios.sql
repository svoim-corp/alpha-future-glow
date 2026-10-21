-- +goose Up
-- Связь пользователя с человеком в данных банка (API_SPEC.md, 4.14) и логины под каждую ветку — backend/SCENARIOS.md.
ALTER TABLE identity.users ADD COLUMN person_inn text;

UPDATE identity.users SET person_inn = '771234567890' WHERE id = 'u-1';
UPDATE identity.users SET person_inn = '771298765432' WHERE id = 'u-2';

-- Пароль у всех: demo123 (bcrypt, cost 10).
INSERT INTO identity.users (id, login, password_hash, role, company_id, person_inn, full_name, position) VALUES
    ('u-3',  'volkov',    '$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy', 'CLIENT', 'c-2',  '770411223344', 'Волков Игорь Николаевич',   'Финансовый директор'),
    ('u-4',  'orlova',    '$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy', 'CLIENT', 'c-3',  '772512345601', 'Орлова Мария Игоревна',     'Индивидуальный предприниматель'),
    ('u-5',  'belov',     '$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy', 'CLIENT', 'c-4',  '772855667788', 'Белов Сергей Андреевич',    'Генеральный директор'),
    ('u-6',  'gromov',    '$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy', 'CLIENT', 'c-5',  '771966778899', 'Громов Павел Евгеньевич',   'Генеральный директор'),
    ('u-7',  'zaitsev',   '$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy', 'CLIENT', 'c-6',  '773377889900', 'Зайцев Дмитрий Олегович',   'Генеральный директор'),
    ('u-8',  'sokolova',  '$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy', 'CLIENT', 'c-7',  '771688990011', 'Соколова Елена Викторовна', 'Генеральный директор'),
    ('u-10', 'morozov',   '$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy', 'CLIENT', 'c-8',  '770799001122', 'Морозов Алексей Игоревич',  'Генеральный директор'),
    ('u-11', 'kozlov',    '$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy', 'CLIENT', 'c-9',  '772400112233', 'Козлов Виктор Петрович',    'Генеральный директор'),
    ('u-12', 'polyakov',  '$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy', 'CLIENT', 'c-10', '773111223344', 'Поляков Роман Сергеевич',   'Генеральный директор'),
    ('u-13', 'kuznetsov', '$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy', 'CLIENT', 'c-11', '773498765432', 'Кузнецов Андрей Павлович',  'Индивидуальный предприниматель'),
    ('u-14', 'sidorov',   '$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy', 'BANK_OFFICER', NULL, NULL,  'Сидоров Максим Андреевич',  'Кредитный инспектор');

-- +goose Down
DELETE FROM identity.users WHERE id IN ('u-3', 'u-4', 'u-5', 'u-6', 'u-7', 'u-8', 'u-10', 'u-11', 'u-12', 'u-13', 'u-14');
ALTER TABLE identity.users DROP COLUMN person_inn;
