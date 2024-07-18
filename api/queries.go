package api

const checkUserExistsQuery = `SELECT * FROM users WHERE username = $1`
const createNewUserQuery = `INSERT INTO users (username, password) VALUES ($1, $2)`