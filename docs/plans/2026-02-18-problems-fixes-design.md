# Design: PROBLEMS.md Fixes — Dark Theme Selection & SQL Playground Display

**Date**: 2026-02-18
**Status**: Approved

## Overview

Two bugs identified in PROBLEMS.md:

1. Selected text in dark mode form fields is hard to see (selection color too close to background)
2. SQL Playground enabled per-challenge does not appear on the challenge detail page

---

## Fix 1: Dark Theme Text Selection Color

### Problem

The `::selection` CSS pseudo-element is never explicitly styled. Browser defaults use a dark selection highlight that blends into the dark background (`#0f172a`), making selected text nearly invisible.

### Solution

Add `::selection` rules to `internal/views/static/css/custom.css` (currently empty, already linked in `base.html`).

```css
::selection {
    background-color: #7c3aed; /* purple-600 — matches UI accent */
    color: #ffffff;
}

.dark ::selection {
    background-color: #a855f7; /* purple-500 — lighter for dark bg */
    color: #ffffff;
}
```

Uses the existing purple accent color (`#A55EEA`) consistent with the rest of the UI.

### Files Changed

- `internal/views/static/css/custom.css` — add `::selection` rules

---

## Fix 2: SQL Playground Inline Display on Challenge Page

### Problem

The SQL Playground can be enabled per-challenge via the admin form. The data is correctly saved to the database and fetched by the challenge detail handler. However, `challenge.html` has no code to display it — the feature is invisible to users.

### Root Cause

Missing `{{if .Challenge.SQLEnabled}}` block in `challenge.html`. The global `/sql` page is fully implemented but never connected to the per-challenge view.

### Approach: Go Template Partial (Approach B)

Extract the DuckDB + CodeMirror editor into reusable Go template blocks, include them from both `challenge.html` and `sql.html`. No code duplication, clean Go template pattern.

### Design

#### 1. `sql.html` — Refactor into template blocks

Wrap the editor HTML in `{{define "sql-playground"}}...{{end}}` and the JS/script includes in `{{define "sql-scripts"}}...{{end}}`. The standalone `/sql` route continues to work by calling these blocks within its existing layout.

The template blocks receive the Challenge data through dot (`.`), reading:
- `.Challenge.SQLDatasetURL` — pre-loads the dataset in the editor
- `.Challenge.SQLSchemaHint` — displays schema reference to the user

#### 2. `challenge.html` — Add conditional SQL section

After the questions loop, add:

```html
{{if .Challenge.SQLEnabled}}
<section class="mt-8 bg-dark-surface border border-dark-border rounded-lg p-6">
  <h2 class="text-xl font-bold text-purple-400 mb-4">SQL Playground</h2>
  {{if .Challenge.SQLSchemaHint}}
  <div class="mb-4">
    <h3 class="text-sm font-medium text-gray-400 mb-2">Schema Reference</h3>
    <pre class="bg-gray-900 text-gray-100 p-3 rounded text-sm overflow-x-auto font-mono">{{.Challenge.SQLSchemaHint}}</pre>
  </div>
  {{end}}
  {{template "sql-playground" .}}
</section>
{{end}}
```

#### 3. `base.html` — Conditional script loading

Include `{{template "sql-scripts" .}}` only on pages that need DuckDB WASM (challenge pages with SQL enabled, and the `/sql` page), to avoid loading ~10MB of WASM on every page.

### Files Changed

- `internal/views/static/css/custom.css` — (same file as Fix 1, no SQL changes here)
- `internal/views/templates/sql.html` — wrap content in named template blocks
- `internal/views/templates/challenge.html` — add `{{if .Challenge.SQLEnabled}}` section with `{{template "sql-playground" .}}`
- `internal/views/templates/base.html` — conditionally include sql-scripts block

### No Backend Changes

The handler and database layer already work correctly. This is a pure template fix.

---

## Validation

After implementation, validate with agent-browser:

1. **Fix 1**: Open any form field in dark mode, select text — highlight should be purple and clearly visible
2. **Fix 2**: Create a challenge with SQL Playground enabled, visit the challenge page — editor should appear below the questions
3. **Fix 2**: Challenge with SQL disabled — no editor section should appear
4. **Fix 2**: Global `/sql` page should still work unchanged

---

## Out of Scope

- SQL Playground edit support in admin (update challenge form) — separate task
- Per-question SQL playground (current design is per-challenge)
