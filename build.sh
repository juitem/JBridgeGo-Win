#!/usr/bin/env bash
set -euo pipefail
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

# electron-builder가 자식 프로세스 에러를 "Exit handler never called!" 같은
# 모호한 메시지로 가리는 경우가 많다. JBRIDGE_DEBUG=1 이거나 빌드가 실패하면
# DEBUG=electron-builder 로 재실행해 실제 원인을 노출한다.
on_fail() {
  trap - ERR
  echo "" >&2
  echo "==> npm run $TARGET 실패. electron-builder 디버그 로그로 재실행한다." >&2
  echo "==> (자식 프로세스의 실제 에러 — fpm/deb, AppImage, OOM 등 — 가 아래에 보인다)" >&2
  echo "" >&2
  DEBUG=electron-builder npm run "$TARGET"
}

echo "==> Detected $(uname -s) → npm run $TARGET"
npm install

if [ "${JBRIDGE_DEBUG:-0}" = "1" ]; then
  DEBUG=electron-builder npm run "$TARGET"
else
  trap on_fail ERR
  npm run "$TARGET"
  trap - ERR
fi

echo "==> Artifacts: release/"
