proto-common:
	protoc --go_out=. --go_opt=paths=source_relative        --go-grpc_out=. --go-grpc_opt=paths=source_relative       proto/common/common.proto
service-common:
	go run src/service/common/main.go
gql:
	go run github.com/99designs/gqlgen generate
server-run:
	go run src/server/main.go
run:
	air