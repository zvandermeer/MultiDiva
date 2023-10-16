env CC="x86_64-w64-mingw32-gcc" CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -o ./bin/MultiDiva-Client.dll -buildmode=c-shared
