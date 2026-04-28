package main

import (
	"context"
	"jbridgego-win/internal/state"
	"net/url"
	"sync"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx     context.Context
	state   *state.AppState
	storage *state.Storage
	mu      sync.Mutex
}

func NewApp() *App {
	storage, _ := state.NewStorage()
	s, _ := storage.Load()
	return &App{
		state:   s,
		storage: storage,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	platformSetup()
}

func (a *App) GetState() *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state
}

func (a *App) SwitchToUrl(targetUrl string) *state.AppState {
	println("[Go] Switching to URL:", targetUrl)
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.ServerURL = targetUrl
	a.state.SetupComplete = true
	platformNavigate(targetUrl)
	// 최근 목록 업데이트
	newRecent := []string{targetUrl}
	for _, u := range a.state.RecentUrls {
		if u != targetUrl { newRecent = append(newRecent, u) }
	}
	if len(newRecent) > 10 { newRecent = newRecent[:10] }
	a.state.RecentUrls = newRecent
	
	foundRot := false
	for _, u := range a.state.RotationUrls { if u == targetUrl { foundRot = true; break } }
	if !foundRot { a.state.RotationUrls = append(a.state.RotationUrls, targetUrl) }
    
	a.storage.Save(a.state)
	return a.state
}

func (a *App) DeleteUrl(targetUrl string) *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	filter := func(list []string) []string {
		res := []string{}
		for _, u := range list { if u != targetUrl { res = append(res, u) } }
		return res
	}
	a.state.PinnedUrls = filter(a.state.PinnedUrls)
	a.state.RecentUrls = filter(a.state.RecentUrls)
	a.state.RotationUrls = filter(a.state.RotationUrls)
	a.state.PreloadUrls = filter(a.state.PreloadUrls)
	delete(a.state.UrlAliases, targetUrl)
	
	if a.state.ServerURL == targetUrl {
		a.state.ServerURL = ""
		if len(a.state.PinnedUrls) > 0 { 
			a.state.ServerURL = a.state.PinnedUrls[0] 
		} else if len(a.state.RecentUrls) > 0 { 
			a.state.ServerURL = a.state.RecentUrls[0] 
		}
	}
	a.storage.Save(a.state)
	return a.state
}

func (a *App) TogglePin(targetUrl string) *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	found := false
	newPins := []string{}
	for _, u := range a.state.PinnedUrls {
		if u == targetUrl { found = true; continue }
		newPins = append(newPins, u)
	}
	if !found { newPins = append(newPins, targetUrl) }
	a.state.PinnedUrls = newPins
	a.storage.Save(a.state)
	return a.state
}

func (a *App) ToggleRotation(targetUrl string) *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	found := false
	newList := []string{}
	for _, u := range a.state.RotationUrls {
		if u == targetUrl { found = true; continue }
		newList = append(newList, u)
	}
	if !found { newList = append(newList, targetUrl) }
	a.state.RotationUrls = newList
	a.storage.Save(a.state)
	return a.state
}

func (a *App) TogglePreload(targetUrl string) *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	found := false
	newList := []string{}
	for _, u := range a.state.PreloadUrls {
		if u == targetUrl { found = true; continue }
		newList = append(newList, u)
	}
	if !found { newList = append(newList, targetUrl) }
	a.state.PreloadUrls = newList
	a.storage.Save(a.state)
	return a.state
}

func (a *App) AdjustZoom(delta int) *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	hostPort := extractHostPort(a.state.ServerURL)
	currentZoom := a.state.ZoomLevel
	if hostPort != "" {
		if z, ok := a.state.HostZoomLevels[hostPort]; ok { currentZoom = z }
	}
	newZoom := currentZoom + delta
	if newZoom < 50 { newZoom = 50 }
	if newZoom > 300 { newZoom = 300 }
	if hostPort != "" { a.state.HostZoomLevels[hostPort] = newZoom } else { a.state.ZoomLevel = newZoom }
	a.storage.Save(a.state)
	return a.state
}

func (a *App) SetAlias(targetUrl string, alias string) *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	if alias == "" { delete(a.state.UrlAliases, targetUrl) } else { a.state.UrlAliases[targetUrl] = alias }
	a.storage.Save(a.state)
	return a.state
}

func (a *App) ToggleGridMode() *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.GridMode = !a.state.GridMode
	a.storage.Save(a.state)
	return a.state
}

