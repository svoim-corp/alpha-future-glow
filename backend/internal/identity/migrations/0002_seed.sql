-- +goose Up
-- Демо-учётки, пароль у всех: demo123 (bcrypt, cost 10).
INSERT INTO identity.users (id, login, password_hash, role, company_id, full_name, position) VALUES
    ('u-1', 'smirnov', '$2a$10$C9/wgrLD./1K8/JJ8sIi5.xHAW42SWN1TOeu9ovHZN8wZtb/K3YDy', 'CLIENT', 'c-1', 'Смирнов Олег Викторович', 'Генеральный директор'),
    ('u-2', 'petrova', '$2a$10$DVLlcdxgBtVIYuPdo4tXweRGUnLQwZtQecz71PKIHv8kPJxOsj4Xa', 'CLIENT', 'c-1', 'Петрова Анна Сергеевна', 'Главный бухгалтер'),
    ('u-9', 'ivanova', '$2a$10$KqoWmwaZl1kOVGcNSv4wCuZc8ZbWQyiAhqLbIBVLgYb//bgKUm89S', 'BANK_OFFICER', NULL, 'Иванова Елена Петровна', 'Кредитный инспектор');

-- +goose Down
DELETE FROM identity.users WHERE id IN ('u-1', 'u-2', 'u-9');
