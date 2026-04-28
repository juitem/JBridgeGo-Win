<script setup>
import { reactive, onMounted, onUnmounted, ref, computed } from 'vue'
import * as api from '../wailsjs/go/main/App'

const state = reactive({
  serverUrl: '', gridMode: false, scrollLock: false, zoomLevel: 100, hideStatusBar: true,
  pinnedUrls: [], recentUrls: [], rotationUrls: [], preloadUrls: [],
  urlAliases: {}, keepScreenOn: false, manualTrustedHosts: [], showRotationBtns: true, showMenu: false
})

// UI States
const expanded = ref(true)
const showAddDialog = ref(false)
const addUrlInput = ref('')
const addAliasInput = ref('')
const editingUrl = ref(null)
const reorderMode = ref(false)
const trustInput = ref('')

const iframeRef = ref(null)
const isMac = ref(false)
const windowState = ref('normal') // 'normal' | 'maximized' | 'fullscreen'
const updateState = (s) => {
  if (!s) return
  if (!s.pinnedUrls) s.pinnedUrls = []
  if (!s.recentUrls) s.recentUrls = []
  if (!s.rotationUrls) s.rotationUrls = []
  if (!s.preloadUrls) s.preloadUrls = []
  if (!s.manualTrustedHosts) s.manualTrustedHosts = []
  if (!s.urlAliases) s.urlAliases = {}
  Object.assign(state, s)
}

function onKeyDown(e) {
  const mod = e.metaKey || e.ctrlKey
  if (!mod) return
  if (e.key === '[' || e.key === 'ArrowLeft') {
    e.preventDefault()
    const rot = effectiveRotation.value
    if (rot.length === 0) return
    handleSwitch(rot[(currentIndex.value - 1 + rot.length) % rot.length])
  } else if (e.key === ']' || e.key === 'ArrowRight') {
    e.preventDefault()
    const rot = effectiveRotation.value
    if (rot.length === 0) return
    handleSwitch(rot[(currentIndex.value + 1) % rot.length])
  }
}

// WKWebView iframe 내 xterm.js에서 compositionend가 안 발화하는 문제 우회.
// 부모 프레임(Wails 네이티브)에서 compositionend를 잡아 iframe으로 postMessage.
function onCompositionStart() {
  iframeRef.value?.contentWindow?.postMessage({ type: 'ime_start' }, '*')
}
function onCompositionEnd(e) {
  const data = e.data
  if (!data || /^[\x20-\x7E]+$/.test(data)) return
  iframeRef.value?.contentWindow?.postMessage({ type: 'ime_input', data }, '*')
}

onMounted(async () => {
  const s = await api.GetState()
  updateState(s)
  isMac.value = await api.IsMacOS()
  windowState.value = await api.GetWindowState()
  window.addEventListener('keydown', onKeyDown)
  window.addEventListener('compositionstart', onCompositionStart)
  window.addEventListener('compositionend', onCompositionEnd)
  // Route target="_blank" / window.open() calls from iframes to system browser.
  // Wails has no multi-window support; iframe popups propagate here when allow="popups" is set.
  window.open = (url) => {
    if (url && url !== 'about:blank' && !String(url).startsWith('javascript:')) {
      api.OpenInBrowser(String(url))
    }
    return null
  }
  resetFade()
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
  window.removeEventListener('compositionstart', onCompositionStart)
  window.removeEventListener('compositionend', onCompositionEnd)
})

// Browser Navigation
const goBack = () => { if (isMac.value) api.ChildGoBack(); else iframeRef.value?.contentWindow?.history.back() }
const goForward = () => { if (isMac.value) api.ChildGoForward(); else iframeRef.value?.contentWindow?.history.forward() }
const reload = () => { if (isMac.value) api.ChildReload(); else if (iframeRef.value) iframeRef.value.src = iframeRef.value.src }
const goHome = () => handleSwitch(state.pinnedUrls[0] || state.serverUrl)

async function handleSwitch(url) {
  if (!url) return
  console.log("[JS] Switching to:", url)
  try {
    const newState = await api.SwitchToUrl(url)
    state.showMenu = false
    state.gridMode = false
    reorderMode.value = false
    // non-mac: iframe src 직접 변경 — :src 바인딩 갱신 전에 미리 이동 시작
    // mac: SwitchToUrl 내부에서 platformNavigate 호출됨
    if (!isMac.value && iframeRef.value) iframeRef.value.src = url
    updateState(newState)
    resetFade()
  } catch (e) {
    console.error("[JS] Switch failed:", e)
    alert("URL 로드 실패: " + e)
  }
}

