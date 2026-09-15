-- ============================================================================
-- dtm_barrier.sql — DTM 所需的数据库准备（任务 1.2）
--
-- 本文件包含两段，须分别执行：
--   第一段：对 postgres 实例执行（任意管理库，如默认库）——创建 dtm 独立存储库，
--           dtm server 启动时自动迁移其存储表（dtm.trans_global 等）。
--   第二段：对业务库 mall 执行——创建 dtm_barrier.barrier 分支幂等表
--           （源自 dtm v1.19.0 sqls/dtmcli.barrier.postgres.sql），
--           供 order/inventory/coupons 三服务的 Saga 分支事务内使用。
-- ============================================================================

-- ---------------------------------------------------------------------------
-- 第一段：dtm 独立存储库（对 postgres 实例执行）
-- ---------------------------------------------------------------------------
CREATE DATABASE dtm;

-- ---------------------------------------------------------------------------
-- 第二段：分支幂等表（对业务库 mall 执行）
-- ---------------------------------------------------------------------------
create schema if not exists dtm_barrier;
drop table if exists dtm_barrier.barrier;
CREATE SEQUENCE if not EXISTS dtm_barrier.barrier_seq;
create table if not exists dtm_barrier.barrier(
  id bigint NOT NULL DEFAULT NEXTVAL ('dtm_barrier.barrier_seq'),
  trans_type varchar(45) default '',
  gid varchar(128) default '',
  branch_id varchar(128) default '',
  op varchar(45) default '',
  barrier_id varchar(45) default '',
  reason varchar(45) default '',
  create_time timestamp(0) with time zone DEFAULT NULL,
  update_time timestamp(0) with time zone DEFAULT NULL,
  PRIMARY KEY(id),
  CONSTRAINT uniq_barrier unique(gid, branch_id, op, barrier_id)
);
