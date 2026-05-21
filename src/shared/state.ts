export interface AppState {
  serverUrl: string
  hideStatusBar: boolean
  zoomLevel: number
  pinnedUrls: string[]
  recentUrls: string[]
  setupComplete: boolean
  keepScreenOn: boolean
  orientation: number
  rotationUrls: string[]
  showRotationBtns: boolean
  preloadUrls: string[]
  urlAliases: Record<string, string>
  scrollLock: boolean
  gridMode: boolean
  manualTrustedHosts: string[]
  hostZoomLevels: Record<string, number>
}

export function defaultState(): AppState {
  return {
    serverUrl: 'http://your-server:7800',
    hideStatusBar: true,
    zoomLevel: 100,
    pinnedUrls: [],
    recentUrls: [],
    setupComplete: false,
    keepScreenOn: false,
    orientation: 0,
    rotationUrls: [],
    showRotationBtns: true,
    preloadUrls: [],
    urlAliases: {},
    scrollLock: false,
    gridMode: false,
    manualTrustedHosts: [],
    hostZoomLevels: {}
  }
}

export type WindowMode = 'normal' | 'maximized' | 'fullscreen'

export interface JBridgeApi {
  getState(): Promise<AppState>
  switchToUrl(url: string): Promise<AppState>
  deleteUrl(url: string): Promise<AppState>
  togglePin(url: string): Promise<AppState>
  toggleRotation(url: string): Promise<AppState>
  togglePreload(url: string): Promise<AppState>
  adjustZoom(delta: number): Promise<AppState>
  setAlias(url: string, alias: string): Promise<AppState>
  toggleGridMode(): Promise<AppState>
  moveUrl(url: string, delta: number): Promise<AppState>
  addTrustedHost(host: string): Promise<AppState>
  removeTrustedHost(host: string): Promise<AppState>
  toggleScrollLock(): Promise<AppState>
  toggleStatusBar(): Promise<AppState>
  toggleKeepScreenOn(): Promise<AppState>
  toggleShowRotationBtns(): Promise<AppState>
  toggleMaximize(): Promise<boolean>
  toggleFullscreen(): Promise<boolean>
  getWindowState(): Promise<WindowMode>
  openInBrowser(url: string): Promise<void>
}
