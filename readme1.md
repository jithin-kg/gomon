go build -o bin/gomon ./cmd/gomon
./bin/gomon


To install from github
go install github.com/jithin-kg/gomon/cmd/gomon@latest

Installing from Local Changes (Development Version)
1) Navigate to the cmd/gomon directory within your project
2) Run go build -o $GOPATH/bin/gomon to build the executable and place it in your $GOPATH/bin directory, which should be in your system's PATH.

cd cmd/gomon
go build -o $GOPATH/bin/gomon
or
go build -o $GOPATH/bin/gomon ./cmd/gomon
rm -rf $GOPATH/bin/gomon
air
--
go build -o $GOPATH/bin/air
rm -rf $GOPATH/bin/air
rm $HOME/go/bin/air