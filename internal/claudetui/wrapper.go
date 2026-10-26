package claudetui

import _ "embed"

//go:embed wrapper/dist/claude-bridge.js
var wrapperBundle []byte

func Bundle() []byte {
	return wrapperBundle
}
