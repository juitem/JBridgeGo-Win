@echo off
echo 🔨 JBridgeGo-Win 빌드를 시작합니다...
cd frontend && npm install && cd ..
wails build -platform windows/amd64
pause
