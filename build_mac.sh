#!/bin/bash
export PATH=$PATH:$(go env GOPATH)/bin
echo "Installing frontend dependencies..."
cd frontend && npm install && cd ..
echo "Building for macOS..."
wails build
open build/bin
