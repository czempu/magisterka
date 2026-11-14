CREATE TABLE `users` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `username` varchar(50) UNIQUE NOT NULL,
  `email` varchar(100) UNIQUE NOT NULL,
  `password` varchar(255) NOT NULL,
  `is_employee` boolean NOT NULL,
  `country_code` char(2) NOT NULL,
  `nominal_day_minutes` int NOT NULL,
  `timezone` varchar(64) NOT NULL,
  `created_at` timestamp,
  `updated_at` timestamp
);

CREATE TABLE `module` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `module_name` varchar(40) NOT NULL,
  `description` varchar(256) NOT NULL
);

CREATE TABLE `roles` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `slug` varchar(64) UNIQUE NOT NULL,
  `name` varchar(128) NOT NULL
);

CREATE TABLE `permissions` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `resource` varchar(64) NOT NULL,
  `action` varchar(32) NOT NULL,
  `scope` ENUM ('own', 'team', 'org') NOT NULL,
  `conditions` json,
  `cond_sig` char(64)
);

CREATE TABLE `role_permissions` (
  `role_id` int NOT NULL,
  `permission_id` int NOT NULL,
  `Primary` Key(role_id,permission_id)
);

CREATE TABLE `user_roles` (
  `user_id` int NOT NULL,
  `role_id` int NOT NULL,
  `team_id` int,
  `Primary` Key(user_id,role_id)
);

CREATE TABLE `audit_logs` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `actor_id` int NOT NULL,
  `resource` varchar(64) NOT NULL,
  `action` varchar(32) NOT NULL,
  `target_id` varchar(64),
  `before_doc` json,
  `after_doc` json,
  `at` timestamp
);

CREATE TABLE `timesheet_entries` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `started_at` datetime NOT NULL,
  `ended_at` datetime,
  `type` ENUM ('work', 'break', 'overtime') NOT NULL,
  `note` varchar(255),
  `created_at` timestamp,
  `updated_at` timestamp
);

CREATE TABLE `work_calendars` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `name` varchar(128) NOT NULL,
  `timezone` varchar(128) NOT NULL,
  `is_default` boolean NOT NULL
);

CREATE TABLE `flex_policies` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `name` text NOT NULL,
  `day_hours` numeric(4,2) NOT NULL,
  `window_start` time NOT NULL,
  `window_end` time NOT NULL,
  `rounding_minutes` int NOT NULL
);

CREATE TABLE `employee_assignments` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `employee_id` int NOT NULL,
  `work_calendar_id` int NOT NULL,
  `flex_policy_id` int NOT NULL,
  `effective_from` date NOT NULL,
  `effective_to` date
);


CREATE TABLE `calendar_holidays` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `work_calendar_id` int NOT NULL,
  `day` date NOT NULL,
  `name` varchar(120) NOT NULL
);

CREATE TABLE `leave_types` (
  `code` varchar(8) PRIMARY KEY,
  `name` varchar(120) NOT NULL,
  `affects_balance` boolean NOT NULL,
  `count_strategy` ENUM ('business', 'calendar') NOT NULL,
  `allow_days` boolean NOT NULL,
  `allow_hours` boolean NOT NULL,
  `allow_minutes` boolean NOT NULL,
  `min_inc_days` int NOT NULL,
  `min_inc_hours` int NOT NULL,
  `min_inc_minutes` int NOT NULL,
  `nominal_day_minutes` int
);

CREATE TABLE `leave_requests` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `employee_id` int NOT NULL,
  `type_code` varchar(8) NOT NULL,
  `created_at` datetime NOT NULL
);

CREATE TABLE `leave_request_items` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `request_id` bigint NOT NULL,
  `start_at` datetime(6) NOT NULL,
  `end_at` datetime(6) NOT NULL,
  `unit` ENUM ('day', 'hour', 'minute') NOT NULL,
  `quantity` decimal(8,2) NOT NULL,
  `effective_minutes` int NOT NULL,
  `note` text
);

CREATE TABLE `user_schedules_weekly` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `effective_from` date NOT NULL,
  `effective_to` date,
  `dow` tinyint NOT NULL,
  `planned_minutes` int NOT NULL
);

CREATE TABLE `user_schedule_overrides` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `day` date NOT NULL,
  `planned_minutes` int NOT NULL
);

