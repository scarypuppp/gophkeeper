ALTER TABLE texts ALTER COLUMN id DROP IDENTITY;
CREATE SEQUENCE texts_id_seq OWNED BY texts.id;
ALTER TABLE texts ALTER COLUMN id SET DEFAULT nextval('texts_id_seq');

ALTER TABLE cards ALTER COLUMN id DROP IDENTITY;
CREATE SEQUENCE cards_id_seq OWNED BY cards.id;
ALTER TABLE cards ALTER COLUMN id SET DEFAULT nextval('cards_id_seq');

ALTER TABLE secrets ALTER COLUMN id DROP IDENTITY;
CREATE SEQUENCE secrets_id_seq OWNED BY secrets.id;
ALTER TABLE secrets ALTER COLUMN id SET DEFAULT nextval('secrets_id_seq');

ALTER TABLE files ALTER COLUMN id DROP IDENTITY;
CREATE SEQUENCE files_id_seq OWNED BY files.id;
ALTER TABLE files ALTER COLUMN id SET DEFAULT nextval('files_id_seq');

ALTER TABLE users ALTER COLUMN id DROP IDENTITY;
CREATE SEQUENCE users_id_seq OWNED BY users.id;
ALTER TABLE users ALTER COLUMN id SET DEFAULT nextval('users_id_seq');
