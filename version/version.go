package version

import (
	"fmt"
	"io"
	"runtime"
)

var (
	Version  = "dev"
	CommitID = "unknown"
)

func Print(writer io.Writer) {
	_, _ = fmt.Fprintf(writer, "Version:\t\t%s\n", Version)
	_, _ = fmt.Fprintf(writer, "Git commit:\t\t%s\n", CommitID)
	_, _ = fmt.Fprintf(writer, "Go version:\t\t%s\n", runtime.Version())
	_, _ = fmt.Fprintf(writer, "OS/Arch:\t\t%s/%s\n", runtime.GOOS, runtime.GOARCH)
}
