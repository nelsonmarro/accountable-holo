-- Drop the new composite unique index
DROP INDEX categories_name_type_idx;

-- Recreate the old unique index on the name column
CREATE UNIQUE INDEX categories_name_idx ON categories (name);
