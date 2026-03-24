## 2024-10-24 - Accessibility and Focus Visible
**Learning:** Adding focus rings manually to icon-only buttons via `focus-visible` ensures accessibility.
**Action:** Consistently apply `focus-visible:ring-2 focus-visible:ring-blue-500/50 outline-none` to icon-only buttons.
## 2024-10-25 - ARIA Labels on Icon-Only Buttons
**Learning:** `title` attributes on icon-only buttons are often insufficient for screen readers without a corresponding `aria-label`.
**Action:** Always ensure icon-only buttons have an explicit `aria-label`, even if they already have a `title`.
## 2026-03-16 - [Empty States]
**Learning:** Empty states should explain the context (e.g., empty search vs empty folder) and provide contextual actions to avoid "dead ends" for users, while hiding unnecessary actions when at the root directory.
**Action:** Always provide descriptive text and relevant primary actions (like 'Clear Search' or 'Back to Library') when designing empty states, ensuring buttons have proper keyboard accessibility focus states.
## 2024-10-25 - Focus Styles on Complex Visual Layers
**Learning:** Default browser outlines are insufficient for accessibility on custom interactive UI components with complex visual layers (e.g., Waterfall view items with overlays or gradients).
**Action:** Always explicitly add focus-visible styling (like `focus-visible:ring-4`) and appropriate `aria-label` attributes to the wrapping `<button>` to ensure keyboard focus visibility.
## 2026-03-24 - Add active state accessibility to toggle buttons
**Learning:** Visually styled toggle buttons (e.g., layout mode or theme toggles) need explicit `aria-pressed` attributes to convey their active state to screen readers. Wrapping related toggles in a `role="group"` with an `aria-label` further improves context.
**Action:** Always verify that custom toggle interactions update `aria-pressed` or `aria-current` rather than relying solely on visual class changes.
