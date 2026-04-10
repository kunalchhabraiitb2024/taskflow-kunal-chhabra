DROP INDEX IF EXISTS idx_tasks_created_by;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_assignee_id;
DROP INDEX IF EXISTS idx_tasks_project_id;
DROP TABLE IF EXISTS tasks;
DROP TYPE IF EXISTS task_priority;
DROP TYPE IF EXISTS task_status;
