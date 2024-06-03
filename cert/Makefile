GO111MODULE = on
export GO111MODULE

#go build google.golang.org/protobuf/cmd/protoc-gen-go
#PATH="$(CURDIR):$(PATH)" protoc --go_out=. --go_opt=paths=source_relative $<
#rm protoc-gen-go
cert.pb.go: cert.proto .FORCE
	protoc --go_out=. --go_opt=paths=source_relative $<

.FORCE:
