## 2024-10-24 - Accessibility and Focus Visible
**Learning:** Adding focus rings manually to icon-only buttons via `focus-visible` ensures accessibility.
**Action:** Consistently apply `focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none` to icon-only buttons.
## 2024-10-25 - ARIA Labels on Icon-Only Buttons
**Learning:** `title` attributes on icon-only buttons are often insufficient for screen readers without a corresponding `aria-label`.
**Action:** Always ensure icon-only buttons have an explicit `aria-label`, even if they already have a `title`.