// Actions
async function handleAdd() {
  const url = addUrlInput.value.trim().startsWith('http') ? addUrlInput.value.trim() : 'http://' + addUrlInput.value.trim()
  if (editingUrl.value) await api.DeleteUrl(editingUrl.value)
  await api.SwitchToUrl(url)
  if (addAliasInput.value.trim()) await api.SetAlias(url, addAliasInput.value.trim())
  updateState(await api.GetState())
  showAddDialog.value = false
  editingUrl.value = null
}
function openEdit(url) { editingUrl.value = url; addUrlInput.value = url; addAliasInput.value = state.urlAliases[url] || ''; showAddDialog.value = true }
async function handlePaste() { try { const text = await navigator.clipboard.readText(); if (text) { addUrlInput.value = text; showAddDialog.value = true; state.showMenu = false } } catch(e){} }
async function handleToggleGrid() { updateState(await api.ToggleGridMode()); state.showMenu = false }
async function handleToggleScroll() { updateState(await api.ToggleScrollLock()); resetFade() }
async function handleTogglePin(url) { updateState(await api.TogglePin(url)) }
async function handleToggleRotation(url) { updateState(await api.ToggleRotation(url)) }
async function handleTogglePreload(url) { updateState(await api.TogglePreload(url)) }
async function handleZoom(delta) { updateState(await api.AdjustZoom(delta)) }
async function handleZoomReset() { updateState(await api.AdjustZoom(100 - effectiveZoom.value)) }
async function handleToggleMaximize() {
  await api.ToggleMaximize()
  windowState.value = await api.GetWindowState()
}
async function handleToggleFullscreen() {
  await api.ToggleFullscreen()
  windowState.value = await api.GetWindowState()
}
async function handleWindowCycle() {
  if (windowState.value === 'normal') {
    await api.ToggleMaximize()
  } else if (windowState.value === 'maximized') {
    await api.ToggleMaximize()
    await api.ToggleFullscreen()
  } else {
    await api.ToggleFullscreen()
  }
  windowState.value = await api.GetWindowState()
}
async function handleMove(url, delta) { updateState(await api.MoveUrl(url, delta)) }
async function handleToggleStatus() { updateState(await api.ToggleStatusBar()) }
async function handleToggleKeepScreen() { updateState(await api.ToggleKeepScreenOn()) }
async function handleToggleRotationBtns() { updateState(await api.ToggleShowRotationBtns()) }
async function handleAddTrust() { if(trustInput.value){ updateState(await api.AddTrustedHost(trustInput.value)); trustInput.value='' } }
async function handleRemoveTrust(h) { updateState(await api.RemoveTrustedHost(h)) }
async function handleDelete(url) { updateState(await api.DeleteUrl(url)) }

// Computed
// 오버레이 메뉴 표시 순서와 일치: pinnedUrls(rotation 포함된 것) → rotationOnly
const effectiveRotation = computed(() => {
  const pins = state.pinnedUrls || []
  const rot = state.rotationUrls || []
  const rotSet = new Set(rot)
  return [
    ...pins.filter(u => rotSet.has(u)),
    ...rot.filter(u => !pins.includes(u)),
  ]
})
const currentIndex = computed(() => effectiveRotation.value.indexOf(state.serverUrl))
// 순환 목록 (pinned 제외) — rotationUrls 순서 유지
const rotationOnly = computed(() => (state.rotationUrls || []).filter(u => !(state.pinnedUrls || []).includes(u)))
// 최근 (pinned도 rotation도 아닌 것만)
const recentOnly = computed(() => (state.recentUrls || []).filter(u => !(state.pinnedUrls || []).includes(u) && !(state.rotationUrls || []).includes(u)))
// 그리드: preloadUrls 우선 → pinnedUrls → rotationUrls
const gridUrls = computed(() => {
  const pre = state.preloadUrls || []
  const pin = state.pinnedUrls || []
  const rot = state.rotationUrls || []
  return pre.length > 0 ? pre : pin.length > 0 ? pin : rot
})
const gridColumns = computed(() => {
  const n = gridUrls.value.length
  if (n <= 1) return '1fr'
  if (n <= 2) return '1fr 1fr'
  if (n <= 4) return '1fr 1fr'
  return '1fr 1fr 1fr'
})
const effectiveZoom = computed(() => {
  if (!state.serverUrl) return state.zoomLevel
  try {
    const hostPort = new URL(state.serverUrl).host
    return state.hostZoomLevels?.[hostPort] ?? state.zoomLevel
  } catch { return state.zoomLevel }
})

