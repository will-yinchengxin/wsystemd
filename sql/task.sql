CREATE DATABASE `wsystemd` IF NOT EXISTS;
USE `wsystemd`;

DROP TABLE IF EXISTS `task`;
CREATE TABLE `task` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `job_id` varchar(64) NOT NULL COMMENT '任务ID',
  `node` varchar(64) NOT NULL COMMENT '节点名称',
  `pid` int(11) NOT NULL DEFAULT '0' COMMENT '进程ID',
  `cmd` varchar(255) NOT NULL COMMENT '执行命令',
  `args` text COMMENT '命令参数',
  `outfile` varchar(255) DEFAULT NULL COMMENT '标准输出文件',
  `errfile` varchar(255) DEFAULT NULL COMMENT '错误输出文件',
  `dc` varchar(64) DEFAULT NULL COMMENT '数据中心',
  `ip` varchar(32) DEFAULT NULL COMMENT 'IP地址',
  `load_method` varchar(32) DEFAULT NULL ,
  `do_once` tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否一次性任务: 0-否 1-是',
  `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '任务状态: 0-停止 1-运行中 2-失败',
  `retry_count` int(11) NOT NULL DEFAULT '0' COMMENT '重试次数',
  `last_error` text COMMENT '最后一次错误信息',
  `heart_beat_time` datetime DEFAULT NULL COMMENT '心跳时间',
  `create_time` datetime DEFAULT NULL COMMENT '创建时间',
  `update_time` datetime DEFAULT NULL  COMMENT '更新时间',
  `big_one` varchar(255) NOT NULL DEFAULT '',
  `type` varchar(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_job_id` (`job_id`),
  KEY `idx_node_pid` (`node`, `pid`),
  KEY `idx_heart_beat` (`heart_beat_time`),
  KEY `idx_node_status` (`node`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;