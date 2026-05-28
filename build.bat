@echo off
setlocal
cd /d "%~dp0"
call npm.cmd install
if errorlevel 1 exit /b %errorlevel%
call npm.cmd run build:win
if errorlevel 1 exit /b %errorlevel%
echo Artifacts: release\
endlocal
