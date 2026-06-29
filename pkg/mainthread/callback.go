package mainthread

import "C"

// bgkRunOnMain is invoked from C on the main thread to run the pending function
// identified by handle.
//
//export bgkRunOnMain
func bgkRunOnMain(handle C.ulonglong) {
	mu.Lock()
	fn := pending[uint64(handle)]
	mu.Unlock()

	if fn != nil {
		fn()
	}
}
