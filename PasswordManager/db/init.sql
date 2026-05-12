-- Create database
CREATE DATABASE IF NOT EXISTS password_manager;
USE password_manager;

-- Password entries table
CREATE TABLE IF NOT EXISTS password_entries (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    username VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    remark TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Operation logs table
CREATE TABLE IF NOT EXISTS operation_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    entry_id INT NOT NULL,
    operation_type ENUM('add', 'update', 'delete') NOT NULL,
    operation_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_id INT NOT NULL,
    FOREIGN KEY (entry_id) REFERENCES password_entries(id) ON DELETE CASCADE
);

-- Users table (for authentication)
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert initial admin user
INSERT INTO users (username, password_hash) 
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MrqK3a7M4l4W8Xj5J2lJ3v8bQ1Q1W5y'); -- password: admin123
