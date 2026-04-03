# OpenAI-Compatible Image Provider Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a dedicated OpenAI-compatible image provider for NewAPI-compatible upstreams without changing existing Gemini or Volcengine behavior.

**Architecture:** Add one new backend image provider registered under a stable internal model ID and drive its upstream model/base URL from environment variables. Keep the frontend on the existing `/api/models` and `/api/pricing` contract, but add model-specific constraints for supported aspect ratios and inpainting.

**Tech Stack:** Go, Gin, Vue 3, Vite, Pinia

---

### Task 1: Add failing backend provider tests

**Files:**
- Create: `backend/internal/provider/openai_compatible_test.go`

**Step 1: Write the failing generation test**

Assert that a text-to-image request calls `/v1/images/generations` with bearer auth, configured model, `response_format=b64_json`, and a supported size.

**Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-build go test ./internal/provider -run TestOpenAICompatibleImageProviderGenerateImageUsesOpenAIImagesEndpoint`

Expected: FAIL because the provider does not exist yet.

**Step 3: Write the failing edit test**

Assert that an edit request calls `/v1/images/edits` as multipart form data and uploads the input image and mask.

**Step 4: Run test to verify it fails**

Run: `GOCACHE=/tmp/go-build go test ./internal/provider -run TestOpenAICompatibleImageProviderGenerateImageUsesOpenAIEditsEndpoint`

Expected: FAIL because the provider does not exist yet.

### Task 2: Implement backend config and provider

**Files:**
- Modify: `backend/internal/config/config.go`
- Create: `backend/internal/provider/openai_compatible.go`

**Step 1: Add config getters**

Add env getters for OpenAI-compatible base URL, API key, upstream image model, display name, and credits.

**Step 2: Implement minimal provider**

Implement the current `ImageGenerator` interface with:
- explicit model availability checks
- JSON request builder for text-to-image
- multipart request builder for image edits
- strict response parsing for `b64_json`

**Step 3: Register the provider**

Register a stable internal model ID so the current `/api/generate` flow can select it without backend routing changes.

### Task 3: Expose the new model to the existing frontend contract

**Files:**
- Modify: `backend/internal/api/pricing.go`
- Modify: `backend/.env.example`

**Step 1: Add model name and pricing entry**

Expose a `1K` price tier only, matching the new provider’s supported size contract.

**Step 2: Document environment variables**

Add example env vars for the new upstream.

### Task 4: Restrict frontend behavior to supported capabilities

**Files:**
- Create: `frontend/src/utils/imageModelCapabilities.js`
- Modify: `frontend/src/components/ComposerBar.vue`
- Modify: `frontend/src/components/ImageEditor.vue`
- Modify: `frontend/src/views/Generate.vue`

**Step 1: Add model capability helpers**

Encode supported aspect ratios and inpainting support for the new model.

**Step 2: Restrict ratios in composer and editor**

Show only supported aspect ratios for the OpenAI-compatible image model.

**Step 3: Make inpainting explicit**

Route inpainting through the selected supported model and block unsupported models with a visible error instead of silent fallback.

### Task 5: Verify

**Files:**
- Test: `backend/internal/provider/openai_compatible_test.go`

**Step 1: Run provider tests**

Run: `GOCACHE=/tmp/go-build go test ./internal/provider`

**Step 2: Run targeted frontend build**

Run: `npm run build`

Workdir: `frontend`

**Step 3: Run targeted backend build**

Run: `GOCACHE=/tmp/go-build go test ./...`

Workdir: `backend`
