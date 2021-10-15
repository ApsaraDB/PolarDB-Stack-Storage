/*
SQLyog Community v13.1.6 (64 bit)
MySQL - 8.0.22
*********************************************************************
*/
create DATABASE IF NOT EXISTS `polardb_sms` ;

use polardb_sms;

drop table IF EXISTS `physical_volume` ;

create TABLE IF NOT EXISTS `physical_volume` (
  `id` int NOT NULL AUTO_INCREMENT,
  `volume_id` varchar(45) NOT NULL DEFAULT '“”',
  `vendor` varchar(45) DEFAULT NULL,
  `size` bigint DEFAULT NULL,
  `sector_size` int DEFAULT NULL,
  `sector_num` bigint DEFAULT NULL,
  `fs_type` varchar(45) DEFAULT NULL,
  `paths`             longtext,
  `node_id`           varchar(45)  DEFAULT NULL,
  `cluster_id`        int          NOT NULL DEFAULT '0',
  `pr_support_status` longtext,
  `desc`              varchar(45)  DEFAULT NULL,
  `created`           datetime     DEFAULT NULL,
  `updated`           datetime     DEFAULT NULL,
  `status`            varchar(255) DEFAULT NULL,
  `deleted_at`        datetime     DEFAULT NULL,
  `volume_name`       varchar(45)  DEFAULT NULL,
  `fs_size`           bigint       DEFAULT NULL,
  `used_size`         bigint       DEFAULT NULL,
  `extend`            mediumtext,
  `path_num`          int          DEFAULT NULL,
  `node_ip`           varchar(45)  DEFAULT NULL,
  `product`           varchar(45)  DEFAULT NULL,
  `pv_type`           varchar(45)  DEFAULT NULL,
  `used_by_type`      int          DEFAULT NULL,
  `used_by_name`      varchar(45)  DEFAULT NULL,
  `serial_number`      varchar(45)  DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `wwid_UNIQUE` (`volume_id`, `node_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;

drop table IF EXISTS `logical_volume` ;

create TABLE IF NOT EXISTS `logical_volume`(
  `id`          int         NOT NULL AUTO_INCREMENT,
  `volume_name` varchar(45) NOT NULL,
  `vendor`      varchar(45) ,
  `product`      varchar(45) ,
  `children`    text,
  `lv_type`     varchar(45) DEFAULT NULL,
  `size`        bigint   DEFAULT NULL,
  `sector_size` int      DEFAULT NULL,
  `sector_num`  bigint   DEFAULT NULL,
  `fs_type`     varchar(45) DEFAULT NULL,
  `node_ids`    text,
  `created`     datetime DEFAULT NULL,
  `updated`     datetime DEFAULT NULL,
  `status`      varchar(255) DEFAULT NULL,
  `deleted_at`  datetime DEFAULT NULL,
  `extend` mediumtext,
  `fs_size` bigint DEFAULT NULL,
  `pr_status` mediumtext,
  `pr_node_id` varchar(45) DEFAULT NULL,
  `used_size` bigint DEFAULT NULL,
  `related_pvc` varchar(45) DEFAULT NULL,
  `volume_id` varchar(45) DEFAULT NULL,
  `cluster_id` int DEFAULT NULL,
  `used_by_type` int DEFAULT NULL,
  `used_by_name` varchar(45) DEFAULT NULL,
  `serial_number`      varchar(45)  DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name_UNIQUE` (`volume_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;

drop table IF EXISTS `workflow` ;

create TABLE IF NOT EXISTS `workflow` (
  `id` int NOT NULL AUTO_INCREMENT,
  `type` int NOT NULL,
  `workflow_id` varchar(45) NOT NULL,
  `step` int DEFAULT NULL,
  `created` datetime DEFAULT NULL,
  `deleted` datetime DEFAULT NULL,
  `updated` datetime DEFAULT NULL,
  `mode` int DEFAULT NULL,
  `status` int DEFAULT NULL,
  `last_err_msg` varchar(1024) DEFAULT NULL,
  `stages` json DEFAULT NULL,
  `trace_context`    text,
  `volume_id` varchar(45) DEFAULT NULL,
  `volume_class` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `workflowId_UNIQUE` (`workflow_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;


drop table IF EXISTS `cluster_agent` ;

create TABLE IF NOT EXISTS `cluster_agent`(
  `id`         int         NOT NULL AUTO_INCREMENT,
  `agent_id`   varchar(45) DEFAULT NULL,
  `ip`         varchar(45) DEFAULT NULL,
  `port`       varchar(45) DEFAULT NULL,
  `online`     tinyint     DEFAULT NULL,
  `created`    datetime    DEFAULT NULL,
  `updated`    datetime    DEFAULT NULL,
  `cluster_id` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `agent_id_UNIQUE` (`agent_id`, `cluster_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;

drop table IF EXISTS `idempotent` ;
create TABLE IF NOT EXISTS `idempotent` (
  `id`            int         NOT NULL AUTO_INCREMENT,
  `source`        varchar(45) DEFAULT NULL,
  `idempotent_id` varchar(45) DEFAULT NULL,
  `workflow_id`   varchar(45) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;

drop table IF EXISTS `pvc` ;
create TABLE IF NOT EXISTS `polardb_sms`.`pvc` (
  `id` INT NOT NULL AUTO_INCREMENT,
  `pvc_name` VARCHAR(45) NULL,
  `pvc_namespace` VARCHAR(45) NULL,
  `pvc_status` VARCHAR(255) NULL,
  `volume_class` VARCHAR(45) NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `pvc_name_UNIQUE` (`pvc_name` ASC, `pvc_namespace` ASC)
  ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;