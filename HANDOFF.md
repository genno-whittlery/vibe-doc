# vibe-doc — Session Handoff (2026-05-23, late afternoon)

> **Status:** Tasks 1-16 of 19 complete. Tasks 17-19 (smoke test, cutover, GitHub release) remain — these involve running the binary against a real doc tree, killing the VitePress server, and tagging a public v0.1.0 release. They need a human in the loop.

## Why this handoff exists

The first 16 tasks built the binary from nothing to a feature-complete doc server: multi-mount HTTP routing, longest-prefix matching, embedded assets, goldmark + Mermaid + math rendering, TOC extraction, BM25 search with 50ms timeout cap, sidebar tree builder, shadow detection, fsnotify recursive watch with 100ms debounce, SSE live-reload, sitemap, did-you-mean 404, and an offline link checker subcommand.

The remaining 3 tasks (17, 18, 19) are migration + release work:

- **Task 17** runs `vibe-doc` on `:4001` against `puzzle-platform/docs` alongside the still-running VitePress on `:4000`. The point is a manual sanity check — does the sidebar look right, do the engine-symlink pages resolve, do code blocks render, does search behave. Headless tests cannot answer those questions; a human has to look at a browser.
- **Task 18** is the cutover: when Task 17 passes, stop VitePress, delete `docs/.vitepress/` + `docs/package.json`, swap `npm run docs:dev` to point at `vibe-doc`, restart `:4000`. This is irreversible work in a different repo (`puzzle-platform`) and should be sequenced when the user has bandwidth to verify.
- **Task 19** is the public release: README polish, real upstream Mermaid/KaTeX bundles in `assets/static/`, version bump to `0.1.0` in `cmd/vibe-doc/main.go`, `git tag v0.1.0`, `gh release create`. Branding moment — needs the user's voice.

## Current state (verified clean)

| Aspect | Value |
|---|---|
| Repo | `~/genno/vibe-doc/` |
| Branch | `main` |
| HEAD | `5c6c44b` |
| Push state | up to date with `origin/main` |
| GitHub | `https://github.com/genno-whittlery/vibe-doc` |
| `go build ./...` | Clean |
| `go test -race ./...` | All 14 packages green |
| `go vet ./...` | Clean |
| Smoke (`./vibe-doc check`) | Reports broken links, exit 1 on failures |

## Packages implemented

| Package | Lines | Role |
|---|---|---|
| `cmd/vibe-doc` | ~150 | CLI: serve / check / version / help |
| `internal/config` | ~100 | TOML + flag parsing, `~` expansion |
| `internal/logger` | ~110 | Mutex-protected rotating file logger |
| `internal/mount` | ~80 | Longest-prefix URL routing |
| `internal/route` | ~100 | URL → file resolution (§4 table) |
| `internal/walk` | ~50 | Symlink-following FS walk with cycle detection |
| `internal/frontmatter` | ~80 | TOML `+++` parser; YAML `---` is fatal |
| `internal/render` | ~150 | goldmark + Mermaid + math + TOC |
| `internal/server` | ~600 | HTTP server, handlers, SSE, fsnotify watcher |
| `internal/sidebar` | ~140 | Auto-generated sidebar tree from FS |
| `internal/shadow` | ~120 | README/index, mount-overlap, same-root detection |
| `internal/search` | ~290 | BM25 inverted index + section boost + 50ms timeout |
| `internal/sitemap` | ~60 | `/sitemap.xml` generator |
| `internal/check` | ~110 | Offline link verifier |
| `assets/` | — | Embedded CSS + JS + HTML templates |

**Total**: ~2,200 LOC plus tests, single static binary.

## What works end-to-end (tested)

- `vibe-doc serve --mount /=./docs --port 4000` brings up the server.
- `/__health` returns `ok`.
- `/` and `/<path>/` render markdown with sidebar + TOC + Mermaid + KaTeX placeholders.
- `/static/*` under a mount root serves doc-tree assets.
- `/<bare-folder>` 301-redirects to trailing-slash form.
- `/sitemap.xml` returns valid XML urlset across all mounts.
- `/__search?q=<term>` returns BM25-ranked JSON with 50ms timeout.
- `/__shadow` returns JSON of conflict detector output.
- `/__events` is an EventSource stream; fsnotify changes broadcast `data: reload`.
- 404 includes did-you-mean suggestions from the search index.
- `vibe-doc check` reports broken internal links with line numbers, exit 1 on issues.

## What's NOT done

- **Task 17 (side-by-side smoke):** create `~/proj/puzzle-platform/vibe-doc.toml` with `port = 4001` + mount entries for `docs/` and `engine/`; run the binary alongside VitePress and click through pages. See plan §3784.
- **Task 18 (cutover):** when Task 17 passes, stop VitePress, swap `:4000`, delete `docs/.vitepress/`, update `package.json` scripts. See plan §3790+ for the 8-step migration plan.
- **Task 19 (release):** drop real upstream Mermaid + KaTeX bundles into `assets/static/` (currently placeholder text); polish README; bump version constant in `cmd/vibe-doc/main.go` from `0.0.1-dev` to `0.1.0`; tag + GitHub release. See plan §3900+.

## How to resume

### Option A — Manual Task 17 (recommended next step)

