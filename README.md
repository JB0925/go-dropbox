# Sample Queries

## Sign Up - returns a JWT
```
curl -kv http://localhost:8080/signup -H 'content-type: application/json' -d '{"username": "jesseb", "password": "foobar"}'
```

## Login - returns a JWT
```
curl -kv http://localhost:8080/login -H 'content-type: application/json' -d '{"username": "jesseb", "password": "foobar"}'
```


# Note - each of the below requires a JWT acquired from signup or login. It is placed as the value in the "Authorization" header. Simply copy and paste yours into this example. Removing it here makes it so that you do not have to erase it if you copy the example.

## Create a Project
```
curl -kv "http://localhost:8080/projects/create" -H 'content-type: application/json' -H 'Authorization: ' -d '{"username": "jesseb", "name": "foo", "directories": {"root": {"files": [], "bar": {"files": []}}}}'
```

## Create a File
- must belong to a project
- content-type is multipart/form-data
- can create a new filepath by passing in what you want the filepath to be
```
curl -kv http://localhost:8080/files/upload -H 'content-type: multipart/form-data' -H 'Authorization: ' -F 'name=YESLogo2.png' -F 'project_name=foo' -F 'path=/root/bar' -F 'file=@/Users/jessebrink/curbside/src/YESLogo2.png'
```

## Download a File
```
curl -kv "http://localhost:8080/files/download?name=YESLogo2.png&project_name=foo" -H 'Authorization: ' -o YESLogo2.png
```

## View a Project
```
curl -kv "http://localhost:8080/projects/view?project_name=foo" -H 'content-type: application/json' -H 'Authorization: ' | jq -r '.project' | jq
```
