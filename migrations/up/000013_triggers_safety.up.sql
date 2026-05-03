DELIMITER $$

CREATE TRIGGER trg_tasks_validate_column_board_before_insert
BEFORE INSERT ON tasks
FOR EACH ROW
BEGIN
    IF NEW.column_id IS NOT NULL AND NEW.board_id IS NULL THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'tasks.column_id requires tasks.board_id';
    END IF;
END$$

CREATE TRIGGER trg_tasks_validate_column_board_before_update
BEFORE UPDATE ON tasks
FOR EACH ROW
BEGIN
    IF NEW.column_id IS NOT NULL AND NEW.board_id IS NULL THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT = 'tasks.column_id requires tasks.board_id';
    END IF;
END$$

CREATE TRIGGER trg_org_owner_membership_after_insert
AFTER INSERT ON organizations
FOR EACH ROW
BEGIN
    INSERT IGNORE INTO organization_members (
        organization_id, user_id, membership_type, status, joined_at, created_at, updated_at
    ) VALUES (
        NEW.id, NEW.owner_user_id, 'owner', 'active', NOW(), NOW(), NOW()
    );
END$$

DELIMITER ;
