# http_test

lightweight http request testing

## go version
- go1.22.5 or above


## run on local

run on local

```bash
go run main.go start
```

## run with docker

```
# pull from docker hub
docker pull cclam0827/httptest:latest
# run on local
docker run --name httptest -p 80:4000 --rm cclam0827/httptest:latest
```

## docker build

```
# change the push path if needed
docker buildx build --platform linux/amd64,linux/arm64 --push -t cclam0827/httptest:latest .
```


## Generate Swagger doc

ref[https://github.com/swaggo/swag]
ref[https://github.com/swaggo/gin-swagger]

```
# Download Swag
go install github.com/swaggo/swag/cmd/swag@latest

# Download gin-swagger
go get -u github.com/swaggo/gin-swagger
go get -u github.com/swaggo/files
```

```
import "github.com/swaggo/gin-swagger" // gin-swagger middleware
import "github.com/swaggo/files" // swagger embed files
```

```
swag init

or

$GOPATH/bin/swag init
```

```
# doc
http://localhost:4000/swagger/doc.json

# UI
http://localhost:4000/swagger/index.html
```
