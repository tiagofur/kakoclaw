package bridge

import _ "embed"

//go:embed ts/bundle.js
var EmbeddedBundleJS []byte

//go:embed ts/opencode-bundle.js
var EmbeddedOpenCodeJS []byte
