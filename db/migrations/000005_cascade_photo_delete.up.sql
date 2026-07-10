ALTER TABLE carousel_photos
DROP CONSTRAINT IF EXISTS carousel_photos_photo_id_fkey,
ADD CONSTRAINT carousel_photos_photo_id_fkey
    FOREIGN KEY (photo_id)
    REFERENCES photos(id)
    ON DELETE CASCADE;
