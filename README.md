# Secure Container Tool

- A cli tool to build encrypted containers;
- You can add files to the container, and remove files from the container;
- Built with Golang.

Usage examples:
- ./bin/app 
- ./bin/app -h
- ./bin/app -create -name=test
- ./bin/app -addFile -filePath=test.txt -containerPath=test.container
- ./bin/app -addFiles -dirPath= myDir -containerPath=test.container
- ./bin/app -mount -containerPath=test.container
- ./bin/app -unmount -containerPath=test.container
- ./bin/app -rest -containerPath=test.container


To build:
- In your terminal: . changePath
- go get github.com/howeyc/gopass
- go install app

To generate an executable file:
- In your terminal: .changePath
- go build -o \<app name\> app