// Toolbar Drag & Fade
const toolbarPos = reactive({ x: 0, y: 0 }) // Controlled via CSS transform
const toolbarOpacity = ref(1)
let fadeTimer = null
function resetFade() { toolbarOpacity.value = 1; clearTimeout(fadeTimer); fadeTimer = setTimeout(() => { toolbarOpacity.value = 0.3 }, 3000) }

const isDragging = ref(false); let startPos = { x: 0, y: 0 }
function startDrag(e) { 
  if (e.target.tagName === 'BUTTON' || e.target.closest('button')) return
  isDragging.value = true
  startPos = { x: e.clientX - toolbarPos.x, y: e.clientY - toolbarPos.y }
  resetFade()
}
function onDrag(e) { if (isDragging.value) { toolbarPos.x = e.clientX - startPos.x; toolbarPos.y = e.clientY - startPos.y } }
function stopDrag() { isDragging.value = false }
</script>

<template>
  <div class="main-container" :class="{ 'is-mac': isMac }" @mousemove="onDrag" @mouseup="stopDrag">
    <!-- Webview Wrapper -->
    <div class="webview-wrapper" :class="{ 'scroll-locked': state.scrollLock, 'is-mac': isMac }">
      <!-- Grid Mode -->
      <div v-if="state.gridMode" class="grid-container" :style="{ gridTemplateColumns: gridColumns }">
        <template v-if="gridUrls.length > 0">
          <div v-for="url in gridUrls" :key="url" class="grid-item">
            <div class="grid-label">{{ state.urlAliases[url] || url }}</div>
            <iframe :src="url" class="grid-webview" allow="fullscreen; clipboard-read; clipboard-write; popups"></iframe>
          </div>
        </template>
        <div v-else class="grid-empty">
          <p>표시할 창이 없습니다.</p>
          <p style="font-size:12px; color:#7f849c;">메뉴에서 URL 옆 ⚡ 또는 &lt;&gt; 버튼으로 등록하세요.</p>
        </div>
      </div>
      <!-- Single Mode -->
      <div v-else class="single-webview-container" :class="{ 'is-mac': isMac }">
        <!-- macOS: child WKWebView renders below; iframe not used -->
        <iframe v-if="state.serverUrl && !isMac" ref="iframeRef" :src="state.serverUrl"
                class="main-webview" :style="{ width: (100 / (effectiveZoom / 100)) + '%', height: (100 / (effectiveZoom / 100)) + '%', transform: 'scale(' + (effectiveZoom / 100) + ')', transformOrigin: 'top left' }"
                allow="fullscreen; clipboard-read; clipboard-write; popups"></iframe>
        <div v-if="!state.serverUrl" class="empty-state">
          <h1 class="rainbow-text">JBridgeGo</h1>
          <button @click="state.showMenu = true" class="btn-primary">시작하기 (메뉴 열기)</button>
        </div>
      </div>
      <div v-if="state.scrollLock" class="lock-overlay"></div>
    </div>

    <!-- Floating Toolbar (11 Buttons, Drag, Fade) -->
    <div v-if="state.showRotationBtns" 
         class="floating-toolbar" 
         :style="{ transform: 'translate(calc(-50% + ' + toolbarPos.x + 'px), ' + toolbarPos.y + 'px)', opacity: toolbarOpacity }" 
         @mousedown="startDrag" @mouseenter="resetFade">
      <div class="toolbar-stack">
        <!-- Row 1: Expanded Browser Controls (7 buttons) -->
        <div v-if="expanded" class="toolbar-row top">
          <button @click.stop="state.showMenu = true" title="메뉴">☰</button>
          <div class="divider"></div>
          <button @click.stop="goBack" title="뒤로">◀</button>
          <button @click.stop="goHome" title="홈">🏠</button>
          <button @click.stop="goForward" title="앞으로">▶</button>
          <div class="divider"></div>
          <button @click.stop="reload" title="새로고침">↻</button>
          <button @click.stop="handleToggleScroll" :class="{active: state.scrollLock}" title="스크롤락">{{ state.scrollLock ? '🔒' : '🔓' }}</button>
          <button @click.stop="handleToggleGrid" :class="{active: state.gridMode}" title="그리드">G</button>
          <button @click.stop="handleWindowCycle" :class="{active: windowState !== 'normal'}" :title="windowState === 'normal' ? '최대화' : windowState === 'maximized' ? '전체화면' : '복원'">{{ windowState === 'fullscreen' ? '⊡' : windowState === 'maximized' ? '⛶' : '⊞' }}</button>
        </div>
        <!-- Row 2: Navigation (4 elements) -->
        <div class="toolbar-row">
          <button @click.stop="state.showMenu = true" class="info-btn">{{ expanded ? 'ⓘ' : '☰' }}</button>
          <div class="divider"></div>
          <button @click.stop="handleSwitch(effectiveRotation[(currentIndex - 1 + effectiveRotation.length) % effectiveRotation.length])" title="이전">⟨</button>
          <div class="indicators">{{ currentIndex >= 0 ? currentIndex + 1 : 0 }} / {{ effectiveRotation.length || 0 }}</div>
          <button @click.stop="handleSwitch(effectiveRotation[(currentIndex + 1) % effectiveRotation.length])" title="다음">⟩</button>
          <div class="divider"></div>
          <button @click.stop="expanded = !expanded" class="toggle-btn">{{ expanded ? '✕' : '⋯' }}</button>
        </div>
      </div>
    </div>

    <!-- Main Menu Overlay -->
    <div v-if="state.showMenu" class="overlay-menu" @click.self="state.showMenu = false">
      <div class="menu-content">
        <h2 class="rainbow-text small">JBridgeGo Desktop</h2>
        <div class="scroll-area">
          
          <!-- Pinned + Rotation 공통 순서변경 버튼 -->
          <div class="flex-between" style="margin-bottom:4px;">
            <div></div>
            <span @click="reorderMode = !reorderMode" class="reorder-toggle" :class="{active: reorderMode}">{{ reorderMode ? '완료' : '순서변경' }}</span>
          </div>

          <!-- Pinned Section -->
          <div class="section">
            <div class="section-title">즐겨찾기</div>
            <div v-for="(url, idx) in state.pinnedUrls" :key="url" class="url-row" :class="{ active: url === state.serverUrl }">
              <template v-if="reorderMode">
                <span @click="handleMove(url, -1)" class="move-btn" :class="{disabled: idx === 0}">▲</span>
                <span @click="handleMove(url, 1)" class="move-btn" :class="{disabled: idx === state.pinnedUrls.length - 1}">▼</span>
              </template>
              <template v-else>
                <span @click="handleTogglePin(url)" class="pin-icon active">★</span>
                <span @click="handleToggleRotation(url)" class="rot-icon" :class="{active: state.rotationUrls.includes(url)}">&lt;&gt;</span>
                <span @click="handleTogglePreload(url)" class="pre-icon" :class="{active: state.preloadUrls.includes(url)}">⚡</span>
              </template>
              <div @click="handleSwitch(url)" class="url-info">
                <div class="alias">{{ state.urlAliases[url] || url }}</div>
                <div v-if="state.urlAliases[url]" class="url-sub">{{ url }}</div>
              </div>
              <span v-if="!reorderMode" @click="openEdit(url)" class="edit-btn">✎</span>
              <span v-if="!reorderMode" @click="handleDelete(url)" class="del-btn">✕</span>
            </div>
          </div>

          <!-- Rotation Section -->
          <div v-if="rotationOnly.length > 0" class="section">
            <div class="section-title">순환 목록</div>
            <div v-for="(url, idx) in rotationOnly" :key="url" class="url-row" :class="{ active: url === state.serverUrl }">
              <template v-if="reorderMode">
                <span @click="handleMove(url, -1)" class="move-btn" :class="{disabled: idx === 0}">▲</span>
                <span @click="handleMove(url, 1)" class="move-btn" :class="{disabled: idx === rotationOnly.length - 1}">▼</span>
              </template>
              <template v-else>
                <span @click="handleTogglePin(url)" class="pin-icon">☆</span>
                <span @click="handleToggleRotation(url)" class="rot-icon active">&lt;&gt;</span>
                <span @click="handleTogglePreload(url)" class="pre-icon" :class="{active: state.preloadUrls.includes(url)}">⚡</span>
              </template>
              <div @click="handleSwitch(url)" class="url-info">
                <div class="alias">{{ state.urlAliases[url] || url }}</div>
                <div v-if="state.urlAliases[url]" class="url-sub">{{ url }}</div>
              </div>
              <span v-if="!reorderMode" @click="openEdit(url)" class="edit-btn">✎</span>
              <span v-if="!reorderMode" @click="handleDelete(url)" class="del-btn">✕</span>
            </div>
          </div>

          <!-- Recent Section (rotation/pinned 아닌 것만) -->
          <div v-if="recentOnly.length > 0" class="section">
            <div class="section-title">최근</div>
            <div v-for="url in recentOnly" :key="url" class="url-row" :class="{ active: url === state.serverUrl }">
              <span @click="handleTogglePin(url)" class="pin-icon">☆</span>
              <span @click="handleToggleRotation(url)" class="rot-icon">&lt;&gt;</span>
              <span @click="handleTogglePreload(url)" class="pre-icon" :class="{active: state.preloadUrls.includes(url)}">⚡</span>
              <div @click="handleSwitch(url)" class="url-info">
                <div class="alias">{{ state.urlAliases[url] || url }}</div>
              </div>
              <span @click="openEdit(url)" class="edit-btn">✎</span>
              <span @click="handleDelete(url)" class="del-btn">✕</span>
            </div>
          </div>

          <button @click="showAddDialog = true; editingUrl = null" class="btn-add">+ 새 주소 추가</button>
          
          <div class="divider-h"></div>
          
          <!-- Settings: Zoom -->
          <div class="zoom-controls">
            <span class="section-title" style="margin:0;">글자 크기</span>
            <div style="display:flex; align-items:center; gap:8px;">
              <button @click="handleZoom(-10)" class="zoom-btn">−</button>
              <span class="zoom-val">{{ effectiveZoom }}%</span>
              <span @click="handleZoomReset" class="reset-btn">↺</span>
              <button @click="handleZoom(10)" class="zoom-btn">+</button>
            </div>
          </div>

          <!-- Settings: Toggles -->
          <div class="toggles">
            <div @click="handleToggleStatus" class="toggle-item" :class="{ active: state.hideStatusBar }">
              <span class="icon">{{ state.hideStatusBar ? '▣' : '□' }}</span>
              <span class="label">노티바</span>
            </div>
            <div @click="handleToggleKeepScreen" class="toggle-item" :class="{ active: state.keepScreenOn }">
              <span class="icon">{{ state.keepScreenOn ? '☀' : '☾' }}</span>
              <span class="label">화면유지</span>
            </div>
            <div @click="handleToggleRotationBtns" class="toggle-item" :class="{ active: state.showRotationBtns }">
              <span class="icon">⟨○⟩</span>
              <span class="label">순환버튼</span>
            </div>
            <div @click="handleToggleMaximize" class="toggle-item" :class="{ active: windowState === 'maximized' }">
              <span class="icon">{{ windowState === 'maximized' ? '⊟' : '⊞' }}</span>
              <span class="label">최대화</span>
            </div>
            <div @click="handleToggleFullscreen" class="toggle-item" :class="{ active: windowState === 'fullscreen' }">
              <span class="icon">{{ windowState === 'fullscreen' ? '⊡' : '⛶' }}</span>
              <span class="label">전체화면</span>
            </div>
          </div>

          <!-- Clipboard -->
          <button @click="handlePaste" class="btn-paste">📋 클립보드 붙여넣기</button>

          <div class="divider-h"></div>

          <!-- Trusted Hosts -->
          <div class="section">
            <div class="section-title">신뢰 호스트 관리</div>
            <div v-for="h in state.manualTrustedHosts" :key="h" class="trust-row">
              <span>🔗 {{ h }}</span>
              <span @click="handleRemoveTrust(h)" class="del-btn">✕</span>
            </div>
            <div class="trust-input">
              <input v-model="trustInput" placeholder="host:port" @keyup.enter="handleAddTrust">
              <button @click="handleAddTrust">추가</button>
            </div>
          </div>

        </div>
        <button @click="state.showMenu = false" class="btn-close">닫기</button>
      </div>
    </div>

    <!-- Edit/Add Dialog -->
    <div v-if="showAddDialog" class="overlay-menu" @click.self="showAddDialog = false; editingUrl = null">
      <div class="menu-content dialog">
        <h3 class="rainbow-text small">{{ editingUrl ? '주소 편집' : '새 주소 추가' }}</h3>
        <input v-model="addAliasInput" placeholder="별명 (선택)" class="input-field">
        <input v-model="addUrlInput" placeholder="URL (http://...)" class="input-field" @keyup.enter="handleAdd">
        <div class="dialog-btns">
          <button @click="showAddDialog = false; editingUrl = null" class="btn-text">취소</button>
          <button @click="handleAdd" class="btn-primary">저장</button>
        </div>
      </div>
    </div>

  </div>
