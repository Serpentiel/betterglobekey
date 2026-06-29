#import <Foundation/Foundation.h>

// bgkRunOnMain is implemented in Go (see callback.go) and invoked here on the
// main thread to run the pending function identified by handle.
extern void bgkRunOnMain(unsigned long long handle);

// bgkDispatchMain runs the pending function identified by handle on the main
// thread, blocking until it completes. If already on the main thread it runs it
// directly; otherwise it marshals the call via the main dispatch queue.
void bgkDispatchMain(unsigned long long handle) {
	if ([NSThread isMainThread]) {
		bgkRunOnMain(handle);
		return;
	}

	dispatch_sync(dispatch_get_main_queue(), ^{
	  bgkRunOnMain(handle);
	});
}
