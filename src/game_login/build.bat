SET BIN_NAME=login.exe
SET BIN_DIR=..\..\deploy\bin

@REM windows
SET CGO_ENABLED=0 GOOS=windows GOARCH=amd64

go build -ldflags="-w -s"

move /Y %BIN_NAME% %BIN_DIR%