CREATE TABLE `time_off_adjustments` (
  `id` int PRIMARY KEY,
  `employee_id` int NOT NULL,
  `leave_type_code` varchar(8) NOT NULL,
  `effective_date` date NOT NULL,
  `delta_hours` numeric(8,2) NOT NULL,
  `reason` text,
  `created_at` timestamp
);


CREATE TABLE `org_units` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `name` varchar(160) NOT NULL,
  `parent_id` int,
  `code` varchar(64)
);

CREATE TABLE `user_org_memberships` (
  `user_id` int NOT NULL,
  `org_unit_id` int NOT NULL,
  `is_primary` boolean NOT NULL,
  `title` varchar(160),
  `since_date` date,
  `until_date` date,
  `Primary` Key(user_id,org_unit_id)
);

CREATE TABLE `org_unit_managers` (
  `org_unit_id` int NOT NULL,
  `user_id` int NOT NULL,
  `since_date` date,
  `until_date` date,
  `is_default` boolean NOT NULL,
  `Primary` Key(org_unit_id,user_id)
);

CREATE TABLE `user_line_managers` (
  `user_id` int NOT NULL,
  `manager_id` int NOT NULL,
  `since_date` date,
  `until_date` date,
  `Primary` Key(user_id,manager_id)
);

CREATE TABLE `approval_policies` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `resource` varchar(64) NOT NULL,
  `conditions` json,
  `workflow_id` int NOT NULL
);

CREATE TABLE `approval_workflows` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `name` varchar(160) NOT NULL,
  `resource` varchar(64) NOT NULL
);

CREATE TABLE `approval_workflow_steps` (
  `workflow_id` int NOT NULL,
  `step_no` int NOT NULL,
  `completion_rule` varchar(8) NOT NULL,
  `approver_type` ENUM ('manager_of_requester', 'manager_of_unit', 'role_in_unit', 'specific_user', 'role_global') NOT NULL,
  `role_slug` varchar(64),
  `levels_up` int,
  `org_scope` ENUM ('requester_unit', 'unit_chain_up', 'whole_org'),
  `required` boolean NOT NULL,
  `Primary` Key(workflow_id,step_no)
);

CREATE TABLE `approval_requests` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `resource` varchar(64) NOT NULL,
  `target_id` bigint NOT NULL,
  `requester_id` int NOT NULL,
  `status` ENUM ('pending', 'approved', 'rejected', 'cancelled') NOT NULL,
  `created_at` datetime NOT NULL,
  `closed_at` datetime
);

CREATE TABLE `approval_request_steps` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `request_id` bigint NOT NULL,
  `step_no` int NOT NULL,
  `approver_id` int NOT NULL,
  `status` ENUM ('pending', 'approved', 'rejected', 'skipped') NOT NULL,
  `decided_at` datetime,
  `note` text
);

CREATE TABLE `approval_targets_leave` (
  `request_id` bigint PRIMARY KEY,
  `leave_id` bigint NOT NULL
);

CREATE TABLE `org_unit_role_holders` (
  `org_unit_id` int NOT NULL,
  `role_slug` varchar(64) NOT NULL,
  `user_id` int NOT NULL,
  `since_date` date,
  `until_date` date,
  `Primary` Key(org_unit_id,role_slug,user_id)
);

CREATE TABLE `employment_terms` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `effective_from` date NOT NULL,
  `effective_to` date,
  `fte_percent` numeric(5,2) NOT NULL,
  `nominal_day_minutes` int,
  `pay_group` varchar(64),
  `position` varchar(160),
  `cost_center` varchar(64)
);

CREATE TABLE `time_clock_events` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `occurred_at` datetime(6) NOT NULL,
  `direction` varchar(8) NOT NULL,
  `source` varchar(32),
  `device_id` varchar(64),
  `geo_lat` decimal(9,6),
  `geo_lon` decimal(9,6),
  `note` text
);

CREATE TABLE `break_policies` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `name` varchar(120) NOT NULL,
  `min_shift_minutes` int NOT NULL,
  `break_minutes` int NOT NULL,
  `unpaid` boolean NOT NULL
);

CREATE TABLE `rounding_policies` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `name` varchar(120) NOT NULL,
  `start_round_to` int NOT NULL,
  `end_round_to` int NOT NULL,
  `grace_minutes` int NOT NULL
);

CREATE TABLE `user_time_policies` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `effective_from` date NOT NULL,
  `effective_to` date,
  `break_policy_id` int,
  `rounding_policy_id` int
);

