package examples

const (
	createUsersTable = `
CREATE TABLE users (
  user_id  INTEGER UNIQUE,
  username VARCHAR(40),
  email    VARCHAR(40)
)
`
	seedUsersTable = `
INSERT INTO users (user_id, username, email) VALUES
  (0, 'root', ''),
  (1, 'dhermes', 'dhermes@mail.invalid')
`
	addUsersCityColumn = `
ALTER TABLE users
  ADD COLUMN city VARCHAR(100)
`
	renameRoot = `
UPDATE users
  SET username = 'admin'
  WHERE username = 'root'
`
	addUsersEmailIndexConcurrently = `
CREATE UNIQUE INDEX CONCURRENTLY uq_users_email ON users (email)
`
	createBooksTable = `
CREATE TABLE books (
  user_id INTEGER,
  title   VARCHAR(40),
  author  VARCHAR(40)
)
`
	createMoviesTable = `
CREATE TABLE movies (
  user_id  INTEGER,
  title    VARCHAR(40),
  director VARCHAR(40)
)
`
)
