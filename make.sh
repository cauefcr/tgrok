#!/usr/env sh

git clone https://github.com/clementauger/tor-prebuilt
cd tor-prebuilt
go mod init github.com/clementauger/tor-prebuilt
go mod tidy
go get
make w
make d
make l
go install .
cd ../
go install .