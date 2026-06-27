// Package mainthread runs functions on the process's main thread, which owns the
// AppKit run loop. Some macOS APIs (notably Text Input Sources) abort when called
// from a thread without a running CFRunLoop, so off-main callers must funnel
// through here.
package mainthread

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation

#import <Foundation/Foundation.h>

void bgkRunOnMain(unsigned long long handle);

static void bgkDispatchMain(unsigned long long handle) {
	if ([NSThread isMainThread]) {
		bgkRunOnMain(handle);
		return;
	}

	dispatch_sync(dispatch_get_main_queue(), ^{
		bgkRunOnMain(handle);
	});
}
*/
import "C"

import "sync"

var (
	mu      sync.Mutex
	lastID  uint64
	pending = make(map[uint64]func())
)

// Run executes fn on the main thread, blocking until fn returns. If called from
// the main thread it runs fn directly.
func Run(fn func()) {
	mu.Lock()
	lastID++
	id := lastID
	pending[id] = fn
	mu.Unlock()

	C.bgkDispatchMain(C.ulonglong(id))

	mu.Lock()
	delete(pending, id)
	mu.Unlock()
}
