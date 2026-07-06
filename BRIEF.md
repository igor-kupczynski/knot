# knot — brief

A single-binary CLI (Go; `brew install` / `go install`) that gives any coding agent, on any machine, read + capture access to a personal markdown knowledge base. Agents working elsewhere (e.g. a coding session in a repo) can pull the context behind decisions and push new knowledge back — without leaving their session.

knot defines its own KB layout (`knot init` scaffolds it).

## Principles

- **Markdown is the system of record; knot is deletable.** If knot vanishes, the KB is a folder of readable markdown that works as-is. Any index knot builds is a disposable, rebuildable cache.
- **What knot writes, knot formats.** Capture files are mechanically well-formed no matter which agent calls it. Editorial conventions (dated sections, links, log entries) live in one place — the KB's own `AGENTS.md` — not scattered across callers' prompts.
- **Capture ≠ filing.** Outside sessions capture raw material instantly; synthesis into pages happens in the KB, by the resident agent, where the editorial rules live.
- **Fully machine-maintained.** No human-log ritual, no hand-edited zone. The human feeds material in through agents; the KB is the agents' artifact.
- **Capture never fails, never blocks.** Any friction and agents silently stop filing.

## v1 scope

**Read** — lexical search over the whole corpus (pages, audit log, unfiled inbox). Reads go through knot rather than the agent's native tools because the KB lives outside the agent's workspace: a once-approved `knot` command clears the sandbox/permission friction that would otherwise make agents silently stop looking things up — the read-side counterpart of "capture never fails":
- `knot grep <pattern>` — named `grep` because models already know exactly what to do with it (it is literally the search tool's name in Cursor and Claude Code). Output is ripgrep's own format, untouched — `path:line:text`, KB-relative paths — because agents already parse it reflexively. The full rg contract — regex, smart-case, flags, exit codes — comes with it.
- `knot read <name-or-path>` — print a page by name, or any KB file by relative path, so grep hits feed straight into read (the grep→read loop agents already run reflexively).
- `knot glob <pattern>` — discovery that scales: list KB files matching a pattern without ingesting the whole index (`index.md` is curated and partial as the KB grows; glob is mechanical and complete).
- Search is delegated to ripgrep via subprocess — knot is a thin wrapper that resolves the KB path, runs `rg` there, and passes output through. No parsing, no result model of knot's own. `rg` is a hard dependency in v1 (declared in the brew formula; checked with a clear install hint).

**Capture** — append-only, instant, mechanically well-formed:
- `knot capture [--page X] [--source URL]...` — one timestamped file per capture in `inbox/`, immediately searchable; presence in `inbox/` is the unfiled marker, no extra state to drift. Captures are universal — no kinds or types; the resident agent judges what each one is when filing.
- Sources are links (a Slack convo summarized in the capture); `--source` repeats, since one capture often has several (a PR, its ticket, the Slack thread).
- No page-edit primitive from outside. Ever (v1).
- Concurrency-safe by construction: one file per capture, no merges, plays fine with file sync (e.g. iCloud) + git + N parallel sessions.

**File** — the resident agent folds the inbox into pages: synthesis, superseded-decision strikethroughs, index and audit-log updates. In v1 this is an interactive session in the KB folder, e.g. Cursor, Claude Code, which already has connectors such as slack/jira for chasing sources.

**Integrate**
- Two-layer self-onboarding: `knot prompt --snippet` emits the short paragraph for global CLAUDE.md / Cursor rules / AGENTS.md equivalents; the snippet just points agents at `knot prompt`, which emits the full agent-facing usage. The real instructions live in the binary, so the pasted config never drifts.
- KB path via `~/.config/knot` or env var.

## KB layout (owned by knot, scaffolded by `knot init`)

```
kb/
  AGENTS.md       # resident agent's contract: filing workflow, page model, conventions (scaffolded once; the resident agent owns it after)
  index.md        # browse layer, one line per page
  pages/*.md      # entity/topic pages; dated decision sections, minimal frontmatter (updated:, sources:)
  inbox/*.md      # unfiled captures, timestamped, one per capture
  log.md          # append-only audit trail of filings
```

Page model (sketch — the authoritative version lives in the scaffolded `AGENTS.md`): entity-centric pages, decisions as dated sections newest-on-top, superseded ones struck through, relative markdown links, cuts beat additions.

## v1.1

- **Semantic search** — optional backend behind `knot semanticsearch`, never a dependency.
- **Attachments** — `knot capture --attach <file>` for the occasional raw artifact worth keeping (a paper, a transcript), stored under `sources/` and referenced from the capture. Safe to add later: capture is append-only.

## Non-goals

- Indexing agent transcripts or anything outside the KB.
- Databases, servers, embeddings-as-requirement.
- Replacing the resident-agent pattern — knot is the remote door; editorial control stays in one place.
- Human UI, human-log standup rituals, task tracking.
