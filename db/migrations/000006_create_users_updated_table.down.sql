-- Создаем старую структуру таблицы пользователей
CREATE TABLE IF NOT EXISTS users_old (
    user_id serial PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(50) NOT NULL,
    email VARCHAR(300) UNIQUE NOT NULL
);

-- Перенос данных из новой таблицы
INSERT INTO users_old (username, password, email)
SELECT name, 'default_password', email FROM users;

-- Удаление новой таблицы
DROP TABLE IF EXISTS users;

-- Переименование старой таблицы
ALTER TABLE users_old RENAME TO users; 