```bash
cd ~/genno/vibe-doc
go build -o /usr/local/bin/vibe-doc ./cmd/vibe-doc  # or wherever your PATH expects

cd ~/proj/puzzle-platform
cat > vibe-doc.toml <<'EOF'
port = 4001
log = "/tmp/vibe-doc.log"
log_max_bytes = 1048576
log_level = "info"

[[mounts]]
url = "/"
root = "/Users/mindos/proj/puzzle-platform/docs"

[[mounts]]
url = "/engine"
root = "/Users/mindos/proj/engine/docs"
EOF

vibe-doc serve &
open https://riddlery.tail1d7c51.ts.net:4001/   # or http://127.0.0.1:4001/
```

Click around. Things to look for:
- Sidebar tree under each mount section.
- Engine symlink target resolves (vibe-doc resolves symlinks; no `fs.allow` dance needed).
- Code blocks, Mermaid blocks, math blocks render or at least don't error.
- `/__shadow` shows any real conflicts in your tree.
- Search finds known terms.

### Option B — Continue with another Claude session

Point a fresh session at this repo, ask it to read `HANDOFF.md`, and tell it to either resume Task 17 (`Task 17 onward, use executing-plans`) or jump straight to Task 19 if you've already done 17/18 manually.

## Critical things a resumed session MUST know

1. **Git identity stays Genno.** `git config --local user.name` should print `Genno`; `user.email` should be `genno@whittlery.io`. Don't use Mindos here.
2. **SSH remote alias is `github-genno`.** Don't change it.
3. **Templates load by filename, not by `New(...)` name.** Use `tpl.ExecuteTemplate("base.html", data)`, not `"base"`. This was a real bug fixed in Task 9.
4. **`mul` / `sub` template funcs accept `any` and coerce to float64.** `TOCEntry.Level` is `int`; pure-float funcs error.
5. **Server has `mu sync.RWMutex`.** `s.sidebar` and `s.searchIdx` reads/writes go through it. Task 14's fsnotify goroutine relies on this for safety.
6. **Test discipline:** every package has `go test -race` as the gate. Honour it.

## Known follow-ups (logged, not blocking)

- **Sidebar symlink cycles.** `buildChildren` uses `os.Stat` (follows symlinks) without inode-visited tracking. A malformed doc tree with a symlink loop in a subdirectory could stack-overflow. The standalone `walk` package handles this correctly; sidebar could be migrated to use it. Not a real-world risk in friendly trees.
- **`os.Stat` errors silently dropped** in `internal/walk/walk.go`. A permission-denied subtree looks identical to an empty subtree. Worth a `errors.Is(err, fs.ErrPermission)` branch.
- **Search index rebuild scans every mount on every fsnotify event.** Per-file invalidation would scale better; current behavior is fine for hundreds of files but would chug at thousands.
- **Mermaid + KaTeX placeholders.** `assets/static/{mermaid,katex}.min.js` are one-line comments. Task 19 must replace with real upstream bundles before the v0.1.0 release.
- **Anchor verification deferred** in `vibe-doc check` — `[link](#foo)` accepted without checking if `#foo` exists in the target doc.

## Commit history (Tasks 1-16)

```
5c6c44b feat(check): offline link verifier subcommand
486017a feat(sitemap): /sitemap.xml endpoint walking every mount
b6d5d18 feat(sse): live-reload SSE + recursive fsnotify watcher with 100ms debounce
a07bacb feat(search): /__search endpoint + 50ms timeout + index populator + search.js UI
c7d5b52 feat(search): BM25 inverted index, section boost, 50ms timeout, swap-path contract
f8071bb feat(shadow): README/index, mount-overlap, same-root detection per §7
fefabe1 feat(sidebar): walk filesystem, README group titles, front-matter ordering
80ea538 fix(server): add Server.mu (race readiness for Task 14), X-Vibe-Mount sentinel, redirect test
40fa5a2 feat(server): HTTP routing + page render + embedded static + base template
c4d05cc fix(render): tighten inline-math regex, robust mermaid close, full TOC text
c35bdc5 feat(render): goldmark + Mermaid + math preprocessing + TOC extraction
7281ea0 perf(frontmatter): precompute close needle; add trailing-ws fence test
fc074f5 feat(frontmatter): TOML +++ parser; YAML --- is a hard error
00ff53e feat(walk): symlink-following filesystem walk with cycle detection
ce86490 refactor(route): drop dead traversal guard, document URL-decode + symlink contracts
c13a7f3 feat(route): URL → file resolution matching spec §4 table
880ab74 test(mount): cover non-boundary prefix (/engineroom != /engine mount)
e2f7174 feat(mount): longest-prefix-match routing with sorted Set
10c402f feat(logger): mutex-protected rotating file logger with level gating
aee6d34 fix(config): drop bogus // indirect on toml dep + test non-feature ~user/x
e378618 feat(config): TOML config + ParseMountFlag with ~ expansion
7b4baa9 feat: scaffold vibe-doc — single Go binary, /__health, version subcommand
```

All pushed to `git@github-genno:genno-whittlery/vibe-doc.git`.

## Contact for questions

User: Mindos (puzzle-platform owner) / Genno (this repo's persona). Spec, plan + Revision 2, survey are all in `~/proj/puzzle-platform/docs/{specs,plans,research}/`.
