import { BrowserWindow, ipcMain, powerSaveBlocker, shell } from 'electron'
import { AppState, GridLayout, WindowMode, gridDims } from '../shared/state'
import { loadState, saveState } from './store'

let state: AppState = loadState()
let powerSaveBlockerId: number | null = null

function applyKeepScreenOn(): void {
  if (state.keepScreenOn) {
    if (powerSaveBlockerId === null || !powerSaveBlocker.isStarted(powerSaveBlockerId)) {
      powerSaveBlockerId = powerSaveBlocker.start('prevent-display-sleep')
    }
  } else if (powerSaveBlockerId !== null) {
    if (powerSaveBlocker.isStarted(powerSaveBlockerId)) {
      powerSaveBlocker.stop(powerSaveBlockerId)
    }
    powerSaveBlockerId = null
  }
}

function commit(): AppState {
  saveState(state)
  return state
}

export function syncOsSideEffects(): void {
  applyKeepScreenOn()
}

function hostOf(url: string): string {
  try {
    return new URL(url).host
  } catch {
    return ''
  }
}

function normalizeRotation(pinned: string[], rotation: string[]): string[] {
  const pinSet = new Set(pinned)
  const rotSet = new Set(rotation)
  const out: string[] = []
  for (const u of pinned) if (rotSet.has(u)) out.push(u)
  for (const u of rotation) if (!pinSet.has(u)) out.push(u)
  return out
}

export function registerIpc(getWin: () => BrowserWindow | null): void {
  ipcMain.handle('state:get', () => state)

  ipcMain.handle('state:switchToUrl', (_e, url: string) => {
    state.serverUrl = url
    state.setupComplete = true
    const newRecent = [url, ...state.recentUrls.filter((u) => u !== url)].slice(0, 10)
    state.recentUrls = newRecent
    if (!state.rotationUrls.includes(url)) state.rotationUrls.push(url)
    return commit()
  })

  ipcMain.handle('state:deleteUrl', (_e, url: string) => {
    const filter = (l: string[]): string[] => l.filter((u) => u !== url)
    state.pinnedUrls = filter(state.pinnedUrls)
    state.recentUrls = filter(state.recentUrls)
    state.rotationUrls = filter(state.rotationUrls)
    state.preloadUrls = filter(state.preloadUrls)
    delete state.urlAliases[url]
    if (state.serverUrl === url) {
      state.serverUrl = state.pinnedUrls[0] ?? state.recentUrls[0] ?? ''
    }
    return commit()
  })

  ipcMain.handle('state:togglePin', (_e, url: string) => {
    state.pinnedUrls = state.pinnedUrls.includes(url)
      ? state.pinnedUrls.filter((u) => u !== url)
      : [...state.pinnedUrls, url]
    return commit()
  })

  ipcMain.handle('state:toggleRotation', (_e, url: string) => {
    state.rotationUrls = state.rotationUrls.includes(url)
      ? state.rotationUrls.filter((u) => u !== url)
      : [...state.rotationUrls, url]
    return commit()
  })

  ipcMain.handle('state:togglePreload', (_e, url: string) => {
    state.preloadUrls = state.preloadUrls.includes(url)
      ? state.preloadUrls.filter((u) => u !== url)
      : [...state.preloadUrls, url]
    return commit()
  })

  ipcMain.handle('state:adjustZoom', (_e, delta: number) => {
    const host = hostOf(state.serverUrl)
    const current = host ? (state.hostZoomLevels[host] ?? state.zoomLevel) : state.zoomLevel
    const next = Math.max(50, Math.min(300, current + delta))
    if (host) state.hostZoomLevels[host] = next
    else state.zoomLevel = next
    return commit()
  })

  ipcMain.handle('state:setAlias', (_e, url: string, alias: string) => {
    if (alias) state.urlAliases[url] = alias
    else delete state.urlAliases[url]
    return commit()
  })

  ipcMain.handle('state:toggleGridMode', () => {
    state.gridMode = !state.gridMode
    return commit()
  })

  ipcMain.handle('state:moveUrl', (_e, url: string, delta: number) => {
    const swap = (list: string[]): string[] => {
      const idx = list.indexOf(url)
      if (idx < 0) return list
      const ni = idx + delta
      if (ni < 0 || ni >= list.length) return list
      const out = [...list]
      ;[out[idx], out[ni]] = [out[ni], out[idx]]
      return out
    }
    const isPinned = state.pinnedUrls.includes(url)
    if (isPinned) {
      state.pinnedUrls = swap(state.pinnedUrls)
    } else {
      const pinSet = new Set(state.pinnedUrls)
      const rotOnly = swap(state.rotationUrls.filter((u) => !pinSet.has(u)))
      let j = 0
      const out = [...state.rotationUrls]
      for (let i = 0; i < out.length; i++) {
        if (!pinSet.has(out[i])) {
          out[i] = rotOnly[j++]
        }
      }
      state.rotationUrls = out
    }
    state.rotationUrls = normalizeRotation(state.pinnedUrls, state.rotationUrls)
    return commit()
  })

  ipcMain.handle('state:addTrustedHost', (_e, host: string) => {
    if (!state.manualTrustedHosts.includes(host)) state.manualTrustedHosts.push(host)
    return commit()
  })

  ipcMain.handle('state:removeTrustedHost', (_e, host: string) => {
    state.manualTrustedHosts = state.manualTrustedHosts.filter((h) => h !== host)
    return commit()
  })

  ipcMain.handle('state:toggleScrollLock', () => {
    state.scrollLock = !state.scrollLock
    return commit()
  })

  ipcMain.handle('state:toggleStatusBar', () => {
    state.hideStatusBar = !state.hideStatusBar
    return commit()
  })

  ipcMain.handle('state:toggleKeepScreenOn', () => {
    state.keepScreenOn = !state.keepScreenOn
    applyKeepScreenOn()
    return commit()
  })

  ipcMain.handle('state:toggleShowRotationBtns', () => {
    state.showRotationBtns = !state.showRotationBtns
    return commit()
  })

  ipcMain.handle('state:setGridLayout', (_e, layout: GridLayout) => {
    state.gridLayout = layout
    const dims = gridDims(layout)
    const need = dims ? dims.rows * dims.cols : 0
    const slots = [...state.gridSlots]
    if (slots.length < need) slots.push(...Array(need - slots.length).fill(''))
    if (slots.length > need) slots.length = need
    state.gridSlots = slots
    return commit()
  })

  ipcMain.handle('state:setGridSlot', (_e, index: number, url: string) => {
    const slots = [...state.gridSlots]
    while (slots.length <= index) slots.push('')
    slots[index] = url
    state.gridSlots = slots
    return commit()
  })

  ipcMain.handle('window:toggleMaximize', () => {
    const win = getWin()
    if (!win) return false
    if (win.isMaximized()) {
      win.unmaximize()
      return false
    }
    win.maximize()
    return true
  })

  ipcMain.handle('window:toggleFullscreen', () => {
    const win = getWin()
    if (!win) return false
    const next = !win.isFullScreen()
    win.setFullScreen(next)
    return next
  })

  ipcMain.handle('window:getState', (): WindowMode => {
    const win = getWin()
    if (!win) return 'normal'
    if (win.isFullScreen()) return 'fullscreen'
    if (win.isMaximized()) return 'maximized'
    return 'normal'
  })

  ipcMain.handle('shell:openInBrowser', (_e, url: string) => {
    void shell.openExternal(url)
  })
}
