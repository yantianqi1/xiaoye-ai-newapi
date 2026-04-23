# Generate Session UI Design

**Goal:** Make `/generate` open as a fresh creation workspace every time while keeping existing generated images and videos available from `/assets`.

## Confirmed Behavior

- `/generate` is a current-session creation surface, not a history surface.
- Opening `/generate` must not automatically render previous generation history.
- Clicking `新对话` clears only the visible session UI and composer inputs.
- No backend delete API is called by this feature.
- Existing generated media remains stored and visible from `/assets`.
- `/assets` remains the history and asset management page.

## Approach

Use a local session-only state in `Generate.vue`. Remove automatic history rendering from the page and keep `currentResults` as the only list displayed in the timeline. Keep the existing `genStore.pendingResult` path so an item opened from `/assets` can still be shown in `/generate` intentionally.

Expose a `reset` method from `ComposerBar.vue` that clears prompt text, uploaded reference images, video frame uploads, and prompt optimization UI. `Generate.vue` calls it from the new `新对话` action.

## UI

- Add a compact session toolbar above the timeline.
- Primary action: `新对话`, clears the current visible session.
- Secondary action: `查看历史`, routes to `/assets`.
- Empty state copy should describe the new session behavior, not missing historical records.

## Testing

- Add node tests for pure session helper functions to verify:
  - historical timeline groups are hidden for new sessions;
  - current session results still render when present;
  - clearing the session removes visible current results without deleting source records.

## Non-Goals

- No multi-session management.
- No history toggle inside `/generate`.
- No backend storage deletion.
- No fallback or mock success path.
