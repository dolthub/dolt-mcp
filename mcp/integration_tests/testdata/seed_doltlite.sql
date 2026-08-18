CREATE TABLE people (
    id VARCHAR(36) PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL
);
INSERT INTO people VALUES (lower(hex(randomblob(16))), 'tim', 'sehn');
INSERT INTO people VALUES (lower(hex(randomblob(16))), 'brian', 'hendricks');
INSERT INTO people VALUES (lower(hex(randomblob(16))), 'aaron', 'son');
