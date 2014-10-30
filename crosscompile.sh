
cwd=$(pwd)

if [[ "$1" == "setup" ]]; then
	echo "settings up for first time"
	mkdir -p ~/go/root/go/
	hg clone -u release https://code.google.com/p/go ~/go/root/go/
	cd ~/go/root/go/src
	./all.bash
	go get github.com/laher/goxc 
	goxc -goroot ~/go/root/go/ -bc "arm amd64" -os linux -t
fi

cd $cwd
goxc -goroot ~/go/root/go/ -bc "arm amd64" -os linux
