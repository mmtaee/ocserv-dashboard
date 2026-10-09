---
name: pagination-ui-ux
description: Design, implement, or review pagination for tables and lists, including server/client ownership, centered layout, responsive behavior, accessibility, RTL, loading, mutation, and boundary states.
---

# Pagination UI/UX

Build pagination that makes location, available navigation, and request state clear without duplicating data ownership.

## Data ownership

- Determine whether pagination is server-owned or client-owned before wiring controls.
- For server pagination, render the returned page directly and send page changes through the existing fetch path. Never paginate the returned page again on the client.
- Use the API's current page, page size, and total-record count as the source of truth. Derive page count and disabled states from them.
- Reset to page one when a server-side filter changes. After deletion, clamp or move to the previous valid page before refreshing.

## Placement and hierarchy

- Center the primary pager relative to its containing card or viewport.
- When status text and pagination share a footer, do not use `justify-between`; unequal side content makes the pager visually off-center. Prefer a symmetric layout such as `grid-cols-[1fr_auto_1fr]`, with status in the first column and pagination in the center column.
- On narrow screens, stack status and controls, center both, and avoid horizontal overflow.
- Keep page status secondary with muted text. Pagination remains the primary interactive element.

## Controls

- Reuse the project's Pagination component and button variants instead of recreating links or page-number state.
- Show a clear active page, previous/next controls, first and last edges when useful, and ellipses for skipped ranges.
- Keep one or two sibling pages around the active page; avoid rendering long runs of page buttons.
- Disable unavailable directions on the first and last pages. Disable navigation while a page request is pending to prevent duplicate or out-of-order requests.
- Preserve a stable footer while loading. Use skeletons for initial table loading and keep existing data visible during ordinary page transitions when the project pattern supports it.

## Accessibility and localization

- Give previous, next, first, last, and ellipsis controls localized visible or accessible labels. Icon-only controls require an accessible name.
- Preserve the component's current-page semantics and keyboard behavior; do not replace semantic buttons or links with clickable generic elements.
- Use logical layout utilities. In RTL locales, reverse directional chevrons while keeping previous/next meaning correct.
- Keep interactive targets comfortably sized on touch screens and maintain visible focus styles.

## Verification

Check zero, one, and many-page datasets; first, middle, and last pages; loading and request failure; deletion of the last row on a page; narrow layouts; keyboard navigation; and at least one RTL locale.
