-- ============================================================
-- Go API Starter 数据库初始化脚本
-- 说明: 此文件用于初始化数据库
--       GORM AutoMigrate 也会自动同步表结构
-- ============================================================

CREATE DATABASE IF NOT EXISTS `go_api_starter`
    DEFAULT CHARACTER SET utf8mb4
    DEFAULT COLLATE utf8mb4_unicode_ci;

USE `go_api_starter`;

-- 设置时区
SET time_zone = '+08:00';
