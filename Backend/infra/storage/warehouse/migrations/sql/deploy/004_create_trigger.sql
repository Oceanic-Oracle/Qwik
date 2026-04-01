CREATE OR REPLACE FUNCTION update_shelf_used_capacity()
RETURNS TRIGGER AS $$ BEGIN
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

DROP TRIGGER IF EXISTS trg_update_shelf_capacity ON shelf_product;
CREATE TRIGGER trg_update_shelf_capacity
AFTER INSERT OR UPDATE OR DELETE ON shelf_product
FOR EACH ROW EXECUTE FUNCTION update_shelf_used_capacity();


CREATE OR REPLACE FUNCTION calculate_allocated_capacity()
RETURNS TRIGGER AS $$ BEGIN
    NEW.allocated_capacity := NEW.quantity * (
        SELECT COALESCE(volume, 0) FROM product WHERE id = NEW.product_id
    );
    RETURN NEW;
END;
 $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_calculate_capacity ON shelf_product;
CREATE TRIGGER trg_calculate_capacity
BEFORE INSERT OR UPDATE OF quantity, product_id ON shelf_product
FOR EACH ROW
WHEN (NEW.product_id IS NOT NULL)
EXECUTE FUNCTION calculate_allocated_capacity();