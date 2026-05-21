import Store from 'electron-store'
import { AppState, defaultState } from '../shared/state'

const store = new Store<{ state: AppState }>({
  name: 'settings',
  defaults: { state: defaultState() }
})

export function loadState(): AppState {
  const s = store.get('state')
  return { ...defaultState(), ...s }
}

export function saveState(state: AppState): void {
  store.set('state', state)
}
