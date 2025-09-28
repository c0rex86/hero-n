-- добавляем verifier для SRP аутентификации
ALTER TABLE users ADD COLUMN verifier BLOB;
