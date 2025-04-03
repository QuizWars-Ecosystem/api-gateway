buf-gen:
	cd ./protobuf && make buf-gen-server

proto-pull:
	git submodule update --remote --force protobuf