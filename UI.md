# UI Design & Refinement Log: Invisible Archive

This document tracks the aesthetic decisions, layout optimizations, and accessibility improvements made to the Invisible Archive frontend.

## 1. Core Visual Language
*   **Typography:** Using `Inter` system-ui stack for maximum readability and a clean, modern feel.
*   **Theming:** 
    *   **Light Mode:** Clean `#f8fafc` background with `slate-900` text.
    *   **Dark Mode (Dracula):** Strict adherence to the official Dracula palette (`#282a36` background, `#f8f8f2` foreground).
    *   **Accents:** Using **Dracula Purple** (`#bd93f9`) for interactive elements (loading bars, focus rings, hover states) in dark mode to ensure a vibrant, branded experience.

---

## 2. Layout Optimizations

### Grid (Big Icon) View
*   **Alignment Strategy:** Icons are wrapped in `shrink-0` containers to prevent compression. Filenames are placed in a fixed-height container (`h-8 sm:h-10`) to ensure icons stay perfectly aligned in their rows regardless of name length.
*   **Spacing:** Balanced gaps (`gap-4 sm:gap-10`) and `justify-center` alignment to make the grid fit various screen sizes without excessive whitespace.
*   **Responsive Columns:** Dynamically calculated based on container width (from 2 columns on mobile to 10 on ultra-wide).

### Details (Table) View
*   **Embedded Sticky Header:** A dedicated `sticky top-0` header row (Name, Size, Modified) that aligns perfectly with the data columns.
*   **Integrated Metadata:** The "Item Count" badge is embedded directly into the sticky header in this mode to reduce visual clutter and maximize vertical space.

---

## 3. Interactive Features

### Video Player (Plyr Integration)
*   **Drag-to-Seek:** Robust horizontal drag/swipe gestures on the video surface allow for high-precision seeking.
*   **Seek Overlay:** An immersive, centered UI overlay appears during seeking, showing the time delta (e.g., `+15s`) and current progress.
*   **Refined Controls:** Custom-styled Plyr controls that match the Dracula theme.

### Image Viewer (PhotoSwipe v5)
*   **Integrated History:** Opening an image pushes a state to the browser history, allowing the "Back" button or gesture to close the gallery naturally.
*   **Dynamic Resolution:** Aspect ratios are recalculated on-the-fly as high-res images load to ensure perfect fit and zoom stability.

---

## 4. Performance & Loading
*   **Non-Blocking Progress Bar:** Replaced full-screen blur loaders with a sleek, indeterminate progress bar pinned to the top of the content area.
*   **Background Preloading:** When previewing an item, the engine automatically pre-fetches the next two images in the background to ensure instantaneous transitions.
*   **Virtualization:** `TanStack Virtual` handles 100k+ item lists with minimal DOM nodes, ensuring 60fps scrolling even on low-end mobile devices.

---

## 5. Accessibility & Standards
*   **Semantic HTML:** Items are rendered as `<button type="button">` to support native keyboard focus and activation.
*   **Focus States:** High-visibility focus rings (`focus-visible:ring-2`) use the theme's accent colors.
*   **PWA Ready:** Full manifest and multi-size icon support for high-quality "Add to Home Screen" experiences on iOS and Android.

---

## 6. Identified Areas for Improvement (UI Audit)
Based on comprehensive Playwright UI testing, the following areas have been identified for future refinement:

### Mobile Experience
*   **Navigation Discoverability:** "Next" and "Prev" navigation buttons in the preview stage are hidden on mobile screens. We should consider adding explicit, translucent, floating tap zones on the screen edges for users who may not discover the swipe gesture.
*   **Layout Cycler Feedback:** The "Cycle layout" button on mobile swaps icons without textual feedback. Adding a small transient toast notification (e.g., "Grid View", "List View") upon tapping would enhance clarity.
*   **Scrubbing Precision:** The Plyr progress bar is thin and can be difficult to scrub accurately on mobile devices. Adding a custom CSS override to increase the height of the `plyr__progress` touch target would improve usability.

### Desktop Layout & Accessibility
*   **Empty States:** Completely empty folders currently display a blank background. Introducing a friendly illustration or clear "This folder is empty" state would provide better user context.
*   **Ultra-Wide Screens:** At `xl` and `2xl` breakpoints, grid items can appear overly spread out. Implementing a `max-w-7xl` container or increasing the column count on larger screens would optimize horizontal space utilization.
*   **Active State Contrast:** Ensure the active state for layout selection buttons (currently `bg-white text-blue-600` vs `text-slate-400`) meets WCAG AA contrast standards, particularly in dark mode.
