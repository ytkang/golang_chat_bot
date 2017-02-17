How to run in linux system

1. change "gopath" file in this folder.

2. enter command "source gopath"

3. install requirment packages using "go get" command

4. set const host values in chat.go file
	<code>const (
		listenAddr = "10.0.100.31:3000" // server address
		mongoHost = "127.0.0.1:27017" // mongodb host
	)</code>

5. go run chat.go