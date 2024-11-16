# Welcome to go-dropbox!
`go-dropbox` is an app designed to allow a user to create projects and upload files associated with those projects. A user can also share files with another person who is not necessarily a user. `go-dropbox` can be used for file storage and file sharing.

`go-dropbox` is a RESTful web API, and the following documentation is written to explain how to interact with it. Please keep in mind that there is no user interface at this time, so the primary way to interact with the API is via cURL or some other HTTP client.

`go-dropbox` allows a user to:
- register an account
- login and get a JWT token for further operations.
- create a project
- view a project
- delete a project ( doing so deletes all associated files )
- upload a file ( doing so will also update the project structure )
- update a file
- delete a file
- download a file
- share a file with anyone
- download a shared file ( available to anyone, even non-account holding users )


# Running the App
Simply run `go run .` to run the app. Note that this requires that you have Go installed on your system.
For more on how to install Go, please see the official [documentation](https://go.dev/doc/install).

# Sample Queries

## Sign Up - returns a JWT
```
curl -kv http://localhost:8080/signup -H 'content-type: application/json' -d '{"username": "jesseb", "password": "foobar"}'
```

## Login - returns a JWT
```
curl -kv http://localhost:8080/login -H 'content-type: application/json' -d '{"username": "jesseb", "password": "foobar"}'
```

## Helpful Tip Re: Auth Tokens in the CLI
When you login to go-dropbox, it will write a JWT token to your home directory. This is the same token that is returned to you via /login. Instead of hunting down this token in your terminal, or copy and pasting from that file each time, you can do something like this:
```
curl -kv "http://localhost:8080/projects/view?project_name=foo" -H 'content-type: application/json' -H "Authorization: $(cat ~/go-dropbox-token-<your-username>.txt)" | jq -r '.project' | jq
```

That will use shell expansion to substitute your actual JWT token into the space where `$(cat ~/go-dropbox-token-username.txt)` resides.


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

## Getting a Shared File
- Someone has shared a file with you and you have the hash of the file. You can use the following api call
to download it:
```
curl -v http://localhost:8080/files/shared -H 'X-GO-DROPBOX-SHARED-HASH: 5ca07cfe530106ffa949c2ac793fcf240a98034d5b42eb3358968f5343745b55' -H 'X-GO-DROPBOX-SHARER: jesseb' -o ./myfile.png
```

- Note the two headers you need to add:
1. `X-GO-DROPBOX-SHARER` - the username of the person who shared it with you.
2. `X-GO-DROPBOX-SHARED-HASH` - the hash of the file content, given to you by the person who shared it.

- Note that this endpoint requires no auth - if the user has the hash and the username of the person who
created it, they will be able to download it.

## Running the Tests
- To run the tests, please run `make test`. If you do not have `make` on your machine, you can run `./run-tests.sh` instead. This script starts a server for testing, runs the tests, and then tears the server down at the end.