CREATE TABLE `overtime_policies` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `name` varchar(120) NOT NULL,
  `day_threshold_minutes` int NOT NULL,
  `week_threshold_minutes` int,
  `rate1_multiplier` numeric(4,2),
  `rate2_multiplier` numeric(4,2)
);

CREATE TABLE `projects` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `code` varchar(64) UNIQUE NOT NULL,
  `name` varchar(160) NOT NULL,
  `active` boolean NOT NULL
);

CREATE TABLE `tasks` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `project_id` int NOT NULL,
  `code` varchar(64) NOT NULL,
  `name` varchar(160) NOT NULL,
  `active` boolean NOT NULL
);

CREATE TABLE `timesheet_daily_totals` (
  `user_id` int NOT NULL,
  `day` date NOT NULL,
  `worked_minutes` int NOT NULL,
  `break_minutes` int NOT NULL,
  `overtime_minutes` int NOT NULL,
  `absence_minutes` int NOT NULL,
  `PRIMARY` KEY(user_id,day)
);

CREATE TABLE `timesheet_monthly_totals` (
  `user_id` int NOT NULL,
  `ym` char(7) NOT NULL,
  `worked_minutes` int NOT NULL,
  `overtime_minutes` int NOT NULL,
  `absence_minutes` int NOT NULL,
  `PRIMARY` KEY(user_id,ym)
);

CREATE TABLE `leave_balance_snapshots` (
  `user_id` int NOT NULL,
  `type_code` varchar(8) NOT NULL,
  `as_of_date` date NOT NULL,
  `balance_hours` numeric(10,2) NOT NULL,
  `PRIMARY` KEY(user_id,type_code,as_of_date)
);

CREATE UNIQUE INDEX `uq_perm` ON `permissions` (`resource`, `action`, `scope`, `cond_sig`);

CREATE INDEX `ix_perm_triplet` ON `permissions` (`resource`, `action`, `scope`);

CREATE INDEX `ix_ur_user` ON `user_roles` (`user_id`);

CREATE INDEX `ix_audit_actor` ON `audit_logs` (`actor_id`);

CREATE INDEX `ix_audit_at` ON `audit_logs` (`at`);

CREATE INDEX `idx_user_start` ON `timesheet_entries` (`user_id`, `started_at`);

CREATE INDEX `idx_emp_assign` ON `employee_assignments` (`employee_id`, `effective_from`, `effective_to`);

CREATE UNIQUE INDEX `uq_cal_day` ON `calendar_holidays` (`work_calendar_id`, `day`);

CREATE INDEX `idx_lr_emp` ON `leave_requests` (`employee_id`, `status`, `created_at`);

CREATE INDEX `idx_li_req` ON `leave_request_items` (`request_id`);

CREATE INDEX `idx_li_req_time` ON `leave_request_items` (`request_id`, `start_at`, `end_at`);

CREATE INDEX `idx_usw_user_rng` ON `user_schedules_weekly` (`user_id`, `effective_from`, `effective_to`, `dow`);

CREATE UNIQUE INDEX `uniq_user_day` ON `user_schedule_overrides` (`user_id`, `day`);

CREATE INDEX `org_units_index_13` ON `org_units` (`parent_id`);

CREATE UNIQUE INDEX `org_units_index_14` ON `org_units` (`code`);

CREATE INDEX `user_org_memberships_index_15` ON `user_org_memberships` (`org_unit_id`);

CREATE INDEX `approval_policies_index_16` ON `approval_policies` (`resource`);

CREATE UNIQUE INDEX `uq_resource_target` ON `approval_requests` (`resource`, `target_id`);

CREATE INDEX `ix_ar_requester` ON `approval_requests` (`requester_id`);

CREATE INDEX `ix_ar_status` ON `approval_requests` (`status`);

CREATE UNIQUE INDEX `uq_req_step_approver` ON `approval_request_steps` (`request_id`, `step_no`, `approver_id`);

CREATE INDEX `ix_ars_req_step` ON `approval_request_steps` (`request_id`, `step_no`);

CREATE INDEX `ix_ars_approver` ON `approval_request_steps` (`approver_id`);

CREATE INDEX `ix_terms_rng` ON `employment_terms` (`user_id`, `effective_from`, `effective_to`);

CREATE INDEX `ix_tce_user_time` ON `time_clock_events` (`user_id`, `occurred_at`);

