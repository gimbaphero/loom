# Embedded Files

This directory contains files that are embedded into the Loom binary at compile time using Go's `embed` directive.

## Files

### weaver.yaml (GENERATED - DO NOT EDIT)

The weaver agent configuration file. This file is **automatically generated** from `weaver.yaml.tmpl` during the build process.

**Source**: `embedded/weaver.yaml.tmpl`
**Generator**: `cmd/generate-weaver/main.go`
**Build Step**: `just generate-weaver` (runs automatically before build)

The generator injects current CLI help documentation into the weaver agent's system prompt, ensuring the weaver always has accurate, up-to-date command reference when instructing users.

**DO NOT edit `weaver.yaml` directly** - your changes will be overwritten on the next build. Instead:
1. Edit `weaver.yaml.tmpl` for template changes
2. Modify `cmd/generate-weaver/main.go` to change CLI help extraction
3. Run `just generate-weaver` to regenerate

### weaver.yaml.tmpl

Template file for the weaver agent configuration. Contains a `{{.CLIHelp}}` placeholder that gets replaced with current CLI help output during build.

To update the weaver prompt:
```bash
# 1. Edit the template
vim embedded/weaver.yaml.tmpl

# 2. Regenerate weaver.yaml
just generate-weaver

# 3. Build
just build
```

## How It Works

1. **Build time**: `just build` → `just generate-weaver` → runs `cmd/generate-weaver`
2. **Generation**:
   - Reads `embedded/weaver.yaml.tmpl`
   - Executes `looms --help`, `looms workflow --help`, etc. (using binary if available, otherwise `go run`)
   - Replaces `{{.CLIHelp}}` with formatted CLI help text
   - Writes `embedded/weaver.yaml`
3. **Compilation**: Go's `embed` directive includes `weaver.yaml` in binary
4. **Runtime**: `embedded.GetWeaver()` returns the embedded content

## agents.go

The Go package that provides access to embedded files:
- `GetWeaver()` - Returns weaver.yaml content
- `GetStartHere()` - Returns base ROM (delegates to pkg/agent/rom_loader.go)

## Why Generate weaver.yaml?

The weaver meta-agent creates other agents and workflows. It needs to know:
- How to tell users to validate workflows (`looms workflow validate`)
- How to tell users to run workflows (`looms workflow run`)
- Exact command syntax and flags

By auto-generating from live CLI help, the weaver:
- ✅ Never gives outdated commands
- ✅ Always knows current flags and options
- ✅ Stays in sync with code changes
- ✅ Reduces hallucination of incorrect commands

## Troubleshooting

**Problem**: `embedded/agents.go:15:12: pattern weaver.yaml: no matching files found`

**Solution**: weaver.yaml doesn't exist yet. This is expected on first build from clean state.
```bash
# Option 1: Build looms first (bootstraps the generator)
just build-server

# Option 2: Create minimal weaver.yaml manually
cp embedded/weaver.yaml.tmpl embedded/weaver.yaml
# Edit to remove {{.CLIHelp}} template placeholder
```

After first successful build, subsequent builds will regenerate automatically.

**Problem**: Changes to `weaver.yaml.tmpl` not appearing in weaver

**Solution**: Run `just generate-weaver` explicitly, then rebuild:
```bash
just generate-weaver
just build
```

## Version Control

- ✅ **Tracked**: `weaver.yaml.tmpl`, `agents.go`, `README.md`
- ❌ **Not Tracked**: `weaver.yaml` (generated file, in .gitignore)

The generated `weaver.yaml` is excluded from git because it's build-time generated and contains system-specific paths.
