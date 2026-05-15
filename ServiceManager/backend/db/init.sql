-- Active: 1778634876908@@127.0.0.1@3306@mysql
-- Create database
CREATE DATABASE IF NOT EXISTS service_manager;
USE service_manager;

-- Create services table
CREATE TABLE IF NOT EXISTS services (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    status ENUM('running', 'stopped') DEFAULT 'stopped',
    start_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    uptime VARCHAR(50) DEFAULT '',
    url VARCHAR(255) DEFAULT '',
    address VARCHAR(255) DEFAULT '',
    command TEXT DEFAULT '',
    pid INT DEFAULT 0
);

-- Insert test data
INSERT INTO services (name, status, start_time, uptime, url, address, command, pid) VALUES
('Web Service', 'running', NOW(), '2 days 3 hours', 'http://localhost:3000', '127.0.0.1:3000', 'npm start', 1234),
('Database Service', 'running', NOW(), '5 days 1 hour', NULL, '127.0.0.1:3306', 'mysqld', 5678),
('Cache Service', 'stopped', NULL, NULL, NULL, '127.0.0.1:6379', 'redis-server', NULL),
('Message Queue', 'stopped', NULL, NULL, NULL, '127.0.0.1:5672', 'rabbitmq-server', NULL);
