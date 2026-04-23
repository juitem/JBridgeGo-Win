<script setup>
import { reactive, onMounted, computed, ref } from 'vue'
import * as api from '../wailsjs/go/main/App'

const state = reactive({
  serverUrl: '',
  hideStatusBar: true,
  zoomLevel: 100,
  pinnedUrls: [],
  recentUrls: [],
  rotationUrls: [],
  preloadUrls: [],
  urlAliases: {},
  keepScreenOn: false,
  manualTrustedHosts: [],
  showRotationBtns: true,
  showMenu: false,
})

const showAddDialog = ref(false)
const addUrlInput = ref('')
const addAliasInput = ref('')
const editingUrl = ref(null)

const updateState = (s) => { if (s) Object.assign(state, s) }

onMounted(async () => {
  const s = await api.GetState()
  updateState(s)
  if (!s.serverUrl && s.pinnedUrls.length > 0) handleSwitch(s.pinnedUrls[0])
})

const recentOnly = computed(() => state.recentUrls.filter(u => !state.pinnedUrls.includes(u)))

async function handleSwitch(url) {
  const s = await api.SwitchToUrl(url)
  updateState(s)
  state.showMenu = false
}

async function handleAdd() {
  if (!addUrlInput.value.trim()) return
  const url = addUrlInput.value.trim().startsWith('http') ? addUrlInput.value.trim() : 'http://' + addUrlInput.value.trim()
  const s = await api.SwitchToUrl(url)
  if (addAliasInput.value.trim()) await api.SetAlias(url, addAliasInput.value.trim())
  const finalS = await api.GetState()
  updateState(finalS)
  showAddDialog.value = false
  addUrlInput.value = ''
  addAliasInput.value = ''
}

async function handleDelete(url) {
  if (confirm('이 주소를 삭제하시겠습니까?')) {
    const s = await api.DeleteUrl(url)
    updateState(s)
  }
}

async function handleTogglePin(url) { updateState(await api.TogglePin(url)) }
async function handleToggleRotation(url) { updateState(await api.ToggleRotation(url)) }
async function handleTogglePreload(url) { updateState(await api.TogglePreload(url)) }
async function handleZoom(delta) { updateState(await api.AdjustZoom(delta)) }
async function handleToggleStatus() { updateState(await api.ToggleStatusBar()) }
async function handleToggleKeepScreen() { updateState(await api.ToggleKeepScreenOn()) }
async function handleToggleRotationBtns() { updateState(await api.ToggleShowRotationBtns()) }

async function handlePaste() {
  try {
    const text = await navigator.clipboard.readText()
    if (text) {
      addUrlInput.value = text
      showAddDialog.value = true
      state.showMenu = false
    }
  } catch (err) { alert('클립보드 접근 권한이 필요합니다.') }
}

function openEdit(url) {
  editingUrl.value = url
  addUrlInput.value = url
  addAliasInput.value = state.urlAliases[url] || ''
  showAddDialog.value = true
}
</script>

