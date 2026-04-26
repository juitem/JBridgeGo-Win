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
