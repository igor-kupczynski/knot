---
name: ingest
description: File inbox captures into KB pages — synthesis, index and log updates, capture cleanup. Use when processing the inbox or filing captures.
---

# Ingest

Your worklist is `inbox/`. Empty inbox = nothing to file: don't invent cleanup; broader maintenance only when explicitly asked. For each capture:

1. Read it. Frontmatter: `captured:` (RFC3339 timestamp), `page:` (optional hint — a hint, not an order; it may name a page that doesn't exist yet), `sources:` (list of links).
2. Decide which pages the material touches — one, several, or a new one. Create a page only when none fits; if the capture spans unrelated topics, split it across the right pages.
3. Synthesize into each page — rewrite, don't paste. Keep commands, links, and evidence that carry weight. One coherent edit per topic, not one page per capture.
4. If it records a decision, add a dated section (newest on top) and strike through what it supersedes.
5. Merge the capture's sources into each touched page's `sources:`; bump `updated:`.
6. Update the touched pages' lines in `index.md` (add one if a page is new).
7. Add a log entry naming every page you touched.
8. Delete the capture file. Deletion is the "filed" marker — nothing else tracks state.

File related captures together: several notes on one thing become one coherent edit, not N appends. A capture can also be junk or a duplicate — deleting it with a one-line log entry is a valid filing.

For the page model, `index.md` format, and `log.md` format, see [AGENTS.md](<../../../AGENTS.md>).

9. Run `knot sync`. If it reports conflicts, run `/git-conflict-resolve` and sync again.
