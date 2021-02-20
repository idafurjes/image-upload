# Upload image API

## Description
There are two endpoints for this API. 

First one is the to upload an image from a current directory into the *image* directory. The allowed extensions are *png* and *jpg/jpeg*.
It is a POST endpoint http://localhost:8080/image
Response of the endpoint is a json with *ID* of the image.

Second endpoint GET http://localhost:8080/image/{id} is an endpoint that return the image when file ID is given.
Response is the image itself.

## Usage example
```
curl http://localhost:8080/image -F "fileupload=@penguin.png" -vvv

curl http://localhost:8080/image/{id} --output test.png
```

## Test
run `go test`