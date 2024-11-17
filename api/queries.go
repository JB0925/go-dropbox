package api

// These queries are used to perform CRUD operations on the various database tables
// used by go-dropbox. To see the table definitions, please see `setup-test-db.sh`,
// located at the root of the project.

const checkUserExistsQuery = `SELECT * FROM users WHERE username = $1`
const createNewUserQuery = `INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`
const getPasswordQuery = `SELECT id, password FROM users WHERE username = $1`
const createProjectQuery = `INSERT INTO projects (project_name, directories, user_id, created_at, mtime) VALUES ($1, $2, $3, $4, $5)`
const getUserQuery = `SELECT id FROM users WHERE username = $1`
const checkProjectExistsQuery = `SELECT id FROM projects WHERE project_name = $1 and user_id = $2`
const uploadFileQuery = `INSERT INTO files (name, path, data, user_id, project_id, created_at, mtime) VALUES ($1, $2, $3, $4, $5, $6, $7)`
const updateFileQuery = `UPDATE files SET data = $1, mtime = $2 WHERE name = $3 AND project_id = $4 AND user_id = $5`
const getProjectIdQuery = `SELECT id, directories FROM projects WHERE project_name = $1 AND user_id = $2`
const checkFileExistsQuery = `SELECT id FROM files WHERE project_id = $1 AND user_id = $2 AND name = $3`

//lint:ignore U1000 Ignore unused SQL query during lint check
const getFilesQuery = `SELECT id FROM files WHERE project_id = $1 AND user_id = $2`
const getProjectDirectoriesQuery = `SELECT directories FROM projects WHERE id = $1`
const getFileQuery = `SELECT files.data, files.user_id FROM files JOIN projects ON files.project_id = projects.id WHERE files.name = $1 AND projects.project_name = $2`
const updateProjectDirectoryQuery = `UPDATE projects SET directories = $1, mtime = $3 WHERE id = $2`
const viewProjectQuery = `SELECT directories FROM projects WHERE project_name = $1 and user_id = $2`
const deleteFileQuery = `DELETE FROM files USING projects WHERE files.project_id = projects.id AND files.name = $1 AND projects.project_name = $2 AND files.user_id = $3 AND files.project_id = $4 AND files.path = $5`
const deleteProjectQuery = `DELETE FROM projects WHERE project_name = $1 AND user_id = $2`
const createSharedFileQuery = `INSERT INTO shared (hash, user_id, project_id, file_name, project_name) VALUES ($1, $2, $3, $4, $5)`
const getSharedFileDetailsQuery = `SELECT file_name, project_name FROM shared WHERE hash = $1`
