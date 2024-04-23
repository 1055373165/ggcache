-- 创建数据库 kv，如果不存在
CREATE DATABASE IF NOT EXISTS kv;

-- 切换到数据库 kv
USE kv;

-- 删除已存在的 students 表，如果存在
DROP TABLE IF EXISTS `students`;