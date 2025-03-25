-- Создаем новую таблицу пользователей с более структурированными полями
CREATE TABLE IF NOT EXISTS users_new (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Перенос данных из старой таблицы, если она существует
DO $$
BEGIN
    IF EXISTS (SELECT FROM pg_tables WHERE schemaname = 'public' AND tablename = 'users') THEN
        INSERT INTO users_new (name, email)
        SELECT username, email FROM users;
    END IF;
END
$$;

-- Удаление старой таблицы
DROP TABLE IF EXISTS users;

-- Переименование новой таблицы
ALTER TABLE users_new RENAME TO users;

-- Создание индексов
CREATE INDEX idx_users_email ON users(email); 