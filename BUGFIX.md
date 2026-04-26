# Bug Fix: macOS 메뉴 렌더링 실패 / null 배열 직렬화

**발견일**: 2026-04-26  
**증상**: 맥에서 앱 실행 시 화면이 안 보이고, 메뉴 버튼을 누르면 메뉴가 즉시 사라지고 아무것도 표시되지 않음

---

## 근본 원인

Go의 URL 리스트 조작 함수들이 `var res []string` 패턴으로 결과 슬라이스를 초기화해, 결과가 비어있을 때 `nil`을 반환한다. Go에서 `nil` 슬라이스는 JSON으로 직렬화하면 `null`이 된다.

이 `null`이 Wails를 통해 JavaScript로 전달되면, Vue computed 프로퍼티(`recentOnly`, `currentIndex`)에서 `null.includes()`, `null.filter()`, `null.indexOf()` 호출 시 **TypeError**가 발생한다. Vue 3는 이 오류를 내부적으로 처리하지만, 메뉴 컴포넌트 렌더링이 실패해 메뉴가 렌더링되지 않거나 즉시 사라지는 것처럼 보인다.

### 실제 저장된 settings.json 상태 (버그 재현)

```json
{
  "pinnedUrls": null,
  "preloadUrls": null,
  "manualTrustedHosts": null
}
```

### 문제가 된 코드 패턴 (app.go)

```go
// 잘못된 패턴 — 결과가 비면 nil 반환 → JSON null
filter := func(list []string) []string {
    var res []string  // nil로 시작
    for _, u := range list { if u != targetUrl { res = append(res, u) } }
    return res  // 빈 결과면 nil
}
```

### 문제가 된 Vue 코드 (App.vue)

```js
// null.includes() → TypeError
const recentOnly = computed(() => state.recentUrls.filter(u => !state.pinnedUrls.includes(u)))
// null.indexOf() → TypeError
const currentIndex = computed(() => state.rotationUrls.indexOf(state.serverUrl))
```

---

## 수정 내용

### 1. `internal/state/storage.go` — Load 시 nil 슬라이스 보정 추가

```go
if state.PinnedUrls == nil       { state.PinnedUrls = []string{} }
if state.RecentUrls == nil       { state.RecentUrls = []string{} }
if state.RotationUrls == nil     { state.RotationUrls = []string{} }
if state.PreloadUrls == nil      { state.PreloadUrls = []string{} }
if state.ManualTrustedHosts == nil { state.ManualTrustedHosts = []string{} }
```

### 2. `app.go` — 5개 함수의 결과 슬라이스 초기화 방식 변경

대상 함수: `DeleteUrl`, `TogglePin`, `ToggleRotation`, `TogglePreload`, `RemoveTrustedHost`

```go
// 수정 — 결과가 비어도 [] 반환
res := []string{}  // nil 대신 빈 슬라이스로 초기화
```

### 3. `frontend/src/App.vue` — null 안전 처리 2곳

```js
// updateState: null 배열 보정
const updateState = (s) => {
  if (!s) return
  if (!s.pinnedUrls) s.pinnedUrls = []
  if (!s.recentUrls) s.recentUrls = []
  // ... 등
  Object.assign(state, s)
}

// computed: || [] 방어 처리
const currentIndex = computed(() => (state.rotationUrls || []).indexOf(state.serverUrl))
const recentOnly = computed(() => (state.recentUrls || []).filter(u => !(state.pinnedUrls || []).includes(u)))
```

### 4. `~/.jbridgego-win/settings.json` — 기존 저장 파일 직접 수정

`null` → `[]` 로 직접 수정 (앱 재시작 시 자동 복구도 됨)

---

## 재발 방지

Go에서 슬라이스를 조작하는 함수는 항상 `res := []string{}` 로 초기화할 것.  
`var res []string` 은 nil을 반환할 수 있으므로 API 응답 경계에서는 사용하지 않는다.

---

# Bug Fix: 페이지 전환 시 흰색 화면 번쩍임

**발견일**: 2026-04-26  
**증상**: URL 전환 시 전체 화면이 흰색으로 순간 번쩍이고 페이지가 새로 뜨는 현상

---

## 근본 원인

`<iframe :key="state.serverUrl">` 로 인해 `serverUrl`이 바뀔 때마다 Vue가 iframe DOM 요소를 **파괴하고 재생성**한다. 이 순간 배경(흰색)이 노출되면서 흰색 번쩍임이 발생한다.

추가로:
- `updateState(newState)` 호출 시 `:src` 바인딩이 src를 갱신
- 이후 `iframeRef.value.src = url` 로 수동으로 또 한번 src 설정
- 이중 네비게이션 발생

---

## 수정 내용

### 1. `frontend/src/App.vue` — `:key` 제거

```html
<!-- 수정 전 -->
<iframe :key="state.serverUrl" :src="state.serverUrl" ...>

<!-- 수정 후 — 같은 iframe DOM을 유지하며 src만 변경 -->
<iframe :src="state.serverUrl" ...>
```

### 2. `handleSwitch` — 이중 네비게이션 방지

```js
// 수정 전: updateState 후 iframeRef.src 를 또 설정 (이중 네비게이션)
updateState(newState)
iframeRef.value.src = url

// 수정 후: 먼저 src 직접 설정 → updateState(:src 바인딩이 같은 url이므로 무시됨)
if (iframeRef.value) iframeRef.value.src = url
updateState(newState)
```

