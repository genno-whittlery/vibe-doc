# vibe-doc — personal web doc system that just works

A small Go binary that serves a directory tree of Markdown files at clean URLs.
`README.md` is the folder index. Symlinks work. Multi-mount via longest-prefix
match. Mermaid + LaTeX render client-side. Search built in. Single static
binary, no Node, no build step.

## Why

VitePress / Docusaurus / MkDocs are excellent at what they do, but they each
ship config-system complexity and plugin churn that's overkill for the
"see my docs at a path" case. vibe-doc is for that case — and is the tool
that now serves the puzzle-platform docs site after the VitePress cutover.

## Quickstart

```
go install github.com/genno-whittlery/vibe-doc/cmd/vibe-doc@latest
vibe-doc serve --mount /=./docs
```

Then open <http://127.0.0.1:4000/>.

Or with a config file (`vibe-doc.toml` in CWD by default):

```toml
port = 4000
log  = "/tmp/vibe-doc.log"

[[mounts]]
url  = "/"
root = "/path/to/docs"

[[mounts]]
url  = "/engine"
root = "/path/to/engine/docs"

# Optional override; defaults to a sensible skip list including
# node_modules, .git, .vitepress, .docusaurus, dist, .next, .cache.
# exclude = ["node_modules", ".git", "build"]
```

```
vibe-doc serve                # uses vibe-doc.toml in CWD
vibe-doc check                # offline link checker; exits 1 on broken links
vibe-doc version
```

## What's in 0.1.0

- **Routing** — longest-prefix mount match; `README.md` is the folder index;
  bare-folder URLs 301 to the trailing-slash form (spec §4 table).
- **Renderer** — goldmark + GFM (tables, strikethrough, task lists, footnotes)
  + auto heading IDs + a table of contents extracted per page.
- **Mermaid + math** — fenced ```mermaid blocks and `$inline$` / `$$display$$`
  math hydrate client-side via the embedded Mermaid + KaTeX bundles.
- **Sidebar** — auto-generated from the file tree under each mount; group
  titles from each directory's `README.md`; leaf titles from each file's
  first `#` heading or TOML `+++ title = "..." +++` frontmatter.
- **Search** — BM25 inverted index with H1×3 / H2×2 / H3×1.5 / tags×2 / body×1
  section boost, 50ms hard timeout, did-you-mean fallback on 404.
- **Shadow detection** — surfaces conflicts at boot and on `/__shadow`:
  `README.md` vs `index.md` in the same dir; mount overlaps; same-root mounts.
- **Live reload** — fsnotify recursive watch with 100ms debounce; SSE stream
  at `/__events` triggers browser reload on any `.md` change.
- **Symlink-follows by default** — no `fs.allow` dance; directory symlinks
  walked once via canonical-path visited set, cycles detected and skipped.
- **`vibe-doc check`** — walks every mount, scans `[text](target)` links,
  reports absolute targets against mounts and relative targets against disk.

## What's not in 0.1.0

- No build step, no static export. vibe-doc is a live server.
- No multi-language support, no versioned docs, no remote-content includes.
- No Vue / React components inside markdown — pages are plain markdown plus
  the goldmark extensions listed above.
- No anchor verification in `vibe-doc check` (`[link](#section)` accepted
  without checking the anchor exists). Planned for 0.2.

## Spec + design

The spec lives at
[`docs/superpowers/specs/2026-05-23-vibe-doc-design.md`](https://github.com/genno-whittlery/vibe-doc/tree/main)
in the parent puzzle-platform repo (where it was authored). The 19-task
implementation plan, the executing-plans skill that drove the work, and the
HANDOFF.md from the build run are all in the vibe-doc repo's git history.

## Issues

[github.com/genno-whittlery/vibe-doc](https://github.com/genno-whittlery/vibe-doc)
— bug reports + feature requests welcome.

## License

MIT — see [LICENSE](./LICENSE).
