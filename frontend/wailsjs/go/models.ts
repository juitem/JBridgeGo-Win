export namespace state {
	
	export class AppState {
	    serverUrl: string;
	    hideStatusBar: boolean;
	    zoomLevel: number;
	    pinnedUrls: string[];
	    recentUrls: string[];
	    setupComplete: boolean;
	    keepScreenOn: boolean;
	    orientation: number;
	    rotationUrls: string[];
	    showRotationBtns: boolean;
	    preloadUrls: string[];
	    urlAliases: Record<string, string>;
	    scrollLock: boolean;
	    gridMode: boolean;
	    manualTrustedHosts: string[];
	    hostZoomLevels: Record<string, number>;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.serverUrl = source["serverUrl"];
	        this.hideStatusBar = source["hideStatusBar"];
	        this.zoomLevel = source["zoomLevel"];
	        this.pinnedUrls = source["pinnedUrls"];
	        this.recentUrls = source["recentUrls"];
	        this.setupComplete = source["setupComplete"];
	        this.keepScreenOn = source["keepScreenOn"];
	        this.orientation = source["orientation"];
	        this.rotationUrls = source["rotationUrls"];
	        this.showRotationBtns = source["showRotationBtns"];
	        this.preloadUrls = source["preloadUrls"];
	        this.urlAliases = source["urlAliases"];
	        this.scrollLock = source["scrollLock"];
	        this.gridMode = source["gridMode"];
	        this.manualTrustedHosts = source["manualTrustedHosts"];
	        this.hostZoomLevels = source["hostZoomLevels"];
	    }
	}

}