</template>

<style scoped>
.main-container { height: 100vh; width: 100vw; background: #1e1e2e; overflow: hidden; position: relative; }
.webview-wrapper { height: 100vh; width: 100vw; position: relative; overflow: hidden; background: #1e1e2e; }
.single-webview-container { width: 100%; height: 100%; position: relative; }
.main-webview { position: absolute; top: 0; left: 0; border: none; background: #1e1e2e; display: block; }

/* macOS: child WKWebView renders below — parent must be fully transparent and non-blocking */
.main-container.is-mac { background: transparent; }
.webview-wrapper.is-mac { background: transparent; }
.single-webview-container.is-mac { background: transparent; pointer-events: none; }
.single-webview-container.is-mac .empty-state { pointer-events: auto; }
.lock-overlay { position: absolute; inset: 0; z-index: 50; background: transparent; }

.grid-container { display: grid; height: 100vh; gap: 4px; padding: 4px; background: #11111b; grid-auto-rows: 1fr; }
.grid-empty { grid-column: 1 / -1; display: flex; flex-direction: column; align-items: center; justify-content: center; height: 200px; color: #cdd6f4; font-size: 14px; gap: 8px; }
.grid-item { display: flex; flex-direction: column; background: white; border-radius: 8px; overflow: hidden; }
.grid-label { background: #313244; color: #cdd6f4; font-size: 10px; padding: 4px 8px; font-weight: bold; }
.grid-webview { flex: 1; border: none; width: 100%; height: 100%; }

.empty-state { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; }

/* Floating Toolbar */
.floating-toolbar { 
  position: fixed; bottom: 30px; left: 50%; /* transform is handled via inline style for dragging */
  background: rgba(24, 24, 37, 0.95); border-radius: 20px; padding: 8px; z-index: 9999;
  border: 2px solid #cba6f7; cursor: move; box-shadow: 0 10px 40px rgba(0,0,0,0.8);
  transition: opacity 0.5s ease; user-select: none;
}
.toolbar-stack { display: flex; flex-direction: column; }
.toolbar-row { display: flex; align-items: center; justify-content: center; gap: 8px; padding: 4px; }
.toolbar-row.top { border-bottom: 1px solid #45475a; margin-bottom: 6px; padding-bottom: 8px; }
.toolbar-row button { 
  background: #313244; border: none; color: white; cursor: pointer; 
  padding: 10px 14px; border-radius: 10px; font-size: 16px; font-weight: bold; transition: background 0.2s;
}
.toolbar-row button:hover { background: #45475a; }
.toolbar-row button.active { background: #cba6f7; color: #1e1e2e; }
.divider { width: 2px; height: 24px; background: #45475a; border-radius: 1px; }
.indicators { color: #cba6f7; font-weight: bold; min-width: 50px; text-align: center; font-family: monospace; font-size: 14px; }
.toggle-btn { color: #cba6f7 !important; }
.info-btn { color: #89dceb !important; }

/* Menus */
.overlay-menu { position: fixed; inset: 0; background: rgba(0,0,0,0.85); display: flex; align-items: center; justify-content: center; z-index: 10000; }
.menu-content { background: #1e1e2e; width: 360px; padding: 24px; border-radius: 20px; border: 1px solid #45475a; color: #cdd6f4; max-height: 90vh; display: flex; flex-direction: column; box-shadow: 0 20px 60px rgba(0,0,0,0.7); }
.menu-content.dialog { width: 320px; display: block; }
.scroll-area { overflow-y: auto; padding-right: 8px; flex: 1; }

/* Menu List Items */
.section { margin-bottom: 16px; }
.flex-between { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.section-title { font-size: 12px; font-weight: bold; color: #89dceb; }
.reorder-toggle { font-size: 11px; color: #7f849c; cursor: pointer; }
.reorder-toggle.active { color: #cba6f7; font-weight: bold; }

.url-row { display: flex; align-items: center; padding: 10px 0; gap: 8px; border-bottom: 1px solid #313244; }
.url-info { flex: 1; min-width: 0; cursor: pointer; }
.alias { font-size: 14px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; color: #cdd6f4; }
.url-row.active .alias { color: #cba6f7; font-weight: bold; }
.url-sub { font-size: 11px; color: #a6adc8; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

.pin-icon { font-size: 18px; color: #6c7086; cursor: pointer; width: 24px; text-align: center; }
.pin-icon.active { color: #f9e2af; }
.rot-icon, .pre-icon { font-size: 12px; color: #6c7086; cursor: pointer; width: 24px; text-align: center; }
.rot-icon.active, .pre-icon.active { color: #cba6f7; font-weight: bold; }

.edit-btn, .del-btn { color: #7f849c; font-size: 14px; padding: 4px; cursor: pointer; }
.del-btn:hover { color: #f38ba8; }
.move-btn { color: #cba6f7; font-size: 16px; cursor: pointer; padding: 0 4px; }
.move-btn.disabled { opacity: 0.2; cursor: default; }

/* Dialog Inputs & Buttons */
.input-field { width: 100%; padding: 12px; background: #313244; border: 1px solid #45475a; color: white; margin-bottom: 12px; border-radius: 8px; box-sizing: border-box; }
.dialog-btns { display: flex; justify-content: flex-end; gap: 12px; margin-top: 8px; }
.btn-text { background: none; border: none; color: #a6adc8; cursor: pointer; font-size: 14px; }
.btn-add { width: 100%; padding: 10px; background: none; border: 1px dashed #a6e3a1; color: #a6e3a1; margin-top: 10px; border-radius: 10px; cursor: pointer; font-size: 13px; }
.btn-paste { width: 100%; padding: 14px; background: rgba(203, 166, 247, 0.1); border: 1px solid rgba(203, 166, 247, 0.3); color: #cba6f7; border-radius: 10px; cursor: pointer; margin-top: 16px; font-weight: bold; }
.btn-primary { background: #cba6f7; color: #1e1e2e; border: none; padding: 12px 24px; border-radius: 10px; font-weight: bold; cursor: pointer; }
.btn-close { width: 100%; padding: 12px; background: #45475a; color: white; border: none; border-radius: 10px; margin-top: 16px; cursor: pointer; font-weight: bold; }

/* Settings */
.zoom-controls { display: flex; align-items: center; justify-content: space-between; padding: 12px 0; }
.zoom-btn { background: #45475a; border: none; color: white; width: 32px; height: 32px; border-radius: 16px; cursor: pointer; font-size: 18px; display: flex; align-items: center; justify-content: center; }
.zoom-val { font-weight: bold; width: 45px; text-align: center; }
.reset-btn { color: #a6adc8; cursor: pointer; font-size: 18px; }

.toggles { display: flex; gap: 8px; margin-top: 8px; }
.toggle-item { flex: 1; display: flex; flex-direction: column; align-items: center; padding: 10px; background: #313244; border-radius: 10px; cursor: pointer; gap: 4px; border: 1px solid transparent; }
.toggle-item.active { background: rgba(203, 166, 247, 0.15); border: 1px solid #cba6f7; }
.toggle-item .icon { font-size: 20px; color: #a6adc8; }
.toggle-item.active .icon { color: #cba6f7; }
.toggle-item .label { font-size: 11px; color: #a6adc8; }

/* Trusted Hosts */
.trust-row { display: flex; justify-content: space-between; font-size: 13px; padding: 8px 0; color: #cdd6f4; border-bottom: 1px solid #313244; }
.trust-input { display: flex; gap: 8px; margin-top: 12px; }
.trust-input input { flex: 1; background: #313244; border: 1px solid #45475a; color: white; padding: 10px; border-radius: 8px; font-size: 13px; }
.trust-input button { background: #cba6f7; color: #1e1e2e; border: none; padding: 0 16px; border-radius: 8px; font-weight: bold; cursor: pointer; }

.divider-h { height: 1px; background: #45475a; margin: 20px 0; }
.rainbow-text { background: linear-gradient(to right, #ff6b6b, #feca57, #a6e3a1, #89dceb, #cba6f7); -webkit-background-clip: text; -webkit-text-fill-color: transparent; font-weight: bold; text-align: center; }
.rainbow-text.small { font-size: 20px; margin: 0 0 20px 0; }
</style>
