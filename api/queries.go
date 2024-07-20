package api

const checkUserExistsQuery = `SELECT * FROM users WHERE username = $1`
const createNewUserQuery = `INSERT INTO users (username, password) VALUES ($1, $2)`
const getPasswordQuery = `SELECT password FROM users WHERE username = $1`
const createProjectQuery = `INSERT INTO projects (project_name, directories, user_id) VALUES ($1, $2, $3)`
const getUserQuery = `SELECT id FROM users WHERE username = $1`
const checkProjectExistsQuery = `SELECT id FROM projects WHERE project_name = $1`