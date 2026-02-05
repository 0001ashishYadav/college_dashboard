-- CREATE TABLE admin(
--     id SERIAL PRIMARY KEY,
--     name TEXT NOT NULL,
--     role TEXT NOT NULL,
--     email TEXT UNIQUE NOT NULL,
--     password TEXT NOT NULL
-- );

CREATE TABLE institutes (
    id SERIAL PRIMARY KEY,
    name TEXT  NOT NULL,
    code TEXT UNIQUE NOT NULL,
    email TEXT ,
    phone TEXT, 
    address TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (now())
);



CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    institute_id INT NOT NULL REFERENCES institutes (id),
    name TEXT  NOT NULL,
    email TEXT  UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT DEFAULT 'admin',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now())
);


CREATE TABLE notices (
    id SERIAL PRIMARY KEY,
    institute_id INT NOT NULL REFERENCES institutes (id),
    title TEXT  NOT NULL,
    description TEXT,
    is_published BOOLEAN DEFAULT true,
    publish_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now())
);


CREATE TABLE carousels (
    id SERIAL PRIMARY KEY,
    institute_id INT NOT NULL REFERENCES institutes (id),
    title TEXT ,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now())
);


CREATE TABLE photos (
    id SERIAL PRIMARY KEY,
    image_url TEXT NOT NULL,
    alt_text TEXT ,
    uploaded_by INT NOT NULL REFERENCES users (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now())
);


CREATE TABLE carousel_photos (
    id SERIAL PRIMARY KEY,
    carousel_id INT NOT NULL REFERENCES carousels(id),
    photo_id INT NOT NULL REFERENCES photos(id),
    display_text TEXT ,   -- text shown on carousel image
    display_order INT DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (now())
);