CREATE INDEX `ix_utp_rng` ON `user_time_policies` (`user_id`, `effective_from`, `effective_to`);

CREATE UNIQUE INDEX `tasks_index_26` ON `tasks` (`project_id`, `code`);

ALTER TABLE `role_permissions` ADD FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`);

ALTER TABLE `role_permissions` ADD FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`);

ALTER TABLE `user_roles` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `user_roles` ADD FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`);

ALTER TABLE `audit_logs` ADD FOREIGN KEY (`actor_id`) REFERENCES `users` (`id`);

ALTER TABLE `timesheet_entries` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `employee_assignments` ADD FOREIGN KEY (`employee_id`) REFERENCES `users` (`id`);

ALTER TABLE `employee_assignments` ADD FOREIGN KEY (`work_calendar_id`) REFERENCES `work_calendars` (`id`);

ALTER TABLE `employee_assignments` ADD FOREIGN KEY (`flex_policy_id`) REFERENCES `flex_policies` (`id`);

ALTER TABLE `calendar_holidays` ADD FOREIGN KEY (`work_calendar_id`) REFERENCES `work_calendars` (`id`);

ALTER TABLE `leave_requests` ADD FOREIGN KEY (`employee_id`) REFERENCES `users` (`id`);

ALTER TABLE `leave_requests` ADD FOREIGN KEY (`type_code`) REFERENCES `leave_types` (`code`);

ALTER TABLE `leave_request_items` ADD FOREIGN KEY (`request_id`) REFERENCES `leave_requests` (`id`);

ALTER TABLE `user_schedules_weekly` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `user_schedule_overrides` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `time_off_adjustments` ADD FOREIGN KEY (`employee_id`) REFERENCES `users` (`id`);

ALTER TABLE `time_off_adjustments` ADD FOREIGN KEY (`leave_type_code`) REFERENCES `leave_types` (`code`);

ALTER TABLE `org_units` ADD FOREIGN KEY (`parent_id`) REFERENCES `org_units` (`id`);

ALTER TABLE `user_org_memberships` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `user_org_memberships` ADD FOREIGN KEY (`org_unit_id`) REFERENCES `org_units` (`id`);

ALTER TABLE `org_unit_managers` ADD FOREIGN KEY (`org_unit_id`) REFERENCES `org_units` (`id`);

ALTER TABLE `org_unit_managers` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `user_line_managers` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `user_line_managers` ADD FOREIGN KEY (`manager_id`) REFERENCES `users` (`id`);

ALTER TABLE `approval_policies` ADD FOREIGN KEY (`workflow_id`) REFERENCES `approval_workflows` (`id`);

ALTER TABLE `approval_workflow_steps` ADD FOREIGN KEY (`workflow_id`) REFERENCES `approval_workflows` (`id`);

ALTER TABLE `approval_requests` ADD FOREIGN KEY (`requester_id`) REFERENCES `users` (`id`);

ALTER TABLE `approval_request_steps` ADD FOREIGN KEY (`request_id`) REFERENCES `approval_requests` (`id`);

ALTER TABLE `approval_request_steps` ADD FOREIGN KEY (`approver_id`) REFERENCES `users` (`id`);

ALTER TABLE `approval_targets_leave` ADD FOREIGN KEY (`request_id`) REFERENCES `approval_requests` (`id`);

ALTER TABLE `approval_targets_leave` ADD FOREIGN KEY (`leave_id`) REFERENCES `leave_requests` (`id`);

ALTER TABLE `org_unit_role_holders` ADD FOREIGN KEY (`org_unit_id`) REFERENCES `org_units` (`id`);

ALTER TABLE `org_unit_role_holders` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `employment_terms` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `time_clock_events` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `user_time_policies` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `user_time_policies` ADD FOREIGN KEY (`break_policy_id`) REFERENCES `break_policies` (`id`);

ALTER TABLE `user_time_policies` ADD FOREIGN KEY (`rounding_policy_id`) REFERENCES `rounding_policies` (`id`);

ALTER TABLE `tasks` ADD FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`);

ALTER TABLE `timesheet_daily_totals` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `timesheet_monthly_totals` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `leave_balance_snapshots` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

ALTER TABLE `leave_balance_snapshots` ADD FOREIGN KEY (`type_code`) REFERENCES `leave_types` (`code`);






