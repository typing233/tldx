package cmd

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func Run() int {
	if len(os.Args) < 2 {
		printUsage()
		return 1
	}

	switch os.Args[1] {
	case "check":
		return runCheck(os.Args[2:])
	case "presets":
		return runPresets(os.Args[2:])
	case "mcp":
		return runMCP(os.Args[2:])
	case "--help", "-h", "help":
		printUsage()
		return 0
	case "--version", "-v":
		fmt.Printf("tldx %s\n", version)
		return 0
	default:
		return runCheck(os.Args[1:])
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `tldx - Domain Availability Checker (v%s)

Generates candidate domain names from keywords, prefixes, suffixes, and TLDs,
then checks availability via RDAP protocol.

Usage:
  tldx [command] [options] [keywords...]

Commands:
  check     Check domain availability (default if no command specified)
  presets   List available TLD preset sets
  mcp       Start MCP (Model Context Protocol) server

Options (for check):
  -k, --keywords <list>     Comma-separated keywords
  -f, --file <path>         Read keywords from file (use - for stdin)
  -p, --prefix <list>       Comma-separated prefixes to prepend
  -s, --suffix <list>       Comma-separated suffixes to append
  -t, --tld <list>          Comma-separated TLDs (default: com)
      --preset <name>       Use a preset TLD set (e.g. popular, tech, startup)
      --max-length <n>      Maximum domain name length (0 = no limit)
  -n, --limit <n>           Stop after finding N available domains
  -c, --concurrency <n>     Number of concurrent workers (default: 5)
      --timeout <duration>  HTTP request timeout (default: 10s)
      --retries <n>         Retry count on transient errors (default: 1)
      --format <fmt>        Output format: text or json-array (default: text)
      --no-color            Disable colored output
      --available-only      Only show available domains

Examples:
  tldx -k hello,world -t com,ai
  tldx -k app -p get,my -s hq --preset tech --available-only
  tldx -f keywords.txt --preset popular --limit 5
  echo -e "test\ndemo" | tldx -f - -t io --format json-array
  tldx presets

`, version)
}
