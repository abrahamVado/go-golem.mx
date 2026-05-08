DROP TRIGGER IF EXISTS trg_org_owner_membership_after_insert ON companies;
DROP TRIGGER IF EXISTS trg_tasks_validate_column_board_before_update ON tasks;
DROP TRIGGER IF EXISTS trg_tasks_validate_column_board_before_insert ON tasks;
DROP FUNCTION IF EXISTS validate_tasks_column_board();
