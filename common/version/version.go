package version

import "fmt"

// Default build-time variable for library-import.
// This file is overridden on build with build-time informations.
var (
	version   = ""
	BuildTime = ""
	CommitID  = ""
)

func Version() {
	fmt.Printf("%s-%s %s\n", version, CommitID, BuildTime)
}
