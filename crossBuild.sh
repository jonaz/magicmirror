go get -d -v
for GOOS in linux; do
	for GOARCH in arm amd64; do
		GOARCH=$GOARCH GOOS=$GOOS go build -v -o magicMirror-$GOOS-$GOARCH
	done
done
