export const OPENAI_COMPAT_IMAGE_MODEL_ID = 'openai-compatible-image'

const GEMINI_FLASH_IMAGE_MODEL_ID = 'gemini-3.1-flash-image-preview'
const INPAINT_MODEL_IDS = new Set([
  'gemini-3-pro-image-preview',
  GEMINI_FLASH_IMAGE_MODEL_ID,
  OPENAI_COMPAT_IMAGE_MODEL_ID,
])
const OPENAI_COMPAT_RATIO_VALUES = new Set(['1:1', '2:3', '3:2'])

export function filterRatiosForModel(modelId, baseRatios, extraRatios = []) {
  if (modelId === GEMINI_FLASH_IMAGE_MODEL_ID) {
    return [...baseRatios, ...extraRatios]
  }
  if (modelId === OPENAI_COMPAT_IMAGE_MODEL_ID) {
    return baseRatios.filter(ratio => OPENAI_COMPAT_RATIO_VALUES.has(ratio.value))
  }
  return [...baseRatios]
}

export function modelSupportsInpainting(modelId) {
  return INPAINT_MODEL_IDS.has(modelId)
}
