create table if not exists mon_tasks (
  `group_id`      varchar(50) not null,
  `status_id`     varchar(50),
  `status_name`   varchar(100),
  `task_id`       varchar(50),
  `task_key`      varchar(50),
  `task_self`     varchar(1500),
  `starts_at`     bigint(20) default 0,
  `ends_at`       bigint(20) default 0,
  unique key IDX_mon_tasks_group_id (group_id)
) engine InnoDB default charset=utf8mb4 collate=utf8mb4_unicode_ci;