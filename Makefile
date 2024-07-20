VERSION=$$(cat VERSION)
GOAMD64:=v3

all: remotebackup

remotebackup:
	cd cmd/remotebackup; GOAMD64=$(GOAMD64) go build -buildmode=pie

vulns:
	govulncheck -show verbose ./...

quality:
	go vet ./...
	golangci-lint run --enable godox --enable gomnd --enable gosec --enable errorlint --enable gofmt \
        --enable unconvert --enable ginkgolinter ./...

clean:
	rm -f cmd/remotebackup/remotebackup
