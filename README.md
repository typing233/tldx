# tldx

A fast domain name availability checker that generates candidate domains from keywords, prefixes, suffixes, and TLDs, then checks availability concurrently via the RDAP protocol.

## Features

- **Combinatorial generation**: Cartesian product of prefixes × keywords × suffixes × TLDs
- **Concurrent RDAP checking**: Parallel availability lookups with configurable concurrency
- **Preset TLD sets**: Quick selection of curated TLD groups (popular, tech, startup, etc.)
- **Multiple input sources**: Keywords from flags, files, or stdin pipes
- **Smart output**: Colored text or JSON, with optional available-only filtering
- **Early stop**: Halt after finding N available domains
- **Rate-limit aware**: Automatic retry with exponential backoff on 429/5xx responses
- **Zero dependencies**: Built entirely with Go standard library (RDAP is just HTTP+JSON)
- **MCP server**: Skeleton Model Context Protocol server for AI tool integration

## Installation

```bash
# From source
git clone <repo-url> && cd tldx
go build -o tldx .

# Or directly
go install tldx@latest
```

## Usage

### Check domain availability

```bash
# Basic keyword check
tldx -k hello,world -t com,ai

# With prefixes and suffixes
tldx -k app -p get,my -s hq,io --preset tech

# Only show available domains
tldx -k cloud,data -t com,dev,io --available-only

# Stop after finding 3 available domains
tldx -k fast,quick,swift -t com,net,io --limit 3

# Read keywords from file
tldx -f keywords.txt --preset popular

# Pipe keywords from stdin
echo -e "alpha\nbeta\ngamma" | tldx -f - -t ai,dev

# JSON array output
tldx -k startup -p get --preset tech --format json-array

# Filter long domains
tldx -k internationalization -p my --max-length 15 -t com
```

### List TLD presets

```bash
tldx presets
```

Output:
```
Available TLD Presets:

  country  us, uk, ca, de, fr, jp
  new      xyz, online, site, store, fun
  popular  com, net, org, io
  short    io, ai, co, me, to
  startup  co, io, ai, app, dev
  tech     dev, io, ai, app, tech
  web      com, net, org, info, biz
```

### Start MCP server

```bash
tldx mcp
```

Starts a JSON-RPC stdio server implementing the Model Context Protocol, exposing a `check_domains` tool.

## Options Reference

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--keywords` | `-k` | | Comma-separated keywords |
| `--file` | `-f` | | File path (use `-` for stdin) |
| `--prefix` | `-p` | | Comma-separated prefixes |
| `--suffix` | `-s` | | Comma-separated suffixes |
| `--tld` | `-t` | `com` | Comma-separated TLDs |
| `--preset` | | | Preset TLD set name |
| `--max-length` | | `0` | Max domain length (0=unlimited) |
| `--limit` | `-n` | `0` | Stop after N available (0=all) |
| `--concurrency` | `-c` | `5` | Concurrent workers |
| `--timeout` | | `10s` | HTTP request timeout |
| `--retries` | | `1` | Retries on transient errors |
| `--format` | | `text` | Output: `text` or `json-array` |
| `--no-color` | | | Disable colored output |
| `--available-only` | | | Only show available domains |

## JSON Output Format

With `--format json-array`, stdout receives a plain JSON array of available domains:

```json
[
  "example.dev",
  "example.ai"
]
```

Statistics are printed to stderr.

## How It Works

1. **Bootstrap**: Fetches the IANA RDAP bootstrap registry to map TLDs to their RDAP servers (with embedded fallback for common TLDs)
2. **Generate**: Produces all combinations of prefix + keyword + suffix + TLD
3. **Check**: Concurrent workers query RDAP servers (`GET /domain/{name}`) — HTTP 404 means available, 200 means registered
4. **Output**: Results streamed in real-time (text) or collected (JSON), with final statistics

## Environment

- `NO_COLOR=1` — Disable colored output (respects [no-color.org](https://no-color.org) convention)

## License

MIT