<template>
  <div class="main-container">
    <div class="webview-wrapper">
      <iframe v-if="state.serverUrl" :src="state.serverUrl" class="main-webview" :style="{ zoom: state.zoomLevel + '%' }"></iframe>
      <div v-else class="empty-state">
        <h2 style="color:var(--ctp-mauve)">JBridgeGo Desktop</h2>
        <p>연결된 주소가 없습니다. 메뉴에서 주소를 추가해 주세요.</p>
        <button @click="state.showMenu = true" class="btn-primary">메뉴 열기</button>
      </div>
    </div>

    <div v-if="state.showRotationBtns && state.rotationUrls.length >= 2" class="floating-toolbar">
      <div class="toolbar-content">
        <button @click="state.showMenu = true" class="btn-icon">☰</button>
        <div class="divider-v"></div>
        <button @click="handleSwitch(state.rotationUrls[(state.rotationUrls.indexOf(state.serverUrl) - 1 + state.rotationUrls.length) % state.rotationUrls.length])" class="btn-nav">⟨</button>
        <div class="indicators">
          <span v-for="(url, i) in state.rotationUrls" :key="url" :class="{ active: url === state.serverUrl }" class="dot">{{ i + 1 }}</span>
        </div>
        <button @click="handleSwitch(state.rotationUrls[(state.rotationUrls.indexOf(state.serverUrl) + 1) % state.rotationUrls.length])" class="btn-nav">⟩</button>
      </div>
    </div>
    <button v-else-if="!state.showMenu" @click="state.showMenu = true" class="fab-menu">☰</button>

    <div v-if="state.showMenu" class="overlay-menu" @click.self="state.showMenu = false">
      <div class="menu-content">
        <div class="menu-header">JBridgeGo Desktop</div>
        
        <div class="scroll-area">
          <div v-if="state.pinnedUrls.length > 0" class="section">
            <div class="section-title">즐겨찾기</div>
            <div v-for="url in state.pinnedUrls" :key="url" class="url-row" :class="{ active: url === state.serverUrl }">
              <span @click="handleTogglePin(url)" class="pin-icon active">★</span>
              <span @click="handleToggleRotation(url)" class="rot-icon" :class="{ active: state.rotationUrls.includes(url) }">&lt;&gt;</span>
              <span @click="handleTogglePreload(url)" class="pre-icon" :class="{ active: state.preloadUrls.includes(url) }">⚡</span>
              <div @click="handleSwitch(url)" class="url-info">
                <div class="alias">{{ state.urlAliases[url] || url }}</div>
                <div v-if="state.urlAliases[url]" class="url-sub">{{ url }}</div>
              </div>
              <span @click="openEdit(url)" class="edit-btn">✎</span>
              <span @click="handleDelete(url)" class="del-btn">✕</span>
            </div>
          </div>

          <div v-if="recentOnly.length > 0" class="section">
            <div class="section-title">최근</div>
            <div v-for="url in recentOnly" :key="url" class="url-row" :class="{ active: url === state.serverUrl }">
              <span @click="handleTogglePin(url)" class="pin-icon">☆</span>
              <span @click="handleToggleRotation(url)" class="rot-icon" :class="{ active: state.rotationUrls.includes(url) }">&lt;&gt;</span>
              <span @click="handleTogglePreload(url)" class="pre-icon" :class="{ active: state.preloadUrls.includes(url) }">⚡</span>
              <div @click="handleSwitch(url)" class="url-info">
                <div class="alias">{{ state.urlAliases[url] || url }}</div>
                <div v-if="state.urlAliases[url]" class="url-sub">{{ url }}</div>
              </div>
              <span @click="openEdit(url)" class="edit-btn">✎</span>
              <span @click="handleDelete(url)" class="del-btn">✕</span>
            </div>
          </div>

          <button @click="showAddDialog = true; editingUrl = null" class="btn-add">+ 새 주소 추가</button>
          
          <div class="divider-h"></div>
          <div class="zoom-section">
            <div class="section-title">글자 크기</div>
            <div class="zoom-controls">
              <button @click="handleZoom(-10)">−</button>
              <span class="zoom-val">{{ state.zoomLevel }}%</span>
              <button @click="handleZoom(10)">+</button>
            </div>
          </div>

          <div class="divider-h"></div>
          <div class="toggles">
            <div @click="handleToggleStatus" class="toggle-item" :class="{ active: state.hideStatusBar }">
              <span class="icon">{{ state.hideStatusBar ? '▣' : '□' }}</span>
              <span class="label">{{ state.hideStatusBar ? '노티바숨김' : '노티바표시' }}</span>
            </div>
            <div @click="handleToggleKeepScreen" class="toggle-item" :class="{ active: state.keepScreenOn }">
              <span class="icon">{{ state.keepScreenOn ? '☀' : '☾' }}</span>
              <span class="label">화면유지</span>
            </div>
            <div @click="handleToggleRotationBtns" class="toggle-item" :class="{ active: state.showRotationBtns }">
              <span class="icon">⟨○⟩</span>
              <span class="label">순환버튼</span>
            </div>
          </div>

          <button @click="handlePaste" class="btn-paste">📋 클립보드 붙여넣기</button>
        </div>
        
        <button @click="state.showMenu = false" class="btn-close">닫기</button>
      </div>
    </div>

    <div v-if="showAddDialog" class="overlay-menu" @click.self="showAddDialog = false">
      <div class="menu-content dialog">
        <div class="menu-header">{{ editingUrl ? '주소 편집' : '새 주소 추가' }}</div>
        <input v-model="addAliasInput" placeholder="별명 (선택)" class="input-field">
        <input v-model="addUrlInput" placeholder="http://your-server:7800" class="input-field">
        <div class="dialog-btns">
          <button @click="showAddDialog = false" class="btn-text">취소</button>
          <button @click="handleAdd" class="btn-primary">저장</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.main-container { height: 100vh; position: relative; background: var(--ctp-base); }
.webview-wrapper { height: 100vh; width: 100vw; }
.main-webview { width: 100%; height: 100%; border: none; background: white; }
.empty-state { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 16px; }

.floating-toolbar { position: absolute; bottom: 24px; left: 50%; transform: translateX(-50%); background: rgba(30, 30, 46, 0.9); backdrop-filter: blur(10px); border-radius: 16px; padding: 6px 16px; box-shadow: 0 8px 32px rgba(0,0,0,0.4); z-index: 100; border: 1px solid var(--ctp-surface1); }
.toolbar-content { display: flex; align-items: center; gap: 12px; }
.fab-menu { position: absolute; bottom: 24px; right: 24px; width: 48px; height: 48px; border-radius: 24px; background: var(--ctp-mauve); color: var(--ctp-base); border: none; font-size: 20px; cursor: pointer; z-index: 100; box-shadow: 0 4px 12px rgba(0,0,0,0.3); }

