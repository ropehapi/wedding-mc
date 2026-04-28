ALTER TABLE guests ADD CONSTRAINT guests_wedding_id_access_code_unique UNIQUE (wedding_id, access_code);
