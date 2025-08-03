-- Drop the old unique index on the name column
DROP INDEX categories_name_idx;

-- Create a new composite unique index on name and type
CREATE UNIQUE INDEX categories_name_type_idx ON categories (name, type);
