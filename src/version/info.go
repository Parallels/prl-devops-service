package version

import (
	"fmt"
	"runtime"
	"strings"
)

// Info holds version and build metadata used for display purposes.
type Info struct {
	Version   string
	Channel   Channel
	AppName   string
	Author    string
	BuildDate string
	GitCommit string
	GoVersion string
}

// Current builds an Info from the build-time ldflag variables.
func Current(appName string) Info {
	v := Get()
	return Info{
		Version:   v.String(),
		Channel:   v.Channel,
		AppName:   appName,
		Author:    "Carlos Lapao",
		BuildDate: buildDate,
		GitCommit: buildCommit,
		GoVersion: runtime.Version(),
	}
}

// PrintBanner prints the startup ASCII-art banner.
// Build metadata (author, date, commit) is only shown when isDebug is true.
func (i Info) PrintBanner(isDebug bool) {
	const banner = `
██████╗  █████╗ ██████╗  █████╗ ██╗     ██╗     ███████╗██╗     ███████╗
██╔══██╗██╔══██╗██╔══██╗██╔══██╗██║     ██║     ██╔════╝██║     ██╔════╝
██████╔╝███████║██████╔╝███████║██║     ██║     █████╗  ██║     ███████╗
██╔═══╝ ██╔══██║██╔══██╗██╔══██║██║     ██║     ██╔══╝  ██║     ╚════██║
██║     ██║  ██║██║  ██║██║  ██║███████╗███████╗███████╗███████╗███████║
╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚══════╝╚══════╝╚══════╝╚══════╝

██████╗ ███████╗██╗   ██╗ ██████╗ ██████╗ ███████╗
██╔══██╗██╔════╝██║   ██║██╔═══██╗██╔══██╗██╔════╝
██║  ██║█████╗  ██║   ██║██║   ██║██████╔╝███████╗
██║  ██║██╔══╝  ╚██╗ ██╔╝██║   ██║██╔═══╝ ╚════██║
██████╔╝███████╗ ╚████╔╝ ╚██████╔╝██║     ███████║
╚═════╝ ╚══════╝  ╚═══╝   ╚═════╝ ╚═╝     ╚══════╝

`
	const width = 72

	fmt.Print(banner)
	fmt.Println(centerText(fmt.Sprintf("%s v%s", i.AppName, i.Version), width))
	fmt.Println(strings.Repeat("=", width))

	fmt.Printf("Version:    %s\n", i.Version)
	fmt.Printf("Channel:    %s\n", i.Channel)
	fmt.Printf("Go Version: %s\n", i.GoVersion)
	fmt.Printf("License:    %s\n", "Fair Source (https://fair.io)")
	fmt.Printf("Author:     %s\n", i.Author)

	if isDebug {
		if i.BuildDate != "" {
			fmt.Printf("Build Date: %s\n", i.BuildDate)
		}
		if i.GitCommit != "" && i.GitCommit != "unknown" {
			fmt.Printf("Git Commit: %s\n", i.GitCommit)
		}
		fmt.Printf("Debug Mode: ENABLED\n")
	}

	fmt.Println(strings.Repeat("=", width))
	fmt.Println()
}

// PrintSimpleVersion prints a single-line version string.
func (i Info) PrintSimpleVersion() {
	fmt.Printf("%s version %s (channel: %s)\n", i.AppName, i.Version, i.Channel)
}

// String returns a compact representation suitable for logging and API responses.
func (i Info) String() string {
	return fmt.Sprintf("%s v%s", i.AppName, i.Version)
}

// centerText pads text to sit in the middle of a fixed-width column.
func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	pad := (width - len(text)) / 2
	return strings.Repeat(" ", pad) + text
}
