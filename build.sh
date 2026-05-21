#!/usr/bin/env bash
set -e
cd "$(dirname "$0")"

case "$(uname -s)" in
  Darwin*)  TARGET="build:mac"   ;;
  Linux*)   TARGET="build:linux" ;;
  MINGW*|MSYS*|CYGWIN*) TARGET="build:win" ;;
  *)
    echo "Unsupported OS: $(uname -s)" >&2
    exit 1
    ;;
esac

echo "==> Detected $(uname -s) → npm run $TARGET"
npm install
npm run "$TARGET"
echo "==> Artifacts: release/"
