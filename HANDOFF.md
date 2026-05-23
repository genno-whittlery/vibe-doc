# vibe-doc — Session Handoff (2026-05-23)

> **Status:** Tasks 1-3 of 19 complete. Tasks 4-19 remain. Resume in a fresh Claude Code session pointed at this repo. The original implementation plan is at `~/proj/puzzle-platform/docs/superpowers/plans/2026-05-23-vibe-doc.md` and includes a "Plan Revision 2" addendum at the end that supersedes parts of several tasks.

## Why this handoff exists

The implementation plan is 19 tasks across 10 phases, each requiring 3-4 subagent dispatches under the `superpowers:subagent-driven-development` skill (implementer + spec-compliance reviewer + code-quality reviewer + sometimes a fix loop). The original session ran out of context capacity around Task 3. Rather than continue partway and risk a stuck session mid-task, the user opted for a clean break and fresh-session resumption. This file is the operator's manual for picking up where we left off.

## Current state (verified clean)

| Aspect | Value |
|---|---|
| Repo | `~/genno/vibe-doc/` |
| Branch | `main` |
| HEAD | `10c402f` |
| Commits ahead of origin | 3 (need to push) |
| GitHub | `https://github.com/genno-whittlery/vibe-doc` (Task 1 pushed up; Tasks 2-3 are local-only until next push) |
| `go build ./...` | Clean |
| `go test -race ./...` | All packages pass |
| `go vet ./...` | Clean |

### Commits

```
10c402f  feat(logger): mutex-protected rotating file logger with level gating
aee6d34  fix(config): drop bogus // indirect on toml dep + test non-feature ~user/x
e378618  feat(config): TOML config + ParseMountFlag with ~ expansion
7b4baa9  feat: scaffold vibe-doc — single Go binary, /__health, version subcommand
```

### Packages implemented

- `cmd/vibe-doc/main.go` — CLI skeleton (`serve` / `check` / `version` / `help`); `serve` registers `/__health` returning `ok\n`, listens on `127.0.0.1:<--port>` (default 4000). `check` is still a placeholder (Task 16 fills it in).
- `internal/config/` — TOML config + `--mount URL=PATH` flag parser + `~` home-dir expansion. 6 tests passing.
- `internal/logger/` — mutex-protected rotating file logger; INFO/WARN/ERROR levels with level gating; rotates to `<path>.old` on size exceeded. 4 tests passing under `-race`.

### What is NOT yet implemented

- Mount routing (`internal/mount/`) — Task 4
- URL → file resolution (`internal/route/`) — Task 5
- Filesystem walk (`internal/walk/`) — Task 6
- Front-matter parser (`internal/frontmatter/`) — Task 7
- Markdown rendering (`internal/render/`) — Task 8
- HTTP server + embedded assets (`internal/server/`, `assets/`) — Task 9
- Sidebar tree (`internal/sidebar/`) — Task 10
- Shadow detection (`internal/shadow/`) — Task 11
- Search inverted index (`internal/search/`) — Task 12
- Search endpoint + UI — Task 13
- fsnotify + SSE auto-reload — Task 14
- `/sitemap.xml` (`internal/sitemap/`) — Task 15
- `vibe-doc check` subcommand (`internal/check/`) — Task 16
- Migration into puzzle-platform docs — Tasks 17-18
- README polish + v0.1.0 release — Task 19

## How to resume

In a fresh Claude Code session, invoke:

```
/cd ~/genno/vibe-doc
```

Then either:

### Option A — Subagent-driven (recommended for thoroughness)

```
Resume executing the vibe-doc plan from Task 4 onward. Plan at
/Users/mindos/proj/puzzle-platform/docs/superpowers/plans/2026-05-23-vibe-doc.md
— the Plan Revision 2 addendum at the bottom supersedes affected steps.
Current HEAD is 10c402f (Tasks 1-3 done). Use subagent-driven-development.
```

The agent should read `HANDOFF.md` first for state, then dispatch the Task 4 implementer.

### Option B — Inline (faster, less review)

Same prompt but with "use executing-plans" instead of "subagent-driven-development." Trades the per-task spec + quality review for batch execution.