### 3. 배경색 변경 — 로딩 중 빈 공간을 다크 테마로

```css
/* 수정 전 */
.webview-wrapper { background: #ffffff; }
.main-webview    { background: white; }

/* 수정 후 */
.webview-wrapper { background: #1e1e2e; }
.main-webview    { background: #1e1e2e; }
```

---

## 재발 방지

Vue에서 iframe의 URL을 바꿀 때 `:key` 를 사용하면 DOM이 재생성되어 번쩍임이 발생한다.  
iframe은 `:key` 없이 `:src` 바인딩 또는 ref를 통해 src를 직접 변경하는 방식을 사용할 것.

---

# Bug Fix: target="_blank" 링크 클릭 시 아무것도 안 열림

**발견일**: 2026-04-26  
**증상**: iframe 안에서 링크를 클릭해도 아무 반응 없음 (Android 버전과 다른 동작)

---

## 근본 원인

Android JBridgeGo는 WebView를 직접 사용해 모든 네비게이션을 앱 레벨에서 제어한다. 데스크탑 버전은 `<iframe>` 구조로, `target="_blank"` 링크가 새 창을 요청할 때 Wails WKWebView가 멀티 윈도우를 지원하지 않아 요청을 **무시**한다. 또한 iframe에 `allow="popups"` 권한이 없어 팝업 요청이 부모 프레임으로 전달조차 안 됐다.

---

## 수정 내용

### 1. `app.go` — OpenInBrowser 메서드 추가

```go
func (a *App) OpenInBrowser(targetUrl string) {
    runtime.BrowserOpenURL(a.ctx, targetUrl)
}
```

### 2. `frontend/src/App.vue` — window.open 오버라이드

```js
// onMounted에서 window.open을 intercept → OS 기본 브라우저로 라우팅
window.open = (url) => {
    if (url && url !== 'about:blank' && !String(url).startsWith('javascript:')) {
        api.OpenInBrowser(String(url))
    }
    return null
}
```

### 3. iframe — allow="popups" 추가

```html
<iframe allow="fullscreen; clipboard-read; clipboard-write; popups">
```

`popups` 권한 없이는 iframe 내부의 팝업 요청이 부모 프레임의 `window.open`까지 전달되지 않는다.

---

## 재발 방지

Wails는 멀티 윈도우 미지원. `target="_blank"` 처리는 항상 `runtime.BrowserOpenURL`로 OS 브라우저에 위임한다.

---

# Bug Fix: 메뉴 삭제 버튼 동작 안 함 / 순환 목록 순서 편집 불가

**발견일**: 2026-04-27  
**증상**: ✕ 버튼 클릭 시 아무 반응 없음. 순환 목록 순서 변경 불가. 메뉴에서 보이는 순서와 ⟨/⟩ 순환 순서가 불일치

---

## 근본 원인

### 1. window.confirm() Wails에서 동작 안 함
`handleDelete`에서 `window.confirm()`을 사용했는데, Wails WKWebView는 `alert()`, `confirm()`, `prompt()` 등 네이티브 JS 다이얼로그를 표시하지 않고 `false`를 반환한다. 결과적으로 `if(confirm(...))` 조건이 항상 false → 삭제 코드가 실행되지 않음.

### 2. 순환 목록 순서 편집 UI 없음
순서편집(▲▼) 기능이 `pinnedUrls`(즐겨찾기) 섹션에만 있었다. 사용자는 ★ 핀을 사용하지 않고 rotation만 사용 중이어서, 메뉴에서 순서 편집이 불가능했다.

### 3. 메뉴 섹션과 순환 순서 불일치
메뉴 "최근" 섹션은 `recentUrls` 순서(최신 방문 순)로 표시하지만, ⟨/⟩ 버튼은 `rotationUrls` 순서로 동작해 사용자가 두 순서를 혼동.

---

## 수정 내용

### 1. `handleDelete` — confirm() 제거
```js
// 수정 전
async function handleDelete(url) { if(confirm('삭제하시겠습니까?')) updateState(await api.DeleteUrl(url)) }

// 수정 후
async function handleDelete(url) { updateState(await api.DeleteUrl(url)) }
```

### 2. 메뉴 섹션 재구성 (App.vue)

- **즐겨찾기**: pinnedUrls (기존 순서편집 유지)
- **순환 목록** ← 새 섹션: rotationUrls 중 pinned 제외, rotation 순서 유지, 순서편집(▲▼) 지원
- **최근**: pinned도 rotation도 아닌 recentUrls만 표시

```js
// 순환 목록 (pinned 제외, rotationUrls 순서 유지)
const rotationOnly = computed(() =>
    (state.rotationUrls || []).filter(u => !(state.pinnedUrls || []).includes(u))
)
// 최근 (rotation/pinned 모두 제외)
const recentOnly = computed(() =>
    (state.recentUrls || []).filter(u =>
        !(state.pinnedUrls || []).includes(u) && !(state.rotationUrls || []).includes(u)
    )
)
```

### 3. pinned 순서 변경 시 rotation 동기화 (Android parity, app.go)

```go
// pinned 항목 이동 시 rotationUrls 안의 pinned 항목 상대 순서를 pinnedUrls에 맞게 동기화
// (Android c8dc9ec 커밋과 동일 로직)
```

---

## 재발 방지

Wails WebView에서 `window.alert/confirm/prompt`는 동작하지 않는다. 사용자 확인이 필요한 경우 Vue 인라인 다이얼로그를 사용할 것.
