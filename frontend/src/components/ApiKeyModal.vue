<script setup>
import { ref, watch, computed } from 'vue'
import { NModal, NInput, NButton } from 'naive-ui'
import { useI18n } from 'vue-i18n'
import { useUserStore } from '../stores/user'

const { t } = useI18n()
const userStore = useUserStore()

const keyInput = ref(userStore.apiKey || '')

watch(() => userStore.showApiKeyModal, (show) => {
  if (show) {
    keyInput.value = userStore.apiKey || ''
  }
})

const isSet = computed(() => !!userStore.apiKey)
const maskedKey = computed(() => {
  const k = userStore.apiKey || ''
  if (!k) return ''
  if (k.length <= 10) return k.slice(0, 2) + '••••'
  return k.slice(0, 6) + '••••••' + k.slice(-4)
})

const handleSave = () => {
  const key = keyInput.value.trim()
  userStore.setApiKey(key)
  userStore.closeApiKeyModal()
}

const handleClear = () => {
  keyInput.value = ''
  userStore.setApiKey('')
  userStore.closeApiKeyModal()
}
</script>

<template>
  <NModal
    :show="userStore.showApiKeyModal"
    preset="card"
    :bordered="false"
    :mask-closable="true"
    :closable="false"
    :show-header="false"
    class="apikey-modal"
    style="max-width: 480px; border-radius: 20px;"
    @update:show="v => { if (!v) userStore.closeApiKeyModal() }"
  >
    <div class="apikey-wrap">
      <!-- Header -->
      <div class="apikey-head">
        <div class="apikey-icon-badge">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/>
          </svg>
        </div>
        <div class="apikey-titles">
          <div class="apikey-title-row">
            <h2 class="apikey-title">{{ t('apiKey.title') }}</h2>
            <span class="apikey-status" :class="{ on: isSet }">
              <span class="dot"></span>
              {{ isSet ? t('sidebar.apiKeySet') : t('sidebar.apiKeyNotSet') }}
            </span>
          </div>
          <p class="apikey-desc">{{ t('apiKey.description') }}</p>
        </div>
        <button class="apikey-close" @click="userStore.closeApiKeyModal()" aria-label="close">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
        </button>
      </div>

      <!-- Current key preview -->
      <div v-if="isSet" class="apikey-current">
        <span class="apikey-current-label">Current</span>
        <code class="apikey-current-value">{{ maskedKey }}</code>
      </div>

      <!-- Input -->
      <div class="apikey-field">
        <label class="apikey-label">API Key</label>
        <NInput
          v-model:value="keyInput"
          type="password"
          show-password-on="click"
          :placeholder="t('apiKey.placeholder')"
          size="large"
          round
          @keyup.enter="keyInput.trim() && handleSave()"
        />
      </div>

      <!-- Info bullets -->
      <ul class="apikey-bullets">
        <li>
          <span class="b-icon">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>
          </span>
          <span>{{ t('apiKey.hint') }}</span>
        </li>
      </ul>

      <!-- Actions -->
      <div class="apikey-actions">
        <NButton v-if="isSet" quaternary @click="handleClear">
          {{ t('apiKey.clear') }}
        </NButton>
        <NButton type="primary" size="large" :disabled="!keyInput.trim()" @click="handleSave">
          {{ t('apiKey.save') }}
        </NButton>
      </div>
    </div>
  </NModal>
</template>

<style scoped>
.apikey-wrap {
  position: relative;
  padding: 4px 4px 2px;
}

/* Header */
.apikey-head {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 18px;
}
.apikey-icon-badge {
  flex: 0 0 auto;
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  background: linear-gradient(135deg, #6366f1 0%, #8b5cf6 50%, #ec4899 100%);
  box-shadow: 0 8px 24px rgba(139, 92, 246, 0.35);
}
.apikey-icon-badge svg { width: 22px; height: 22px; }

.apikey-titles { flex: 1; min-width: 0; }
.apikey-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}
.apikey-title {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: var(--color-text-primary, #111);
  letter-spacing: -0.01em;
}
.apikey-status {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 600;
  padding: 3px 9px;
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.15);
  color: var(--color-text-muted, #94a3b8);
  border: 1px solid rgba(148, 163, 184, 0.25);
}
.apikey-status .dot {
  width: 6px; height: 6px; border-radius: 50%;
  background: currentColor;
}
.apikey-status.on {
  background: rgba(34, 197, 94, 0.12);
  color: #16a34a;
  border-color: rgba(34, 197, 94, 0.3);
}
.apikey-desc {
  margin: 6px 0 0;
  font-size: 13px;
  line-height: 1.55;
  color: var(--color-text-secondary, #6b7280);
}

.apikey-close {
  position: absolute;
  top: -2px;
  right: -2px;
  width: 30px;
  height: 30px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: var(--color-text-muted, #9ca3af);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}
.apikey-close svg { width: 16px; height: 16px; }
.apikey-close:hover {
  background: rgba(148, 163, 184, 0.15);
  color: var(--color-text-primary, #111);
}

/* Current key preview */
.apikey-current {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  margin-bottom: 14px;
  border-radius: 12px;
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.06), rgba(236, 72, 153, 0.06));
  border: 1px solid rgba(139, 92, 246, 0.18);
}
.apikey-current-label {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-muted, #9ca3af);
}
.apikey-current-value {
  flex: 1;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  color: var(--color-text-primary, #111);
  letter-spacing: 0.02em;
}

/* Field */
.apikey-field { margin-bottom: 14px; }
.apikey-label {
  display: block;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--color-text-muted, #9ca3af);
  margin-bottom: 8px;
}

/* Bullets */
.apikey-bullets {
  list-style: none;
  margin: 0 0 18px;
  padding: 12px 14px;
  border-radius: 12px;
  background: var(--color-tint-white-03, rgba(148, 163, 184, 0.07));
  border: 1px solid var(--color-border-medium, rgba(148, 163, 184, 0.15));
}
.apikey-bullets li {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  font-size: 12.5px;
  line-height: 1.55;
  color: var(--color-text-secondary, #6b7280);
}
.b-icon {
  flex: 0 0 auto;
  width: 22px;
  height: 22px;
  border-radius: 7px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(99, 102, 241, 0.12);
  color: #6366f1;
}
.b-icon svg { width: 13px; height: 13px; }

/* Actions */
.apikey-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
.apikey-actions :deep(.n-button) {
  min-width: 96px;
}
</style>

<style>
.apikey-modal {
  border-radius: 20px !important;
  box-shadow: 0 30px 80px rgba(0, 0, 0, 0.25) !important;
}
.apikey-modal .n-card__content {
  padding: 24px 24px 22px !important;
}
</style>
