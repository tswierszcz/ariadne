---
name: catch-up
description: Session primer for the ariadne repo (CVE-reachability-analysis tool for JVM apps). Use this at the very start of any session working in this repository, before doing anything else, and whenever the user asks what's implemented, what's missing, what the current project state is, or says something like "catch me up", "co jest zrobione", "gdzie stoimy", or jumps straight into a task without re-explaining context. Reading this first means the user never has to re-explain the project from scratch.
---

# Catching up on ariadne

Read these in order. Stop as soon as you have enough context for the task at
hand — you don't need to read every file in full on every session, but know
where each kind of answer lives.

## 1. PLAN.md — the source of truth for "what's actually done"

PLAN.md mixes two kinds of text and it is easy to confuse them:

- **The original plan bullets** (the text immediately under each `### Phase N`
  heading) are the pre-implementation sketch, written before any code existed.
  They describe intent, not reality, and are often stale or aspirational.
- **The `**Status: ...**` note** that follows each phase's bullets is written
  *after* the work happened, and is corrected/extended over time as later
  sessions build on it. **This is ground truth.** When the two disagree,
  believe the Status note.

PLAN.md has 8 phases (0–8), each with its own Status note, and they cover
different ground — don't assume Phase 7 alone answers "what's implemented":

| Phase | Covers |
|---|---|
| 0 | environment/toolchain setup |
| 1 | target-app prep, `-app`/`-deps`/`-deps-list` input normalization |
| 2 | OSV querying, CVE list |
| 3 | bytecode parser, hierarchy, call graph, CHA |
| 4 | CVE→method mapping: manual table, fix-commit diffs, the LLM source |
| 5 | the reachability engine (BFS, statuses, paths) |
| 6 | CLI + report rendering; two real bugs found running it end-to-end |
| 7 | input formats beyond a plain jar (Gradle, WAR/EAR) and the profile
    system (`spring`/`servlet`/`jakarta`/`micronaut`/`quarkus`/`generic`,
    in `internal/profile/profiles/*.yaml`) |
| 8 | optional exploitability check (not built) |

For a specific question, read that phase's Status note. For a broad "what's
done / what's missing" question, skim every phase's Status note plus section
8 ("Acceptance criteria"), which is a consolidated checklist across all of
them. As of the last update, the only pieces deliberately left unimplemented
are RTA and `--max-dep-depth` (Phase 7) and an exploitability check (Phase
8, always optional) — don't treat those as forgotten work needing a status
check.

Each profile has a real, small, hand-written control app under
`scripts/prepare-*-demo.sh` (petclinic, webgoat, petclinic-gradle,
servlet-demo, jakarta-demo, micronaut-demo, quarkus-demo) used to verify it
end-to-end against `bin/ariadne`, not just unit tests — if asked to extend or
debug a profile, these are the right reference point and the right pattern
to follow for a new one.

## 2. README.md — how the finished pipeline behaves today

This is the user-facing reference for current flags, the 8-step pipeline, and
known limitations. Use it to answer "how do I run X" or "what does flag Y do"
— it describes the present, not the history. If it disagrees with PLAN.md,
that's a documentation bug worth flagging (it happened before this session
and got fixed; it can happen again).

## 3. FAQ.md — the *why*, and a parallel step-by-step walkthrough

Read this when the task is about explaining a design decision, or when you
need the pipeline explained a different way than README's. It also has its
own "what's supported, what's still missing" section, kept in sync with
PLAN.md's Phase 7 status.

**Do not touch `FAQ_PL.MD`** (the Polish FAQ) unless the user explicitly asks.
It is deliberately left out of sync with the other three files — this is a
standing decision, not an oversight to "fix" on sight.

## 4. Check the repo's actual state, not just the docs

Docs describe intent; `git` describes what's actually there right now:

```bash
git log --oneline -15
git status
```

Recent commit subjects are usually enough to tell whether docs and code are
still in sync, and `git status` catches anything uncommitted from a prior
session.

## What "done" looks like for a change in this repo

Every substantive change in this project's history ships as: the code change
+ a real verification (not just unit tests — run `bin/ariadne` against a real
small app, often a new `scripts/prepare-*-demo.sh`) + an update to whichever
of README.md/PLAN.md/FAQ.md the change affects, in the same commit. If you
add or change a feature, match that pattern rather than leaving docs for
later — "later" tends not to happen, and this repo has already had to spend
a session re-auditing docs that drifted for exactly that reason.
