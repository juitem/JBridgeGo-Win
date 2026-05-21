import type { JBridgeApi } from '../shared/state'

declare global {
  interface Window {
    api: JBridgeApi
  }
}

export {}
