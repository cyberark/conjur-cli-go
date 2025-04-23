package version

import (
	"fmt"
)

// Version field is a SemVer that should indicate the baked-in version
// of the CLI
var Version = "unset"

// Tag field denotes the specific build type for the CLI. It may
// be replaced by compile-time variables if needed to provide the git
// commit information in the final binary. See `Static long version tags`
// in the `Building` section of `CONTRIBUTING.md` for more information on
// this variable.
var Tag = "unset"

var buildYear = "2025"

// DisclaimerText field
var disclaimerText = fmt.Sprintf("Copyright (c) %v CyberArk Software Ltd. All rights reserved.\n<www.cyberark.com>", buildYear)

// FullVersionName is the user-visible aggregation of version and tag
// of this codebase
var FullVersionName = fmt.Sprintf("%s-%s\n\n%s", Version, Tag, disclaimerText)
