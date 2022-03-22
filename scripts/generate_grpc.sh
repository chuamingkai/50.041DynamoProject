# Might need to run sudo chmod 777 or smth on this file first
export PATH="$PATH:$(go env GOPATH)/bin"
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/internalcomm/internalcomm.proto