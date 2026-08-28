//go:build windows

package svc

// daemonize is a no-op stub on Windows to satisfy svc.go's
// Daemonize(proc) symbol. The supervisor daemon is not supported on
// Windows; callers should invoke Server.Run directly. Mirrors the
// legacy daemonize_windows.go in the root package.
func daemonize(proc func()) {
	proc()
}