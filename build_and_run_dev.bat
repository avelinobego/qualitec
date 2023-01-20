taskkill /im server-dev.exe /f
call embed_assets.bat
del .build\server-dev.exe
go build -o .build/server-dev.exe -tags dev
IF %ERRORLEVEL% EQU 0 (
  echo ok
) ELSE ( 
  pause
)
cd .build
server-dev.exe