.btn-icon, .btn-nav { background: none; border: none; color: var(--ctp-text); font-size: 18px; cursor: pointer; padding: 8px; }
.divider-v { width: 1px; height: 24px; background: var(--ctp-surface1); }
.indicators { display: flex; gap: 6px; }
.dot { width: 22px; height: 22px; border-radius: 11px; background: var(--ctp-surface1); font-size: 11px; display: flex; align-items: center; justify-content: center; color: var(--ctp-text); }
.dot.active { background: var(--ctp-mauve); color: var(--ctp-base); font-weight: bold; }

.overlay-menu { position: absolute; inset: 0; background: rgba(17, 17, 27, 0.8); display: flex; align-items: center; justify-content: center; z-index: 200; }
.menu-content { background: var(--ctp-base); width: 360px; max-height: 85vh; border-radius: 20px; border: 1px solid var(--ctp-surface1); display: flex; flex-direction: column; padding: 20px; box-shadow: 0 20px 50px rgba(0,0,0,0.5); }
.menu-content.dialog { width: 320px; gap: 12px; }
.scroll-area { flex: 1; overflow-y: auto; padding-right: 8px; }
.menu-header { font-size: 20px; font-weight: bold; color: var(--ctp-mauve); margin-bottom: 20px; text-align: center; }

.section { margin-bottom: 20px; }
.section-title { font-size: 12px; font-weight: bold; color: var(--ctp-sky); margin-bottom: 8px; }
.url-row { display: flex; align-items: center; padding: 10px 0; border-bottom: 1px solid var(--ctp-surface0); gap: 10px; cursor: pointer; }
.url-row:hover { background: var(--ctp-surface0); }
.url-row.active .alias { color: var(--ctp-mauve); }

.pin-icon { color: var(--ctp-overlay0); font-size: 18px; width: 24px; }
.pin-icon.active { color: var(--ctp-yellow); }
.rot-icon, .pre-icon { font-size: 12px; color: var(--ctp-overlay0); width: 24px; text-align: center; }
.rot-icon.active, .pre-icon.active { color: var(--ctp-mauve); font-weight: bold; }

.url-info { flex: 1; min-width: 0; }
.alias { font-size: 14px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.url-sub { font-size: 11px; color: var(--ctp-overlay1); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.edit-btn, .del-btn { color: var(--ctp-overlay1); font-size: 14px; padding: 4px; }
.del-btn:hover { color: var(--ctp-red); }

.btn-add { width: 100%; padding: 12px; background: none; border: 1px dashed var(--ctp-green); color: var(--ctp-green); border-radius: 10px; cursor: pointer; margin: 12px 0; }
.btn-paste { width: 100%; padding: 14px; background: rgba(203, 166, 247, 0.1); border: none; color: var(--ctp-text); border-radius: 12px; cursor: pointer; margin-top: 16px; font-size: 14px; }

.zoom-controls { display: flex; align-items: center; justify-content: space-between; padding: 10px 0; }
.zoom-controls button { width: 36px; height: 36px; border-radius: 18px; border: none; background: var(--ctp-surface1); color: var(--ctp-text); font-size: 20px; cursor: pointer; }
.zoom-val { font-size: 16px; font-weight: bold; }

.toggles { display: flex; gap: 8px; margin-top: 16px; }
.toggle-item { flex: 1; display: flex; flex-direction: column; align-items: center; padding: 10px; border-radius: 10px; background: var(--ctp-surface0); cursor: pointer; gap: 4px; }
.toggle-item.active { background: rgba(203, 166, 247, 0.2); }
.toggle-item.active .icon, .toggle-item.active .label { color: var(--ctp-mauve); }
.toggle-item .icon { font-size: 20px; color: var(--ctp-overlay1); }
.toggle-item .label { font-size: 10px; color: var(--ctp-overlay1); }

.input-field { background: var(--ctp-surface0); border: 1px solid var(--ctp-surface1); color: var(--ctp-text); padding: 12px; border-radius: 8px; width: 100%; box-sizing: border-box; }
.dialog-btns { display: flex; justify-content: flex-end; gap: 12px; margin-top: 12px; }
.btn-text { background: none; border: none; color: var(--ctp-overlay1); cursor: pointer; }
.btn-primary { background: var(--ctp-mauve); color: var(--ctp-base); border: none; padding: 8px 20px; border-radius: 8px; font-weight: bold; cursor: pointer; }
.btn-close { width: 100%; padding: 12px; margin-top: 20px; background: var(--ctp-surface1); border: none; color: var(--ctp-text); border-radius: 10px; cursor: pointer; }

.divider-h { height: 1px; background: var(--ctp-surface1); margin: 16px 0; }
</style>
