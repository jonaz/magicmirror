#docker run --rm -it -v "$(pwd)":/usr/src/myapp -w /usr/src/myapp golang:1.3.1-cross bash -c ./crossBuild.sh

docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp -e GOOS=linux -e GOARCH=arm golang:1.6 /bin/bash -c "go get -d -v ; go build -v -o magicMirror-arm"

