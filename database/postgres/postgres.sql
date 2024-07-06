CREATE TYPE Contact_linkPrecedence 
AS 
ENUM('primary', 'secondary');

CREATE TABLE IF NOT EXISTS Contact (
    id SERIAL PRIMARY KEY,
    phoneNumber VARCHAR(12),
    email VARCHAR(80),
    linkedId INTEGER REFERENCES Contact (id) ON DELETE CASCADE,
    linkPrecedence Contact_linkPrecedence NOT NULL DEFAULT 'primary',
    createdAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deletedAt TIMESTAMP
);