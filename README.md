# knot

A single-binary CLI that gives any coding agent, on any machine, read + capture
access to a personal markdown knowledge base — pull the context behind
decisions, push new knowledge back, without leaving the session.

The KB is a plain folder of markdown. knot is deletable: if it vanishes, the
knowledge base still works as-is.

## How it works

- **Outside agents** (any repo, any machine) search and read through knot, and
  push raw notes into the KB's inbox with `knot capture`. They never edit pages.
- **The resident agent** — an interactive agent session inside the KB folder —
  periodically files the inbox into entity/topic pages, following the contract
  in the KB's `AGENTS.md` (scaffolded by `knot init`).

This split is the design: capture is instant and mechanical everywhere;
editorial judgment lives in one place.

## Install

    go install github.com/igor-kupczynski/knot@latest

Requires [ripgrep](https://github.com/BurntSushi/ripgrep)
(`brew install ripgrep`) — `knot grep` delegates to it.

## Quick start

    knot init ~/kb                        # first machine: scaffold a knowledge base
    knot init ~/kb --from <git-url>       # other machines: clone an existing KB
    knot prompt --snippet                 # paste into your global agent config
                                          # (CLAUDE.md, Cursor rules, AGENTS.md)

Then, from any agent session:

    knot grep postgres                  # search pages + inbox + log
    knot read deploys                   # print a page by name or path
    knot glob inbox/                    # list KB files
    knot capture --page deploys --source https://example.com/pr/1 "CI-only deploys: ..."
    knot sync                         # commit + pull + push the KB

## Sync

`knot sync` commits any local changes (`sync: <hostname> <timestamp>`), then `git pull --rebase` and `git push` when a remote is configured. Safe to run repeatedly or from a scheduler.

One-time remote setup (your step): create a private repo, then `git remote add origin <url>` in the KB root. After that, `knot sync` handles commit/pull/push. On every other machine, `knot init <path> --from <git-url>` clones the KB with the remote already set, so `knot sync` works immediately.

On rebase conflict, knot aborts the rebase (your local commit is preserved), lists the conflicting paths, and tells you to run `/git-conflict-resolve` in a resident agent session. knot never merges content itself.

`knot init` scaffolds the KB as a git repo (initial commit) and includes resident-agent skills under `.agents/skills/` (`/ingest` for filing, `/git-conflict-resolve` for conflicts).

## Commands

| command | what it does |
|---|---|
| `knot init <path>` | scaffold the KB layout (incl. its `AGENTS.md` contract); `--from <git-url>` clones an existing KB instead |
| `knot grep <pattern>` | search the KB — ripgrep passthrough, KB-relative paths |
| `knot read <name-or-path>` | print a page by name, or any KB file by path |
| `knot glob <pattern>` | list KB files matching a pattern |
| `knot capture` | append a timestamped note to the KB inbox |
| `knot sync` | commit local changes, pull --rebase, and push via git |
| `knot prompt` | agent-facing usage; `--snippet` for the config paragraph |

The KB path comes from `KNOT_KB` or `~/.config/knot/config` (written by
`knot init`).

See [BRIEF.md](BRIEF.md) for the design brief.
