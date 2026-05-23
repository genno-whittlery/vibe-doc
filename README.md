# vibe-doc — personal web doc system that just works

A small Go binary that serves a directory tree of Markdown files at clean URLs.
README.md is the folder index. Symlinks work. Multi-mount via longest-prefix.
Mermaid + LaTeX render client-side. Search built in.

## Why

VitePress / Docusaurus / MkDocs are excellent at what they do, but they each
ship config-system complexity and plugin churn that's overkill for the
"see my docs at a path" case. vibe-doc is for that case.

## Quickstart

    go install github.com/genno-whittlery/vibe-doc/cmd/vibe-doc@latest
    vibe-doc serve --mount /=./docs

Then open <http://127.0.0.1:4000/>. See `docs/quickstart.md` for the longer story.

## Status

Pre-1.0. Spec at <docs/spec.md>. Issues welcome on
[github.com/genno-whittlery/vibe-doc](https://github.com/genno-whittlery/vibe-doc).

## License

MIT — see [LICENSE](./LICENSE).
