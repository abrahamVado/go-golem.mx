CREATE OR REPLACE FUNCTION validate_tasks_column_board()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.column_id IS NOT NULL AND NEW.board_id IS NULL THEN
        RAISE EXCEPTION 'tasks.column_id requires tasks.board_id';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_tasks_validate_column_board_before_insert ON tasks;
CREATE TRIGGER trg_tasks_validate_column_board_before_insert
BEFORE INSERT ON tasks
FOR EACH ROW
EXECUTE FUNCTION validate_tasks_column_board();

DROP TRIGGER IF EXISTS trg_tasks_validate_column_board_before_update ON tasks;
CREATE TRIGGER trg_tasks_validate_column_board_before_update
BEFORE UPDATE ON tasks
FOR EACH ROW
EXECUTE FUNCTION validate_tasks_column_board();
