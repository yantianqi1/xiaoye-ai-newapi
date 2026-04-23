# Generate Session UI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make `/generate` open as a fresh current-session workspace and add a manual new conversation action without deleting generated media.

**Architecture:** Keep generation history in `useGenerationStore` for `/assets`, but remove history rendering from `Generate.vue`. Add small pure helpers for session display behavior, then use them from tests and the Vue page. `ComposerBar.vue` exposes one reset method so the parent can clear composer-only state.

**Tech Stack:** Vue 3, Pinia, Vue Router, Naive UI, Node test runner.

---

### Task 1: Add Session View Tests

**Files:**
- Create: `frontend/src/utils/generateSessionView.js`
- Create: `frontend/tests/generate-session-view.test.mjs`

**Step 1: Write the failing test**

Test that history groups are not included in the generate page session view, current session results are included, and clearing removes only visible session items.

**Step 2: Run test to verify it fails**

Run: `cd frontend && node --test tests/generate-session-view.test.mjs`

Expected: FAIL because `generateSessionView.js` does not exist yet.

**Step 3: Write minimal implementation**

Create pure helpers:

- `buildGenerateSessionView({ currentResults })`
- `clearGenerateSessionResults()`

**Step 4: Run test to verify it passes**

Run: `cd frontend && node --test tests/generate-session-view.test.mjs`

Expected: PASS.

### Task 2: Update Generate Page

**Files:**
- Modify: `frontend/src/views/Generate.vue`
- Modify: `frontend/src/locales/zh.json`
- Modify: `frontend/src/locales/en.json`

**Step 1: Remove automatic history rendering**

Stop rendering `timelineGroups` in `Generate.vue`. Keep `genStore.pendingResult` handling for explicit item handoff from `/assets`.

**Step 2: Add session toolbar**

Add `新对话` and `查看历史` actions. `新对话` clears `currentResults` and calls `composerRef.value?.reset()`. `查看历史` routes to `/assets`.

**Step 3: Update empty copy**

Add locale keys for new session copy and buttons.

**Step 4: Verify function lengths**

Keep new functions short and avoid new nested branching.

### Task 3: Add Composer Reset

**Files:**
- Modify: `frontend/src/components/ComposerBar.vue`

**Step 1: Extract reset helper**

Create `resetComposer()` to clear prompt, uploaded image state, optimization state, upload expansion, and video frame state.

**Step 2: Reuse helper after send**

Replace existing post-submit manual clearing with `resetComposer()`.

**Step 3: Expose helper**

Expose `reset: resetComposer` from `defineExpose`.

### Task 4: Verify

**Files:**
- `frontend/tests/generate-session-view.test.mjs`
- `frontend/tests/generation-user-bubble.test.mjs`
- Frontend build output

**Step 1: Run targeted tests**

Run: `cd frontend && node --test tests/generate-session-view.test.mjs tests/generation-user-bubble.test.mjs`

Expected: PASS.

**Step 2: Run production build**

Run: `cd frontend && npm run build`

Expected: PASS.
