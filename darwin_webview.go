//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework WebKit -framework Foundation

#import <Cocoa/Cocoa.h>
#import <WebKit/WebKit.h>

static WKWebView* _jbChildWV = nil;

static const char* BLANK =
    "<html><body style='margin:0;background:#1e1e2e'></body></html>";

static void _doSetup(void);
static void _retrySetup(void) {
    dispatch_after(
        dispatch_time(DISPATCH_TIME_NOW, (int64_t)(0.4 * NSEC_PER_SEC)),
        dispatch_get_main_queue(), ^{ _doSetup(); });
}

static void _doSetup(void) {
    NSWindow* win = [[NSApplication sharedApplication] mainWindow];
    if (!win) { _retrySetup(); return; }

    // child WKWebView 생성
    WKWebViewConfiguration* cfg = [[WKWebViewConfiguration alloc] init];

    NSRect frame = win.contentView.bounds;
    _jbChildWV = [[WKWebView alloc] initWithFrame:frame configuration:cfg];
    _jbChildWV.autoresizingMask = NSViewWidthSizable | NSViewHeightSizable;
    [_jbChildWV loadHTMLString:[NSString stringWithUTF8String:BLANK] baseURL:nil];

    // Wails parent WKWebView 아래에 삽입
    [win.contentView addSubview:_jbChildWV positioned:NSWindowBelow relativeTo:nil];

    // Wails parent WKWebView를 투명하게 (child가 비치도록)
    for (NSView* sub in win.contentView.subviews) {
        if ([sub isKindOfClass:[WKWebView class]] && sub != _jbChildWV) {
            [(NSView*)sub setWantsLayer:YES];
            [(NSView*)sub setValue:@NO forKey:@"opaque"];
            [sub layer].backgroundColor = CGColorGetConstantColor(kCGColorClear);
            break;
        }
    }
}

void jbSetupChild(void) {
    dispatch_async(dispatch_get_main_queue(), ^{ _doSetup(); });
}

void jbNavigate(const char* url) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (!_jbChildWV) return;
        if (!url || strlen(url) == 0) {
            [_jbChildWV loadHTMLString:[NSString stringWithUTF8String:BLANK] baseURL:nil];
            return;
        }
        NSString* s = [NSString stringWithUTF8String:url];
        NSURL* nsurl = [NSURL URLWithString:s];
        if (!nsurl) return;
        [_jbChildWV loadRequest:[NSURLRequest requestWithURL:nsurl]];
    });
}

void jbGoBack(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (_jbChildWV && [_jbChildWV canGoBack]) [_jbChildWV goBack];
    });
}

void jbGoForward(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (_jbChildWV && [_jbChildWV canGoForward]) [_jbChildWV goForward];
    });
}

void jbReload(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (_jbChildWV) [_jbChildWV reload];
    });
}
*/
import "C"
import "unsafe"

func platformSetup()              { C.jbSetupChild() }
func platformNavigate(url string) { cs := C.CString(url); defer C.free(unsafe.Pointer(cs)); C.jbNavigate(cs) }
func platformGoBack()             { C.jbGoBack() }
func platformGoForward()          { C.jbGoForward() }
func platformReload()             { C.jbReload() }
func platformIsMac() bool         { return true }
