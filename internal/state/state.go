package state

type AppState struct {
	ServerURL          string            `json:"serverUrl"`
	HideStatusBar      bool              `json:"hideStatusBar"`
	ZoomLevel          int               `json:"zoomLevel"`
	PinnedUrls         []string          `json:"pinnedUrls"`
	RecentUrls         []string          `json:"recentUrls"`
	SetupComplete      bool              `json:"setupComplete"`
	KeepScreenOn       bool              `json:"keepScreenOn"`
	Orientation        int               `json:"orientation"` // 0=자동, 1=세로, 2=가로 (데스크탑에선 창 모드 대응)
	RotationUrls       []string          `json:"rotationUrls"`
	ShowRotationBtns   bool              `json:"showRotationBtns"`
	PreloadUrls        []string          `json:"preloadUrls"`
	UrlAliases         map[string]string `json:"urlAliases"`
	ScrollLock         bool              `json:"scrollLock"`
	GridMode           bool              `json:"gridMode"`
	ManualTrustedHosts []string          `json:"manualTrustedHosts"`
	HostZoomLevels     map[string]int    `json:"hostZoomLevels"`
}

func NewDefaultState() *AppState {
	return &AppState{
		ServerURL:        "http://your-server:7800",
		HideStatusBar:    true,
		ZoomLevel:        100,
		PinnedUrls:       []string{},
		RecentUrls:       []string{},
		RotationUrls:     []string{},
		PreloadUrls:      []string{},
		UrlAliases:       make(map[string]string),
		HostZoomLevels:   make(map[string]int),
		ShowRotationBtns: true,
	}
}
