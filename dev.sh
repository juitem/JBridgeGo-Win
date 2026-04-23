#!/bin/bash
# Wails 실행 경로 추가
export PATH=$PATH:$(go env GOPATH)/bin

echo "🚀 JBridgeGo-Desktop을 개발 모드로 실행합니다..."
wails dev
