package api

const checkUserExistsQuery = `SELECT * FROM users WHERE username = $1`
const createNewUserQuery = `INSERT INTO users (username, password) VALUES ($1, $2)`
const getPasswordQuery = `SELECT password FROM users WHERE username = $1`
const createProjectQuery = `INSERT INTO projects (project_name, directories, user_id) VALUES ($1, $2, $3)`
const getUserQuery = `SELECT id FROM users WHERE username = $1`
const checkProjectExistsQuery = `SELECT id FROM projects WHERE project_name = $1`
const uploadFileQuery = `INSERT INTO files (name, path, project_name, data, user_id, project_id) VALUES ($1, $2, $3, $4, $5, $6)`
const getProjectIdQuery = `SELECT id, directories FROM projects WHERE project_name = $1`
const checkFileExistsQuery = `SELECT id FROM files WHERE project_id = $1 AND user_id = $2 AND name = $3`
const getFilesQuery = `SELECT id FROM files WHERE project_id = $1 AND user_id = $2`
const getProjectDirectoriesQuery = `SELECT directories FROM projects WHERE id = $1`
const getFileQuery = `SELECT data, user_id FROM files WHERE name = $1 AND project_name = $2`
