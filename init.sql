-- 初始化脚本（可选）
-- 这个脚本会在PostgreSQL容器启动时自动执行

-- 创建必要的扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建users表（如果GORM自动迁移不完全）
-- 注意：通常GORM会自动创建表，所以这个可能不需要

-- 你可以在这里添加：
-- 1. 创建数据库用户和权限
-- 2. 初始化基础数据
-- 3. 创建索引
-- 等等
