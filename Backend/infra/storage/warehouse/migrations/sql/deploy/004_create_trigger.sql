CREATE OR REPLACE FUNCTION update_shelf_used_capacity()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE shelf
    SET used_capacity = (
        SELECT COALESCE(SUM(allocated_capacity), 0)
        FROM shelf_product
        WHERE shelf_id = COALESCE(NEW.shelf_id, OLD.shelf_id)
    )
    WHERE id = COALESCE(NEW.shelf_id, OLD.shelf_id);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_shelf_capacity
AFTER INSERT OR UPDATE OR DELETE ON shelf_product
FOR EACH ROW EXECUTE FUNCTION update_shelf_used_capacity();