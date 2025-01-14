@echo off
@REM author r3inbowari
@REM date 2021/10/1
@REM 编译前请确保已安装 git

set packageName=cmd
set appName=bilicoin
set buildVersion=v1.1.2
set major=1
set minor=1
set patch=2
set Mode=REL

for /f "delims=" %%i in ('go version') do (set goVersion=%%i)
for /f "delims=" %%i in ('git show -s --format^=%%H') do (set gitHash=%%i)
for /f "delims=" %%i in ('git show -s --format^=%%cd') do (set buildTime=%%i)

echo ===================================================
echo                  Go build running
echo ===================================================
echo %goVersion%
echo build hash %gitHash%
echo build time %buildTime%
echo build tag %buildVersion%
echo ===================================================

set GoBuildEntry=go build -ldflags "-X 'main.Major=%major%' -X 'main.Minor=%minor%'-X 'main.Patch=%patch%' -X 'main.ReleaseVersion=%buildVersion%' -X 'main.Mode=%Mode%' -X 'main.goVersion=%goVersion%' -X 'main.GitHash=%gitHash%' -X 'main.buildTime=%buildTime%'"

if not exist build_upx (
    md build_upx
)

cd %packageName%

set GOOS=windows
set GOARCH=amd64
call:compose
%GoBuild%.exe
copy ..\build\%appName%_%GOOS%_%GOARCH%_%buildVersion%.exe ..\build_upx\
set upxArgs=%upxArgs% %appName%_%GOOS%_%GOARCH%_%buildVersion%.exe


pause
exit

@rem return a global value as generic builder named "GoBuild"
:compose
echo compiling %appName%_%GOOS%_%GOARCH%_%buildVersion%
@rem set output
set Output=-o ../build/%appName%_%GOOS%_%GOARCH%_%buildVersion%
@rem generate builder
set GoBuild=%GoBuildEntry% %Output%
@REM echo %Go%
goto:eof