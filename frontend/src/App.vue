<script setup>
import { reactive, onMounted, computed } from 'vue'
import { GetState, SwitchToUrl, TogglePin, AdjustZoom, ToggleRotation } from '../wailsjs/go/main/App'

const state = reactive({
  serverUrl: '',
  pinnedUrls: [],
  recentUrls: [],
  rotationUrls: [],
  urlAliases: {},
  zoomLevel: 100,
  showMenu: false,
})

const updateState = (s) => {
  Object.assign(state, s)
}

onMounted(async () => {
  const s = await GetState()
  updateState(s)
})

const currentAlias = computed(() => state.urlAliases[state.serverUrl] || state.serverUrl)

async function handleSwitch(url) {
  const s = await SwitchToUrl(url)
  updateState(s)
  state.showMenu = false
}

async function handleZoom(delta) {
  const s = await AdjustZoom(delta)
  updateState(s)
}

async function handleToggleRotation(url) {
    const s = await ToggleRotation(url)
    updateState(s)
}
</script>

<template>
  <div class="main-container">
    <!-- Webview Area -->
    <div class="webview-wrapper">
      <iframe 
        :src="state.serverUrl" 
        class="main-webview" 
        :style="{ zoom: state.zoomLevel + '%' }"
      ></iframe>
    </div>

    <!-- Floating Toolbar -->
    <div class="floating-toolbar">
      <div class="toolbar-content">
        <button @click="state.showMenu = !state.showMenu" class="btn-menu">☰</button>
        <div class="divider"></div>
        <button @click="handleSwitch(state.rotationUrls[(state.rotationUrls.indexOf(state.serverUrl) - 1 + state.rotationUrls.length) % state.rotationUrls.length])" class="btn-nav">⟨</button>
        <div class="indicators">
          <span v-for="(url, i) in state.rotationUrls" :key="url" 
                :class="{ active: url === state.serverUrl }" 
                class="dot">
            {{ i + 1 }}
          </span>
        </div>
        <button @click="handleSwitch(state.rotationUrls[(state.rotationUrls.indexOf(state.serverUrl) + 1) % state.rotationUrls.length])" class="btn-nav">⟩</button>
      </div>
    </div>

    <!-- Overlay Menu -->
    <div v-if="state.showMenu" class="overlay-menu" @click.self="state.showMenu = false">
      <div class="menu-content">
        <div class="menu-header">JBridgeGo Desktop</div>
        
        <div class="section-title">즐겨찾기</div>
        <div v-for="url in state.pinnedUrls" :key="url" class="url-row" :class="{ active: url === state.serverUrl }">
          <span @click="handleSwitch(url)" class="url-text">{{ state.urlAliases[url] || url }}</span>
          <span @click="handleToggleRotation(url)" :class="{ active: state.rotationUrls.includes(url) }" class="icon-btn">🔄</span>
        </div>

        <div class="section-title">최근</div>
        <div v-for="url in state.recentUrls.filter(u => !state.pinnedUrls.includes(u))" :key="url" class="url-row" :class="{ active: url === state.serverUrl }">
          <span @click="handleSwitch(url)" class="url-text">{{ state.urlAliases[url] || url }}</span>
        </div>

        <div class="divider-h"></div>
        <div class="zoom-control">
          <span>글자 크기: {{ state.zoomLevel }}%</span>
          <div class="zoom-btns">
            <button @click="handleZoom(-10)">−</button>
            <button @click="handleZoom(10)">+</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.main-container {
  height: 100vh;
  display: flex;
  flex-direction: column;
  position: relative;
}

.webview-wrapper {
  flex: 1;
  background: white;
}

.main-webview {
  width: 100%;
  height: 100%;
  border: none;
}

.floating-toolbar {
  position: absolute;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(30, 30, 46, 0.85);
  backdrop-filter: blur(8px);
  border-radius: 12px;
  padding: 8px 16px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
  z-index: 100;
}

.toolbar-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.btn-menu, .btn-nav {
  background: none;
  border: none;
  color: var(--ctp-text);
  font-size: 18px;
  cursor: pointer;
  padding: 4px 8px;
}

.btn-menu:hover, .btn-nav:hover {
  color: var(--ctp-mauve);
}

.divider {
  width: 1px;
  height: 20px;
  background: var(--ctp-surface1);
}

.indicators {
  display: flex;
  gap: 6px;
}

.dot {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: var(--ctp-surface1);
  font-size: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dot.active {
  background: var(--ctp-mauve);
  color: var(--ctp-base);
}

.overlay-menu {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(17, 17, 27, 0.7);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 200;
}

.menu-content {
  background: var(--ctp-base);
  width: 320px;
  padding: 24px;
  border-radius: 16px;
  border: 1px solid var(--ctp-surface1);
}

.menu-header {
  font-size: 18px;
  font-weight: bold;
  color: var(--ctp-mauve);
  margin-bottom: 20px;
  text-align: center;
}

.section-title {
  font-size: 12px;
  color: var(--ctp-sky);
  margin-top: 16px;
  margin-bottom: 8px;
  font-weight: bold;
}

.url-row {
  display: flex;
  justify-content: space-between;
  padding: 8px 0;
  cursor: pointer;
  border-bottom: 0.5px solid var(--ctp-surface0);
}

.url-row.active .url-text {
  color: var(--ctp-mauve);
}

.url-text {
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}

.icon-btn {
  font-size: 14px;
  filter: grayscale(1);
  margin-left: 8px;
}

.icon-btn.active {
  filter: grayscale(0);
}

.zoom-control {
  margin-top: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.zoom-btns button {
  background: var(--ctp-surface1);
  border: none;
  color: var(--ctp-text);
  padding: 4px 12px;
  margin-left: 8px;
  border-radius: 4px;
  cursor: pointer;
}
</style>
