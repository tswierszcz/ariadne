---
name: archive-completed
description: Use this right after a RECOMMENDATION phase (or any other planned item from PLAN.md/RECOMMENDATION) actually ships and is verified — when the user says something is done, finished, merged, or asks to "zarchiwizuj", "oznacz fazę jako zrobioną", "archive this". Moves the completed planning content into archive/<date>/, updates docs/SUPPORT.MD's status table, and updates RECOMMENDATION/IMPLEMENTATION_STATUS.MD. Do not invoke this on a plan or proposal alone — only after real, verified completion.
---

# Archiving a completed item

Purpose: once a planned item is actually done, move its planning text out of the
active backlog into a dated archive record, and make the two "current state" documents
(`docs/SUPPORT.MD`, `RECOMMENDATION/IMPLEMENTATION_STATUS.MD`) reflect reality. This keeps
`RECOMMENDATION/` and `PLAN.md` from accumulating stale, already-finished backlog items
next to the still-open ones.

## 0. Confirm it's actually done, not just coded

Do not run this on a plan, a proposal, or code that hasn't been verified. Require:
- the code change itself (or a clear statement of what changed, if this is a
  documentation-only/contract-only phase);
- a real verification already performed (tests run, or an end-to-end `bin/ariadne` run —
  matching this repo's own standard from the catch-up skill: "code change + a real
  verification + a doc update, in the same commit").

If verification is missing, say so and stop — ask the user to run it, or run it yourself,
before archiving. Archiving a half-done item just hides the gap.

## 1. Identify exactly what completed

Ask for (or infer from the conversation, confirming back) one of:
- a RECOMMENDATION phase id, `AREA:PHASE` (e.g. `SERVICE_LOADING:S1`, `CVE:F4`) — bare `F2`
  or `R0` is ambiguous across documents, always use the full id;
- a `PLAN.md` phase or sub-bullet not tracked in RECOMMENDATION at all (e.g. one of this
  repo's own Phase 9-13 backlog items).

## 2. Build the archive path

`archive/<YYYY-MM-DD>/<area-or-feature-slug>/`, date = today. Slug: the RECOMMENDATION
area name lowercased (`service_loading`), or a short kebab-case name for a PLAN.md-only
item (`cves-file-input`).

## 3. Move the completed content

The exact move depends on where the item lives — **PLAN.md's own convention is to never
move or delete its content** (status notes accumulate in place; see the catch-up skill),
so only RECOMMENDATION content actually moves:

- **A single completed phase within a RECOMMENDATION area** (most common case): extract
  that phase's text from `RECOMMENDATION/<AREA>_IMPLEMENTATION_PLAN.MD` into a new file
  `archive/<date>/<area>/<AREA>_<PHASE>.md`, prefixed with a short completion record (date,
  what shipped, files touched, verification commands/evidence, remaining limitations).
  Replace the phase's text in the live implementation-plan doc with a one-line pointer to
  the archive file; leave every other phase in that doc untouched.
- **Every phase in a RECOMMENDATION area is now done**: move the whole pair
  (`<AREA>.MD` + `<AREA>_IMPLEMENTATION_PLAN.MD`) into `archive/<date>/<area>/` and remove
  that area's row from `RECOMMENDATION/README.MD`'s index table (both the "Dokumenty według
  obszaru" and "Pokrycie pierwotnych rekomendacji" tables).
- **A `PLAN.md`-only item (not tracked in RECOMMENDATION)**: do not touch `PLAN.md`. Instead
  write a completion record to `archive/<date>/<feature-slug>/SUMMARY.md` (same fields as
  above) purely as a dated record — `PLAN.md`'s own status note, updated in place as usual,
  stays the actual source of truth for that item.

## 4. Update `docs/SUPPORT.MD`

Flip the relevant capability row's status (`Not implemented`/`Partial` → `Partial`/
`Implemented` as the evidence actually supports — never jump straight to `Implemented` if
real gaps remain), update "Available scope and conditions" and "Limits and evidence" with
the new behavior and links to the actual changed files/tests, per that document's own
"Reading the status" rules. Don't describe a phase as full support if it isn't.

## 5. Update `RECOMMENDATION/IMPLEMENTATION_STATUS.MD` (RECOMMENDATION items only)

Add/update the phase's row per the "Rejestr faz" format already specified there: full
phase id and state (`DONE`, or `NEEDS_VERIFICATION` if some criteria are still open),
scope actually executed, changed files and user-visible behavior, met/unmet criteria with
evidence links, commands actually run, which of SUPPORT/README/report/FAQ changed (or why
not), remaining limitations, and the next step. Update the area's row in the overview table
too — don't mark a whole area `DONE` after one phase.

## 6. Report back

Summarize: what moved where, what `docs/SUPPORT.MD`/`IMPLEMENTATION_STATUS.MD` rows
changed, and whether `README.md`/`FAQ.md` also need a touch-up (they already point readers
at `docs/SUPPORT.MD`/`RECOMMENDATION` rather than duplicating status, so usually no change
is needed there — only if their own prose directly described the now-completed limitation).
