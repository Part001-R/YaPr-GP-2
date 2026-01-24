-- Таблица пользователя.
CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_name TEXT UNIQUE NOT NULL,
  user_password TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Логин/пароль.
CREATE TABLE data1 (  
  field_1 TEXT UNIQUE NOT NULL,
  field_2 TEXT NOT NULL,
	field_3 TEXT NOT NULL,
  created_at DATETIME NOT NULL
);

-- Текст.
CREATE TABLE data2 (  
  field_1 TEXT UNIQUE NOT NULL,
  field_2 TEXT NOT NULL,
  created_at DATETIME NOT NULL
);

-- Банковские карты.
CREATE TABLE data4 (  
  field_1 TEXT UNIQUE NOT NULL,
  field_2 TEXT NOT NULL,
	field_3 TEXT NOT NULL,
	field_4 TEXT NOT NULL,
	field_5 TEXT NOT NULL,
  created_at DATETIME NOT NULL
);

-- Индексы.
CREATE INDEX idx_data1_field_1 ON data1(field_1);
CREATE INDEX idx_data2_field_1 ON data2(field_1);
CREATE INDEX idx_data4_field_1 ON data4(field_1);
