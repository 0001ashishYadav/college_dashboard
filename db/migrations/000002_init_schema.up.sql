ALTER TABLE photos
ADD COLUMN institute_id INT NOT NULL REFERENCES institutes(id);
