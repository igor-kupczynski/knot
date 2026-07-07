# Agent instructions

This is a personal, agent-maintained markdown knowledge base. You — the resident agent, working in this folder — are its only editor. Outside agents read it through `knot` (grep/read/glob) and push raw notes into `inbox/` with `knot capture`; that is their only write path. Turning those notes into knowledge is your job.

You own this file. Update it when the KB's conventions have intentionally changed — never as a side effect of filing — so it always describes how this KB actually works.

## Layout

- `index.md` — browse layer: every page, one line, grouped by topic.
- `pages/*.md` — entity/topic pages; all synthesized knowledge lives here.
- `inbox/*.md` — unfiled captures; presence here means not yet filed.
- `log.md` — audit trail of filings, newest on top.

## Filing (the main job)

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

## Page model

- One page per thing — a system, a project, a runbook, a pattern. The title is the thing; the first `#` heading matches the filename. No filename prefixes, no filing by type.
- New information about a thing goes on that thing's page. Prefer fewer, fatter pages over many tiny ones; merge pages that turn out to be the same topic, then fix their links and `index.md` lines.
- Frontmatter is exactly this, in this order, nothing else (`sources: []` when there are none):

```
---
updated: 2026-03-02
sources:
  - https://github.com/org/repo/pull/1234
---
```

- Decisions are dated sections, newest on top. Never delete a superseded decision: strike through its heading and body — the old date stays as written — and add `Superseded YYYY-MM-DD: <one-line why>` immediately after:

```
## 2026-03-02 — deploys go through CI only

Manual deploys caused the February outage; CI is the only path now.

## ~~2026-01-15 — deploying from laptops is fine~~

~~Anyone can `make deploy` from a clean checkout.~~

Superseded 2026-03-02: caused the February outage.
```

- Relative markdown links everywhere — `[Shards](<Shards.md>)` page-to-page, `[Shards](<pages/Shards.md>)` from the root; angle brackets around paths with spaces. Never `[[wikilinks]]`.
- Don't hard-wrap prose; let editors soft-wrap.

## index.md

`##` topic headings; under each, one bullet per page, alphabetized — a relative link plus an em-dash clause on what the page answers:

```
## Infrastructure

- [Deploys](<pages/Deploys.md>) — how code ships; current CI-only policy.
```

New page, renamed page, or a drifted one-liner means fixing the line. Nothing else lives here.

## log.md

One entry per filing, under today's `## YYYY-MM-DD` heading at the top of the file (create it above older dates if missing); newest first throughout:

```
## 2026-03-02

- Filed shard-rebalancing capture into [Shards](<pages/Shards.md>): new decision section, struck through the 2026-01 approach.
```

## Editorial judgment

- Synthesis over transcription. A page is the answer to "tell me about X" — a faithful digest with source links, not an archive of originals.
- Cuts beat additions. Default to less; when a page bloats, rewrite it down.
- When captures conflict and neither clearly supersedes, don't guess: record both claims on the page, each with its sources, plus what would settle it. Pick a winner only when the reason is clear — newer, better sourced, or explicitly a decision.
- Sections only where there's content — no empty scaffolds. No new top-level folders as part of filing; if one is genuinely needed, update this contract first to define its purpose.
- Extend an existing page unless the material genuinely has no home; a new page needs a title that names a thing, not a category.
