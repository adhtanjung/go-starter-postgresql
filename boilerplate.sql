/*
 Navicat Premium Data Transfer

 Source Server         : boilerplate
 Source Server Type    : MySQL
 Source Server Version : 80031
 Source Host           : localhost:3306
 Source Schema         : casbin

 Target Server Type    : MySQL
 Target Server Version : 80031
 File Encoding         : 65001

 Date: 02/12/2022 17:21:27
*/

GRANT ALL PRIVILEGES ON *.* TO 'user'@'%';
CREATE DATABASE casbin;

SET FOREIGN_KEY_CHECKS = 1;
/*
 Navicat Premium Data Transfer

 Source Server         : boilerplate
 Source Server Type    : MySQL
 Source Server Version : 80031
 Source Host           : localhost:3306
 Source Schema         : boilerplate

 Target Server Type    : MySQL
 Target Server Version : 80031
 File Encoding         : 65001

 Date: 01/12/2022 16:01:44
*/


SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for role
-- ----------------------------
DROP TABLE IF EXISTS `role`;
CREATE TABLE `role` (
  `id` varchar(255) NOT NULL,
  `name` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  FULLTEXT KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of role
-- ----------------------------
BEGIN;
INSERT INTO `role` (`id`, `name`, `created_at`, `updated_at`, `deleted_at`) VALUES ('092be819-19af-447c-93a0-ea99f99f4442', 'superadmin', '2022-11-17 20:49:50', '2022-11-17 20:49:50', NULL);
INSERT INTO `role` (`id`, `name`, `created_at`, `updated_at`, `deleted_at`) VALUES ('69701270-8160-42b0-a2ea-c32ad08056c6', 'user', '2022-11-17 20:51:04', '2022-11-17 20:51:04', NULL);
INSERT INTO `role` (`id`, `name`, `created_at`, `updated_at`, `deleted_at`) VALUES ('c2612106-3425-47bb-bf3b-483bf4e92e34', 'admin', '2022-11-17 20:51:00', '2022-11-17 20:51:00', NULL);
COMMIT;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user` (
  `id` varchar(255) NOT NULL,
  `username` varchar(255) DEFAULT NULL,
  `email` varchar(255) NOT NULL,
  `password` varchar(255) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of user
-- ----------------------------
BEGIN;
INSERT INTO `user` (`id`, `username`, `email`, `password`, `name`, `created_at`, `updated_at`, `deleted_at`) VALUES ('77a3fa90-545c-45ca-b5c4-df0d8216e100', 'anjoex', 'adhtanjung1@gmail.com', '$2a$14$.vn.4WM2u94hDuyFWOr/JuQAa5lziUcTK9A/S2wm6duwiO1y51.3G', 'adhi tanjung', '2022-11-26 16:44:06', '2022-11-26 16:44:06', NULL);
INSERT INTO `user` (`id`, `username`, `email`, `password`, `name`, `created_at`, `updated_at`, `deleted_at`) VALUES ('b98d1f29-17b6-4f2a-b192-f34e92aae404', 'ads', 'asd@mail.com', '$2a$14$.vn.4WM2u94hDuyFWOr/JuQAa5lziUcTK9A/S2wm6duwiO1y51.3G', 'name', '2022-11-23 21:12:24', '2022-11-26 20:11:42', NULL);
COMMIT;

-- ----------------------------
-- Table structure for user_role
-- ----------------------------
DROP TABLE IF EXISTS `user_role`;
CREATE TABLE `user_role` (
  `id` varchar(255) NOT NULL,
  `user_id` varchar(255) NOT NULL,
  `role_id` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL,
  `updated_at` datetime NOT NULL ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `role` (`role_id`),
  KEY `user` (`user_id`),
  CONSTRAINT `role` FOREIGN KEY (`role_id`) REFERENCES `role` (`id`),
  CONSTRAINT `user` FOREIGN KEY (`user_id`) REFERENCES `user` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of user_role
-- ----------------------------
BEGIN;
INSERT INTO `user_role` (`id`, `user_id`, `role_id`, `created_at`, `updated_at`, `deleted_at`) VALUES ('0d9569dd-204a-4185-b1a2-03c0847645a6', 'b98d1f29-17b6-4f2a-b192-f34e92aae404', '092be819-19af-447c-93a0-ea99f99f4442', '2022-11-23 21:12:24', '2022-11-24 17:29:49', NULL);
INSERT INTO `user_role` (`id`, `user_id`, `role_id`, `created_at`, `updated_at`, `deleted_at`) VALUES ('4d8863a5-c765-4569-ad5f-2687a3f94b9a', '77a3fa90-545c-45ca-b5c4-df0d8216e100', '69701270-8160-42b0-a2ea-c32ad08056c6', '2022-11-26 16:44:06', '2022-11-26 16:44:06', NULL);
COMMIT;

SET FOREIGN_KEY_CHECKS = 1;
