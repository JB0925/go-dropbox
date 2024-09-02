package api

const checkUserExistsQuery = `SELECT * FROM users WHERE username = $1`
const createNewUserQuery = `INSERT INTO users (username, password) VALUES ($1, $2)`
const getPasswordQuery = `SELECT password FROM users WHERE username = $1`
const createProjectQuery = `INSERT INTO projects (project_name, directories, user_id, created_at, mtime) VALUES ($1, $2, $3, $4, $5)`
const getUserQuery = `SELECT id FROM users WHERE username = $1`
const checkProjectExistsQuery = `SELECT id FROM projects WHERE project_name = $1`
const uploadFileQuery = `INSERT INTO files (name, path, project_name, data, user_id, project_id, created_at, mtime) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
const getProjectIdQuery = `SELECT id, directories FROM projects WHERE project_name = $1`
const checkFileExistsQuery = `SELECT id FROM files WHERE project_id = $1 AND user_id = $2 AND name = $3`
const getFilesQuery = `SELECT id FROM files WHERE project_id = $1 AND user_id = $2`
const getProjectDirectoriesQuery = `SELECT directories FROM projects WHERE id = $1`
const getFileQuery = `SELECT data, user_id FROM files WHERE name = $1 AND project_name = $2`
const updateProjectDirectoryQuery = `UPDATE projects SET directories = $1, mtime = $3 WHERE id = $2`
const viewProjectQuery = `SELECT directories FROM projects WHERE project_name = $1 and user_id = $2`
const deleteFileQuery = `DELETE FROM files WHERE name = $1 AND project_name = $2 AND user_id = $3 AND project_id = $4 AND path = $5`
