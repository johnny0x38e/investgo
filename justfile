# Local development, check, and packaging entrypoints for InvestGo.
set minimum-version := "1.56.0"
set default-list
set shell := ["sh", "-eu", "-c"]
set windows-shell := ["powershell.exe", "-NoLogo", "-NoProfile", "-Command"]

powershell := if os() == "windows" { "powershell.exe" } else { "pwsh" }
prepare-icon := if os() == "windows" { "New-Item -ItemType Directory -Force build | Out-Null; Copy-Item scripts/icon/appicon.png build/appicon.png -Force -ErrorAction Stop" } else { "mkdir -p build && cp scripts/icon/appicon.png build/appicon.png" }
render-icon := if os() == "windows" { "bash ./scripts/icon/process/render-app-icon.sh" } else { "./scripts/icon/process/render-app-icon.sh" }

# Install frontend dependencies.
[group("frontend")]
install:
    pnpm install

# Start the Vite frontend development server.
[group("frontend")]
dev:
    pnpm dev

# Build frontend assets into frontend/dist.
[group("frontend")]
build-frontend:
    pnpm build

# Typecheck the Vue/TypeScript frontend.
[group("check")]
typecheck:
    pnpm typecheck

# Format frontend sources with Biome.
[group("check")]
fmt:
    pnpm format

# Run Biome check without writing.
[group("check")]
lint:
    pnpm lint

# Apply Biome safe fixes, including import sorting.
[group("check")]
lint-fix:
    pnpm lint:fix

# Prepare embeds required by `go test` and desktop builds.
[group("check")]
prepare-embeds:
    {{prepare-icon}}
    pnpm build

# Run Go tests after preparing embeds (matches CI).
[group("check")]
test: prepare-embeds
    go test ./...

# Run local verification: typecheck, Biome check, and Go tests.
[group("check")]
check: typecheck lint test

# Render build/appicon.png from the icon pipeline.
[group("build")]
icon:
    {{render-icon}}

# Build the macOS Apple Silicon binary. Pass --dev for DevTools.
[group("build")]
[unix]
build-darwin-aarch64 *args:
    ./scripts/build-darwin-aarch64.sh {{ args }}

# Build the macOS Intel binary. Pass --dev for DevTools.
[group("build")]
[unix]
build-darwin-x86_64 *args:
    ./scripts/build-darwin-x86_64.sh {{ args }}

# Package an Apple Silicon DMG. Example: VERSION=1.0.0 just package-darwin-aarch64
# Pass --dev for DevTools. VERSION defaults to the packaging script's own default.
[group("package")]
[unix]
package-darwin-aarch64 *args:
    ./scripts/package-darwin-aarch64.sh {{ args }}

# Package an Intel DMG. Example: VERSION=1.0.0 just package-darwin-x86_64
# Pass --dev for DevTools. VERSION defaults to the packaging script's own default.
[group("package")]
[unix]
package-darwin-x86_64 *args:
    ./scripts/package-darwin-x86_64.sh {{ args }}

# Build the Windows amd64 binary. Pass -Dev or other PowerShell script parameters.
[group("build")]
build-windows-amd64 *args:
    {{powershell}} -NoLogo -NoProfile -ExecutionPolicy Bypass -File ./scripts/build-windows-amd64.ps1 {{args}}
