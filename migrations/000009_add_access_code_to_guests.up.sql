ALTER TABLE guests ADD COLUMN access_code CHAR(6);
UPDATE guests SET access_code = lpad(floor(random() * 1000000)::int::text, 6, '0');
ALTER TABLE guests ALTER COLUMN access_code SET NOT NULL;