### Option C — Manual (if you want to drive)

The plan's full text is in the markdown — Task 4 onward is mostly mechanical Go + TDD if you've read the spec. You can implement task-by-task yourself, using the plan as a checklist.

## Critical things a resumed session MUST know

1. **Plan Revision 2 at the bottom of the plan supersedes parts of several original tasks.** In particular:
   - Task 9's original `staticSubFS()` is a compile error — use `fs.Sub(assets.FS, "static")` per Revision 2.
   - Task 9's original test file has a broken `mustListen` stub — use the corrected `httptest.NewRecorder` form.
   - Task 9's page handler must execute the `base.html` template (sidebar + TOC + conditional Mermaid/Math script loading). Original task wrote raw `<!DOCTYPE html>` strings; that's wrong.
   - Task 12's `Add()` was hand-rolled; use `AddDoc(IndexedDoc{Headings, Tags})` from Revision 2 for BM25 section boost.
   - Task 13's search-index population was hand-waved; Revision 2 has the concrete `rebuildSearchIndex()` code.
   - Task 14's fsnotify watcher must walk recursively and use 100ms debounce per Revision 2.
   - Tasks 9, 11, 17 expect Server.New() to run `shadow.Scan` at startup and honor `--strict`.

2. **Git identity must stay Genno.** Already configured locally in `~/genno/vibe-doc/`:
   - `user.name = Genno`
   - `user.email = genno@whittlery.io`
   Verify with `git config --local user.name`. Don't use the global Mindos identity here.

3. **SSH alias for GitHub push.** Remote is `git@github-genno:genno-whittlery/vibe-doc.git`. Don't change it. Push with `git push origin main` — there are 3 unpushed commits (Tasks 2, 3, 3-fix).

4. **Test before commit.** Every task in the plan has a `go test -race` step. Honor it. The mutex-protected logger especially needs `-race`.

5. **Discipline: don't pre-empt later tasks.** Each task touches a specific set of files. Tasks 2 and 3 deliberately did NOT modify `cmd/vibe-doc/main.go` — Task 9 wires everything together. Resist the urge to thread the logger into main.go before Task 9.

## Known follow-ups (logged, not blocking)

From the code reviewer's notes during the original session:

**Minor (defer to v0.1.0 cleanup):**
- `main.go:54` — `w.Write` return ignored; use `_, _ = w.Write(...)` for linter cleanliness
- README links to `docs/quickstart.md` + `docs/spec.md` are broken (those files don't exist yet; Task 19 creates them)
- `.gitignore` uses anchored `/vibe-doc` and `/vibe-doc.exe` — correct fix; the plan still shows unanchored (worth back-porting to plan text)
- `logger.write()` reads `l.level` without mutex — safe today (no setter); add atomic if a setter ever lands
- `logger.New()` ignores `f.Stat()` error — practically unreachable on a just-opened file

**v0.1.1 work explicitly deferred from v0.1.0 (already documented in Plan Revision 2's release-notes block):**
- Pagefind benchmark against the hand-rolled BM25
- Pagination on `/__search` (currently hard-capped at 50)
- External HTTP link checking in `vibe-doc check`
- Light theme via `--theme` flag
- Recursive symlink loop multi-hop test
- Mount-root-disappears-mid-run explicit test

## Push state before next session

Before resuming, push Tasks 2-3 to GitHub so the remote is current:

```bash
cd ~/genno/vibe-doc
git push origin main
```

Expected: push 3 commits (e378618, aee6d34, 10c402f) to the existing `main` branch.

## TaskList state (for continuation)

The puzzle-platform TaskList tracker has 19 tasks (IDs 7-25). Tasks 1-3 are marked completed. Task 4 (#9 in tracker, "Mount struct + longest-prefix routing") is the next to claim.

In a fresh session, you can re-create the task list or query the existing one — the IDs are persistent.

## Contact for questions

User: mindos (puzzle-platform owner) / Genno (this repo's persona). Spec, plan, and survey are all committed in `~/proj/puzzle-platform/docs/`.
