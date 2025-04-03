buf-gen:
	cd ./protobuf && make buf-gen-server

proto-pull:
	git submodule update --remote --force protobuf

go-fmt:
	gofumpt -l -w .

go-lint:
	golangci-lint run ./...