CREATE OR REPLACE VIEW v_user_permissions AS
SELECT
  ur.user_id,
  p.resource,
  p.action,
  p.scope,
  p.conditions
FROM user_roles ur
JOIN role_permissions rp ON rp.role_id = ur.role_id
JOIN permissions p       ON p.id = rp.permission_id;




INSERT INTO roles (slug, name) VALUES
 ('employee','Employee'),
 ('supervisor','Supervisor'),
 ('hr','HR'),
 ('payroll','Payroll'),
 ('auditor','Auditor'),
 ('admin','Admin');




INSERT INTO permissions (resource, action, scope, conditions) VALUES
 -- employee
 ('time_entry','read','own',   JSON_OBJECT()),
 ('time_entry','create','own', JSON_OBJECT()),
 ('time_entry','update','own', JSON_OBJECT('max_days_back', 7, 'forbid_if_approved', true)),
 ('timesheet','read','own',    JSON_OBJECT()),

 -- supervisor
 ('time_entry','read','team',   JSON_OBJECT()),
 ('time_entry','update','team', JSON_OBJECT('max_days_back', 30)),
 ('approvals','approve','team', JSON_OBJECT()),
 ('timesheet','read','team',    JSON_OBJECT()),
 ('reports','read','team',      JSON_OBJECT()),

 -- hr / payroll
 ('time_entry','read','org',    JSON_OBJECT()),
 ('time_entry','update','org',  JSON_OBJECT('soft_correction', true)),
 ('timesheet','read','org',     JSON_OBJECT()),
 ('reports','read','org',       JSON_OBJECT()),
 ('payroll','export','org',     JSON_OBJECT()),

 -- auditor
 ('*','read','org',             JSON_OBJECT()),

 -- admin (systemowy)
 ('users','*','org',            JSON_OBJECT()),
 ('settings','*','org',         JSON_OBJECT()),
 ('projects','*','org',         JSON_OBJECT());
 

 
 -- employee
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='employee' AND p.resource='time_entry' AND p.action='read'  AND p.scope='own';
 
 
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='employee' AND p.resource='time_entry' AND p.action='create' AND p.scope='own';


INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='employee' AND p.resource='time_entry' AND p.action='update' AND p.scope='own';


INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='employee' AND p.resource='timesheet'  AND p.action='read'   AND p.scope='own';



-- supervisor
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='supervisor' AND p.resource='time_entry' AND p.action='read' AND p.scope='team';


INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='supervisor' AND p.resource='time_entry' AND p.action='update' AND p.scope='team';


INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='supervisor' AND p.resource='approvals' AND p.action='approve' AND p.scope='team';


INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='supervisor' AND p.resource='timesheet' AND p.action='read' AND p.scope='team';


INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='supervisor' AND p.resource='reports' AND p.action='read' AND p.scope='team';

-- hr
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='hr' AND p.resource='time_entry' AND p.action='read' AND p.scope='org';


INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='hr' AND p.resource='time_entry' AND p.action='update' AND p.scope='org';


INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='hr' AND p.resource='timesheet' AND p.action='read' AND p.scope='org';


INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='hr' AND p.resource='reports' AND p.action='read' AND p.scope='org';



-- payroll
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='payroll' AND p.resource='reports' AND p.action='read' AND p.scope='org';


INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='payroll' AND p.resource='payroll' AND p.action='export' AND p.scope='org';

-- auditor
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='auditor' AND p.resource='*' AND p.action='read' AND p.scope='org';

-- admin
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='admin' AND p.resource='users' AND p.action='*' AND p.scope='org';


INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='admin' AND p.resource='settings' AND p.action='*' AND p.scope='org';


INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
  ON r.slug='admin' AND p.resource='projects' AND p.action='*' AND p.scope='org';


 
 INSERT INTO roles (slug, name) VALUES ('superadmin','Super Admin');

-- permission: wszystko (resource='*', action='*', scope='org')
INSERT INTO permissions (resource, action, scope, conditions)
VALUES ('*','*','org', JSON_OBJECT());

-- mapowanie roli -> perm (po slugach)
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p
WHERE r.slug='superadmin' AND p.resource='*' AND p.action='*' AND p.scope='org';
 
 

INSERT INTO time_off_adjustments(id,employee_id, leave_type_code, effective_date, delta_hours, reason)
VALUES (1,1, 'WYP', '2025-01-01', 160, 'Roczny przydział');














































