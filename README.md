# Sample Queries

# Note - Handling Auth
When you login or signup, go-dropbox will write your JWT to a file in your `$HOME` directory. Instead of copy/pasting the JWT every time you call the API, you can do this:
```
curl -kv "http://localhost:8080/projects/view?project_name=foo" -H 'content-type: application/json' -H "Authorization: $(< ~/go-dropbox-token-jesseb.txt)" | jq -r '.project' | jq
```
Note the use of `-H "Authorization: $(< ~/go-dropbox-token-jesseb.txt)"`, which is saying to read the token from a file. Also note the use of double quotes here.

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
Please note that the `"directories"` key in the payload below is optional - leaving it out will create a basic setup that looks like `{"root": {"files": []}}`.
```
curl -kv "http://localhost:8080/projects/create" -H 'content-type: application/json' -H 'Authorization: ' -d '{"username": "jesseb", "name": "foo", "directories": {"root": {"files": [], "bar": {"files": []}}}}'
```

# Note - Deleting a project also deletes all of the files associated with it.

## Delete a Project
```
curl -kv http://localhost:8080/projects/delete?project_name=foor -H 'content-type: application/json' -H 'Authorization: '
```

## Upload a File
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

## Delete a File
```
curl -kv http://localhost:8080/files/delete -H 'Authorization: ' -d '{"name": "login.go", "path": "/root/foo/baz", "project_name": "foo"}'
```

## Update a File
```
 curl -kv http://localhost:8080/files/update -H 'content-type: multipart/form-data' -H "Authorization: $(< ~/go-dropbox-token-jesseb.txt)" -F 'name=login.go' -F 'project_name=hello' -F 'file=@login.go'
```

## Sharing a File
- YOU own the file and want to share it with someone else. Note that once the link is in the wild,
anyone can access it, so share wisely!
- Also, note that the hash is a checksum of the file's current content. If the file content changes, the
hash is invalid and the **owner will have to reshare with the user.**
```
curl -kv http://localhost:8080/files/sharing -H 'content-type: application/json' -d '{"name": "main.go", "project_name": "foo", "project_id": 6}' -H 'Authorization: '
```