func (a *App) MoveUrl(targetUrl string, delta int) *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()

	swapIn := func(list []string) []string {
		idx := -1
		for i, u := range list { if u == targetUrl { idx = i; break } }
		if idx == -1 { return list }
		newIdx := idx + delta
		if newIdx < 0 || newIdx >= len(list) { return list }
		newList := make([]string, len(list))
		copy(newList, list)
		newList[idx], newList[newIdx] = newList[newIdx], newList[idx]
		return newList
	}

	isPinned := false
	for _, u := range a.state.PinnedUrls { if u == targetUrl { isPinned = true; break } }

	if isPinned {
		a.state.PinnedUrls = swapIn(a.state.PinnedUrls)
	} else {
		pinnedSet := map[string]bool{}
		for _, u := range a.state.PinnedUrls { pinnedSet[u] = true }
		rotOnly := []string{}
		for _, u := range a.state.RotationUrls {
			if !pinnedSet[u] { rotOnly = append(rotOnly, u) }
		}
		rotOnly = swapIn(rotOnly)
		// pinned 항목 자리는 그대로 두고 non-pinned 슬롯만 교체
		j := 0
		for i, u := range a.state.RotationUrls {
			if !pinnedSet[u] { a.state.RotationUrls[i] = rotOnly[j]; j++ }
		}
	}
	// rotationUrls 정규화: pinnedUrls 순 → rotationOnly 순
	a.state.RotationUrls = normalizeRotation(a.state.PinnedUrls, a.state.RotationUrls)

	a.storage.Save(a.state)
	return a.state
}

func (a *App) AddTrustedHost(host string) *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, h := range a.state.ManualTrustedHosts { if h == host { return a.state } }
	a.state.ManualTrustedHosts = append(a.state.ManualTrustedHosts, host)
	a.storage.Save(a.state)
	return a.state
}

func (a *App) RemoveTrustedHost(host string) *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	newList := []string{}
	for _, h := range a.state.ManualTrustedHosts { if h != host { newList = append(newList, h) } }
	a.state.ManualTrustedHosts = newList
	a.storage.Save(a.state)
	return a.state
}

func (a *App) ToggleScrollLock() *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.ScrollLock = !a.state.ScrollLock
	a.storage.Save(a.state)
	return a.state
}

func (a *App) ToggleStatusBar() *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.HideStatusBar = !a.state.HideStatusBar
	a.storage.Save(a.state)
	return a.state
}

func (a *App) ToggleKeepScreenOn() *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.KeepScreenOn = !a.state.KeepScreenOn
	a.storage.Save(a.state)
	return a.state
}

func (a *App) ToggleShowRotationBtns() *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.ShowRotationBtns = !a.state.ShowRotationBtns
	a.storage.Save(a.state)
	return a.state
}

func (a *App) ToggleMaximize() bool {
	if runtime.WindowIsMaximised(a.ctx) {
		runtime.WindowUnmaximise(a.ctx)
		return false
	}
	runtime.WindowMaximise(a.ctx)
	return true
}

func (a *App) ToggleFullscreen() bool {
	if runtime.WindowIsFullscreen(a.ctx) {
		runtime.WindowUnfullscreen(a.ctx)
		return false
	}
	runtime.WindowFullscreen(a.ctx)
	return true
}

func (a *App) GetWindowState() string {
	if runtime.WindowIsFullscreen(a.ctx) { return "fullscreen" }
	if runtime.WindowIsMaximised(a.ctx)  { return "maximized" }
	return "normal"
}

func (a *App) OpenInBrowser(targetUrl string) {
	runtime.BrowserOpenURL(a.ctx, targetUrl)
}

// normalizeRotation: rotationUrls를 [pinnedUrls 순 → rotationOnly 순]으로 재구성
// 오버레이 메뉴 표시 순서와 툴바 인덱스를 항상 일치시키기 위함
func normalizeRotation(pinnedUrls, rotationUrls []string) []string {
	pinnedSet := map[string]bool{}
	for _, u := range pinnedUrls { pinnedSet[u] = true }
	rotSet := map[string]bool{}
	for _, u := range rotationUrls { rotSet[u] = true }

	result := []string{}
	for _, u := range pinnedUrls {
		if rotSet[u] { result = append(result, u) }
	}
	for _, u := range rotationUrls {
		if !pinnedSet[u] { result = append(result, u) }
	}
	return result
}

func (a *App) IsMacOS() bool       { return platformIsMac() }
func (a *App) ChildGoBack()        { platformGoBack() }
func (a *App) ChildGoForward()     { platformGoForward() }
func (a *App) ChildReload()        { platformReload() }

func extractHostPort(rawUrl string) string {
	u, err := url.Parse(rawUrl)
	if err != nil { return "" }
	return u.Host
}
