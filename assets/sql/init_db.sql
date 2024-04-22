-- 创建数据库 kv，如果不存在
CREATE DATABASE IF NOT EXISTS kv;

-- 切换到数据库 kv
USE kv;

-- 创建表 students，如果不存在
CREATE TABLE IF NOT EXISTS `students` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` longtext,
  `score` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_students_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
