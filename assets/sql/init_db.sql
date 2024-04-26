-- 创建数据库 kv，如果不存在
CREATE DATABASE IF NOT EXISTS kv;

-- 切换到数据库 kv
USE kv;

-- 删除已存在的 students 表，如果存在
DROP TABLE IF EXISTS `students`;

-- 创建表 students，如果不存在
CREATE TABLE IF NOT EXISTS `students` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` VARCHAR(255) NOT NULL,
  `score` VARCHAR(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY (`name`),
  KEY `idx_students_name_score` (`name`(255), `score`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
