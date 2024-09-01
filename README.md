# Sample Queries

## Sign Up - returns a JWT
```
curl -kv http://localhost:8080/signup -H 'content-type: application/json' -d '{"username": "jesseb", "password": "foobar"}'
```

## Login - returns a JWT
```
curl -kv http://localhost:8080/login -H 'content-type: application/json' -d '{"username": "jesseb", "password": "foobar"}'
```

## Create a Project
```
curl -kv "http://localhost:8080/projects/create" -H 'content-type: application/json' -H 'Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjUyNDYyMjIsImlzcyI6Implc3NlYiJ9.-bRqHr0JGEQUqAEtEmoAsxLqOckD828iHVyV82Db6CY' -d '{"username": "jesseb", "name": "foo", "directories": {"root": {"files": [], "bar": {"files": []}}}}'
```

## Create a File
- must belong to a project
- content-type is multipart/form-data
- can create a new filepath by passing in what you want the filepath to be
```
curl -kv http://localhost:8080/files/upload -H 'content-type: multipart/form-data' -H 'Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjUxNDA5NjEsImlzcyI6Implc3NlYiJ9.yaXz7O6llUhFKEZCxS0Mr2ps8lUHSaQC0eQq8I59EPE' -F 'name=YESLogo2.png' -F 'project_name=foo' -F 'path=/root/bar' -F 'file=@/Users/jessebrink/curbside/src/YESLogo2.png'
```

## Download a File
```
curl -kv "http://localhost:8080/files/download?name=YESLogo2.png&project_name=foo" -H 'Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjUxNDA5NjEsImlzcyI6Implc3NlYiJ9.yaXz7O6llUhFKEZCxS0Mr2ps8lUHSaQC0eQq8I59EPE' -o YESLogo2.png
```
