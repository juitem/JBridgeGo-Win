# JBridge Desktop

Electron 기반 데스크탑 셸 — JBridge5 등 웹 서비스를 macOS / Windows / Linux에서 동일한 동작으로 임베드합니다.

## Requirements
- Node.js 20+
- npm

## Development
```bash
npm install
npm run dev
```

## Build
```bash
npm run build:mac     # .dmg
npm run build:win     # .exe (NSIS)
npm run build:linux   # .AppImage, .deb
```
결과물은 `release/` 디렉토리에 생성됩니다.

## Settings
설정 파일 위치(OS별):
- macOS: `~/Library/Application Support/jbridge-desktop/settings.json`
- Windows: `%APPDATA%/jbridge-desktop/settings.json`
- Linux: `~/.config/jbridge-desktop/settings.json`

## History
이전 Wails(Go) 구현은 `legacy` 브랜치에 보존되어 있습니다.
