swag fmt ./...
swag init --parseDependency -d "../../internal/apps/server" -g "server.go" -o "../../internal/apps/server/docs/"
go build .