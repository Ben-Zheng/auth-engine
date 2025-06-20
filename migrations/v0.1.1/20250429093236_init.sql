-- Create "token_validity_policies" table
CREATE TABLE `token_validity_policies` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT "ID",
  `workspace_id` varchar(64) NOT NULL COMMENT "工作空间ID",
  `token_id` varchar(64) NOT NULL COMMENT "token ID",
  `policy_type` enum('DAILY','WEEKLY','DATERANGE') NOT NULL COMMENT "策略类型",
  `start_time` datetime(3) NULL COMMENT "开始时间",
  `end_time` datetime(3) NULL COMMENT "结束时间",
  `start_day` enum('MONDAY','TUESDAY','WEDNESDAY','THURSDAY','FRIDAY','SATURDAY','SUNDAY','') NULL COMMENT "开始日",
  `end_day` enum('MONDAY','TUESDAY','WEDNESDAY','THURSDAY','FRIDAY','SATURDAY','SUNDAY','') NULL COMMENT "结束日",
  `start_date` date NULL COMMENT "开始日期",
  `end_date` date NULL COMMENT "结束日期",
  `create_time` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT "创建时间",
  `update_time` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT "更新时间",
  PRIMARY KEY (`id`),
  INDEX `idx_token_id` (`token_id`),
  INDEX `idx_workspace_id` (`workspace_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
-- Create "tokens" table
CREATE TABLE `tokens` (
  `create_time` datetime NULL DEFAULT CURRENT_TIMESTAMP COMMENT "创建时间",
  `update_time` datetime NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT "更新时间",
  `del_time` datetime NULL DEFAULT "9999-12-31 00:00:00" COMMENT "逻辑删除时间",
  `update_by` varchar(36) NULL COMMENT "更新人",
  `create_by` varchar(36) NULL COMMENT "创建人",
  `del_flag` int NULL DEFAULT 0 COMMENT "逻辑删除标志【0 ： 未删除 1： 已删除】",
  `id` varchar(64) NOT NULL COMMENT "ID",
  `workspace_id` varchar(64) NOT NULL COMMENT "工作空间ID",
  `token` varchar(255) NOT NULL COMMENT "token",
  `expired_time` datetime NULL COMMENT "过期时间",
  `app_scenario_name` varchar(255) NOT NULL COMMENT "应用场景名称",
  `model_name` varchar(255) NOT NULL COMMENT "模型名称",
  `env_name` varchar(255) NOT NULL COMMENT "环境名称",
  `enable_validity_policy` bool NOT NULL COMMENT "是否启用有效期策略",
  `policy_type` enum('DAILY','WEEKLY','DATERANGE','') NULL COMMENT "策略类型",
  `max_concurrency` bigint NOT NULL COMMENT "最高并发量",
  PRIMARY KEY (`id`),
  UNIQUE INDEX `idx_app_model_env_name` (`app_scenario_name`, `model_name`, `env_name`),
  UNIQUE INDEX `idx_token` (`token`),
  INDEX `idx_workspace_id` (`workspace_id`)
) CHARSET utf8mb4 COLLATE utf8mb4_general_ci;
