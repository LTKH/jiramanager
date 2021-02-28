create table if not exists mon_issues (
  `group_id`      varchar(50) not null,
  `status_id`     varchar(50) default '',
  `status_name`   varchar(100) default '',
  `issue_id`      varchar(50),
  `issue_key`     varchar(50),
  `issue_self`    varchar(1500),
  `created`       bigint(20) default 0,
  `updated`       bigint(20) default 0,
  `template`      varchar(250), 
  unique key IDX_mon_issues_group_id (group_id)
) engine InnoDB default charset=utf8mb4 collate=utf8mb4_unicode_ci;
