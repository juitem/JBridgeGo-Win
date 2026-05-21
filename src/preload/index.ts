import { contextBridge, ipcRenderer } from 'electron'
import type { JBridgeApi } from '../shared/state'

const api: JBridgeApi = {
  getState: () => ipcRenderer.invoke('state:get'),
  switchToUrl: (url) => ipcRenderer.invoke('state:switchToUrl', url),
  deleteUrl: (url) => ipcRenderer.invoke('state:deleteUrl', url),
  togglePin: (url) => ipcRenderer.invoke('state:togglePin', url),
  toggleRotation: (url) => ipcRenderer.invoke('state:toggleRotation', url),
  togglePreload: (url) => ipcRenderer.invoke('state:togglePreload', url),
  adjustZoom: (delta) => ipcRenderer.invoke('state:adjustZoom', delta),
  setAlias: (url, alias) => ipcRenderer.invoke('state:setAlias', url, alias),
  toggleGridMode: () => ipcRenderer.invoke('state:toggleGridMode'),
  moveUrl: (url, delta) => ipcRenderer.invoke('state:moveUrl', url, delta),
  addTrustedHost: (host) => ipcRenderer.invoke('state:addTrustedHost', host),
  removeTrustedHost: (host) => ipcRenderer.invoke('state:removeTrustedHost', host),
  toggleScrollLock: () => ipcRenderer.invoke('state:toggleScrollLock'),
  toggleStatusBar: () => ipcRenderer.invoke('state:toggleStatusBar'),
  toggleKeepScreenOn: () => ipcRenderer.invoke('state:toggleKeepScreenOn'),
  toggleShowRotationBtns: () => ipcRenderer.invoke('state:toggleShowRotationBtns'),
  setGridLayout: (layout) => ipcRenderer.invoke('state:setGridLayout', layout),
  setGridSlot: (index, url) => ipcRenderer.invoke('state:setGridSlot', index, url),
  toggleMaximize: () => ipcRenderer.invoke('window:toggleMaximize'),
  toggleFullscreen: () => ipcRenderer.invoke('window:toggleFullscreen'),
  getWindowState: () => ipcRenderer.invoke('window:getState'),
  openInBrowser: (url) => ipcRenderer.invoke('shell:openInBrowser', url)
}

contextBridge.exposeInMainWorld('api', api)
