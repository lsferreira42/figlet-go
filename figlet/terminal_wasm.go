//go:build js && wasm

package figlet

// GetColumns returns a default width for WASM builds.
// Since there's no terminal in the browser, we return 0 to use the default width.
func GetColumns() int {
	return 0
}
