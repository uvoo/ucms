go fmt main.go
find ./internal -type f -name "*.go" -exec gofmt -w {} +
