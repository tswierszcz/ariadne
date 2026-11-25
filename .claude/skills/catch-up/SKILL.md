---
name: catch-up
description: Session primer for the ariadne repo (CVE-reachability-analysis tool for JVM apps). Use this at the very start of any session working in this repository, before doing anything else, and whenever the user asks what's implemented, what's missing, what the current project state is, or says something like "catch me up", "co jest zrobione", "gdzie stoimy", or jumps straight into a task without re-explaining context. Reading this first means the user never has to re-explain the project from scratch.
---

# Catching up on ariadne

Read these in order. Stop as soon as you have enough context for the task at
hand. You don't need to read every file in full on every session, but know
where each kind of answer lives. (Documentation layout since 2026-11-25: the
old `PLAN.md`, `FAQ.md`, `FAQ_PL.MD` and the long README are archived in
`archive/2026-11-25/docs/`.)

## 1. `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`: what happened last

The log of work, newest entry first, with the overview table of every area
and its state. The first few entries tell you what the last sessions did and
what the next step is. `RECOMMENDATION/README.MD` lists the plans still open
(today: the research interface RI*, the research harness RH-A*) and which
areas are done and archived.

Areas done or withdrawn are archived under `archive/<date>/<area>/`, each
with a `COMPLETION.md` (or `WITHDRAWN.md`) saying what shipped, how it was
verified, and what would reopen it. Believe those over older plan text.

## 2. `docs/SUPPORT.MD`: what is supported now

The capability matrix: status, scope and conditions, limits and the evidence
(tests, fixtures, measurements) for each claim. Use it for "does ariadne
handle X".

## 3. `README.md` and `docs/HOW_IT_WORKS.md`: how to run it and how it works

- `README.md` is the usage reference: build, input preparation, every flag,
  environment variables, network/cache, snapshots, library mode, runtime
  evidence, the research commands, outputs and exit codes.
- `docs/HOW_IT_WORKS.md` (English) and `docs/JAK_TO_DZIALA.md` (Polish) describe
  the pipeline step by step and how to read the verdicts. **The two are kept in
  sync**: a change to one is made to the other in the same commit.

## 4. History, when needed

`archive/2026-11-25/docs/` holds the original `PLAN.md` (phases 0–12 with their
status notes), the FAQs, the full earlier README (every model described in
detail, with measurements) and the old audit. They are read-only history:
look things up there, don't update them.

## 5. Check the repo's actual state, not just the docs

```bash
git log --oneline -15
git status
```

Recent commit subjects tell you whether docs and code are in sync, and
`git status` catches anything uncommitted from a prior session.

Each profile and model has a small control application or fixture under
`testdata/` (with an `EXPECTED.MD`, often a real-JVM oracle script in
`scripts/`), and the reference apps are prepared by `scripts/prepare-*.sh`.
They are the right reference point for extending or debugging behavior.

## What "done" looks like for a change in this repo

Every substantive change ships as: the code change, a real verification (not
just unit tests: run `bin/ariadne` against a real small app or fixture, often
with a JVM oracle), and an update to whatever it affects in the same commit:
`README.md` (flags, usage), `docs/HOW_IT_WORKS.md` + `docs/JAK_TO_DZIALA.md`
(behavior), `docs/SUPPORT.MD` (capability), and an entry in
`RECOMMENDATION/IMPLEMENTATION_STATUS.MD`. "Later" tends not to happen.
