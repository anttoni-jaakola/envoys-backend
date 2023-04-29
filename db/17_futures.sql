/*
 Navicat PostgreSQL Data Transfer

 Source Server         : envoys
 Source Server Type    : PostgreSQL
 Source Server Version : 140007 (140007)
 Source Host           : localhost:5432
 Source Catalog        : envoys
 Source Schema         : public

 Target Server Type    : PostgreSQL
 Target Server Version : 140007 (140007)
 File Encoding         : 65001

 Date: 29/04/2023 05:13:44
*/


-- ----------------------------
-- Table structure for futures
-- ----------------------------
DROP TABLE IF EXISTS "public"."futures";
CREATE TABLE "public"."futures" (
  "id" int4 NOT NULL,
  "assignning" varchar(8) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'open'::character varying,
  "position" varchar(8) COLLATE "pg_catalog"."default" NOT NULL DEFAULT 'long'::character varying,
  "trading" varchar(8) COLLATE "pg_catalog"."default",
  "base_unit" varchar(8) COLLATE "pg_catalog"."default",
  "quote_unit" varchar(8) COLLATE "pg_catalog"."default",
  "price" numeric(16,8),
  "quantity" numeric(16,8),
  "take_profit" numeric(16,8),
  "stop_loss" numeric(4,4),
  "status" varchar(8) COLLATE "pg_catalog"."default",
  "create_at" timestamptz(6)
)
;
ALTER TABLE "public"."futures" OWNER TO "envoys";

-- ----------------------------
-- Records of futures
-- ----------------------------
BEGIN;
COMMIT;

-- ----------------------------
-- Primary Key structure for table futures
-- ----------------------------
ALTER TABLE "public"."futures" ADD CONSTRAINT "futures_pkey" PRIMARY KEY ("id");
