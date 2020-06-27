create table if not exists mon_tasks (
  `group_id`      varchar(50) not null,
  `status`        varchar(10),
  `starts_at`     bigint(20) default 0,
  `ends_at`       bigint(20) default 0,
  `task_id`       varchar(50),
  `task_key`      varchar(50),
  `task_self`     varchar(1500),
  unique key IDX_mon_tasks_group_id (group_id)
) engine InnoDB default charset=utf8mb4 collate=utf8mb4_unicode_ci;