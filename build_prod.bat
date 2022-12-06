call embed_assets.bat
del .build\telemetria
set GOOS=linux
set GOARCH=amd64
go build -ldflags "-s -w" -o .build\telemetria -tags dev
pause