CREATE TABLE IF NOT EXISTS news (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
    title VARCHAR(255) NOT NULL COMMENT '新闻标题',
    summary TEXT NOT NULL COMMENT '摘要',
    source VARCHAR(50) NOT NULL COMMENT '来源',
    publish_time DATETIME NOT NULL COMMENT '发布时间',
    lat DECIMAL(10,8) NOT NULL COMMENT '纬度',
    lng DECIMAL(11,8) NOT NULL COMMENT '经度',
    geo_level ENUM('world', 'continent', 'country', 'city') NOT NULL DEFAULT 'country' COMMENT '地理级别',
    is_breaking TINYINT(1) NOT NULL DEFAULT 0 COMMENT '突发标志',
    priority INT NOT NULL DEFAULT 3 COMMENT '优先级(1-5)',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '入库时间',
    INDEX idx_lat_lng (lat, lng),
    INDEX idx_publish_time (publish_time),
    INDEX idx_is_breaking (is_breaking),
    INDEX idx_geo_level (geo_level)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='新闻表';