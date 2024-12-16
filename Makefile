.PHONY: all clean
all: clean
	@go mod tidy
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o bin/TapTapSpider.exe
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o bin/TapTapSpider
	@upx -9 bin/*

clean:
	@rm -rf bin/*