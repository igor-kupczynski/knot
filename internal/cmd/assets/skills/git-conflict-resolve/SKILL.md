---
name: git-conflict-resolve
description: Resolve git merge/rebase conflicts in this KB after a failed knot sync. Use when knot sync exits with conflicting paths.
---

# Git conflict resolve

A failed `knot sync` means local and remote histories diverged. knot aborted the rebase, so the working tree is clean and your local commit is intact — re-surface the conflict first, then resolve it.

## Surface the conflict

```bash
git pull --rebase
```

This stops on the same conflict knot reported. `git status` lists the conflicted files; each contains conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`).

## Per-file strategies

**`log.md`** — append-only audit trail. Keep both sides: union all entries, preserve newest-first ordering (date headings and bullets within each day). Do not drop either side's filings.

**`index.md`** — derived from `pages/`. Do not merge conflict markers as text. Regenerate the affected lines from what is actually in `pages/` (correct one-liner per page under the right `##` topic heading).

**`pages/*.md`** — synthesize both sides into one coherent page per the page model in [AGENTS.md](<../../../AGENTS.md>). Never drop a side's content silently. Frontmatter: `updated:` takes the newer date; `sources:` is the union of both lists.

**`inbox/*.md`** — should not conflict (unique filenames per capture). If two versions of one capture somehow conflict, keep both as separate files rather than merging bodies.

## Finish

```bash
git add <resolved-files>
git rebase --continue
knot sync
```

If the rebase stops on further conflicts, repeat the strategies and `git rebase --continue` until it completes; the final `knot sync` pushes the merged result.
