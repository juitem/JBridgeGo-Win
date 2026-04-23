package main

import (
	"context"
	"jbridgego-win/internal/state"
	"net/url"
	"sync"
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
}

func (a *App) GetState() *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state
}

func (a *App) SwitchToUrl(targetUrl string) *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.ServerURL = targetUrl
	a.state.SetupComplete = true
	newRecent := []string{targetUrl}
	for _, u := range a.state.RecentUrls {
		if u != targetUrl { newRecent = append(newRecent, u) }
	}
	if len(newRecent) > 10 { newRecent = newRecent[:10] }
	a.state.RecentUrls = newRecent
	a.storage.Save(a.state)
	return a.state
}

func (a *App) TogglePin(targetUrl string) *state.AppState {
	a.mu.Lock()
	defer a.mu.Unlock()
	found := false
	var newPins []string
	for _, u := range a.state.PinnedUrls {
		if u == targetUrl { found = true; continue }
		newPins = append(newPins, u)
	}
	if !found { newPins = append(newPins, targetUrl) }
	a.state.PinnedUrls = newPins
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
	if hostPort != "" {
		a.state.HostZoomLevels[hostPort] = newZoom
	} else {
		a.state.ZoomLevel = newZoom
	}
	a.storage.Save(a.state)
	return a.state
}

func (a *App) SetAlias(targetUrl string, alias string) *state.AppState {
    a.mu.Lock()
    defer a.mu.Unlock()
    if alias == "" {
        delete(a.state.UrlAliases, targetUrl)
    } else {
        a.state.UrlAliases[targetUrl] = alias
    }
    a.storage.Save(a.state)
    return a.state
}

func (a *App) ToggleRotation(targetUrl string) *state.AppState {
    a.mu.Lock()
    defer a.mu.Unlock()
    found := false
    var newList []string
    for _, u := range a.state.RotationUrls {
        if u == targetUrl { found = true; continue }
        newList = append(newList, u)
    }
    if !found { newList = append(newList, targetUrl) }
    a.state.RotationUrls = newList
    a.storage.Save(a.state)
    return a.state
}

func extractHostPort(rawUrl string) string {
	u, err := url.Parse(rawUrl)
	if err != nil { return "" }
	return u.Host
}
