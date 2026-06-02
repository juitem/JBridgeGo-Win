# JBridge Desktop

Electron 기반 네이티브 셸. JBridge5 등 웹 서비스(localhost:7900 등)를 OS 일관성 있게 임베드.

## Stack
- Electron 42 + electron-vite + TypeScript
- Vue 3 (Composition API) + Pinia (renderer)
- electron-store (JSON 영속화 → `app.getPath('userData')/settings.json`)
- electron-builder (dmg / nsis / AppImage / deb)

## Layout
```
src/
  main/        Electron main process (Node, TS)
    index.ts   BrowserWindow + lifecycle
    ipc.ts     IPC handlers (app.go에서 포팅됨)
    store.ts   electron-store 래퍼
  preload/
    index.ts   contextBridge로 window.api 노출
    index.d.ts renderer용 타입 선언
  renderer/
    index.html
    src/
      App.vue        UI (메뉴, 툴바, iframe 임베드)
      main.ts        Vue 엔트리
      style.css
      components/, assets/
  shared/
    state.ts   AppState 인터페이스 + JBridgeApi 계약 (main/preload/renderer 공유)
```

## Commands
- `npm run dev` — electron-vite 개발 모드 (HMR)
- `npm run build` — 프로덕션 번들 (out/)
- `npm run build:mac|win|linux` — 플랫폼 패키지 (release/)
- `npm run typecheck` — main + renderer 타입체크

## Architecture Notes
- **OS 분기 코드는 0줄이 원칙** — Electron이 Chromium + Node를 OS별 동일 버전으로 번들하므로 비즈니스 로직은 플랫폼 무관해야 함
- **Renderer는 `window.api`만 통해 main과 통신** — direct fs/shell 접근 금지 (`contextIsolation: true`, `nodeIntegration: false`)
- **임베드는 현재 `<iframe>` 사용** — Phase 2에서 X-Frame-Options 차단 사이트나 추가 격리 필요시 `WebContentsView`로 전환 예정
- **상태 전체를 한 번에 직렬화** — 작은 상태(< 수 KB)라 부분 업데이트 IPC 채널은 만들지 않음. 모든 mutator는 새 AppState를 반환

## Migration History
- 2026-05 Wails(Go) → Electron(TS) 전환. 이유: macOS WKWebView IME 버그(한글 입력) + OS별 WebView 차이로 인한 동작 불일치. legacy 브랜치에 Wails 코드 보존.

## DO NOT
- `nodeIntegration: true` 또는 `contextIsolation: false` 켜지 말 것 — preload bridge 우회로 보안 구멍
- main process에서 동기 IO(`readFileSync` 등) 핫패스에 쓰지 말 것 — UI 멈춤
- renderer에서 `require()`/`process` 직접 접근하지 말 것 — `window.api` 경유

## Related
- legacy 브랜치: Wails(Go) 구현, macOS WKWebView 시도 이력
- README.md: 사용자 대상 빌드/설치 안내
