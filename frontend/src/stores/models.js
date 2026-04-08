import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

export const useModelsStore = defineStore('models', () => {
  // State
  const models = ref([])
  const loading = ref(false)
  const loaded = ref(false)
  const defaultModelId = ref('')

  // Getters
  const imageModels = computed(() => {
    return models.value.filter(m => m.type === 'image')
  })

  const videoModels = computed(() => {
    return models.value.filter(m => m.type === 'video')
  })

  const availableModels = computed(() => {
    return models.value
  })

  const getModelById = computed(() => (id) => {
    return models.value.find(m => m.id === id) || null
  })

  const getDisplayName = computed(() => (id) => {
    const model = models.value.find(m => m.id === id)
    return model ? model.name : null
  })

  const ensureModelId = computed(() => (id) => {
    const validIds = models.value.map(m => m.id)
    if (validIds.includes(id)) return id
    if (defaultModelId.value) return defaultModelId.value
    if (validIds.length > 0) return validIds[0]
    return id
  })

  // Actions
  async function loadModels(force = false) {
    if (loaded.value && !force) return

    loading.value = true
    try {
      const { data } = await axios.get('/api/models')
      models.value = data.models || []

      // Default model: first image model
      const firstImage = models.value.find(m => m.type === 'image')
      if (firstImage) {
        defaultModelId.value = firstImage.id
      } else if (models.value.length > 0) {
        defaultModelId.value = models.value[0].id
      }

      loaded.value = true
    } catch (e) {
      console.error('Failed to load models:', e)
      models.value = []
    } finally {
      loading.value = false
    }
  }

  function reset() {
    models.value = []
    loading.value = false
    loaded.value = false
    defaultModelId.value = ''
  }

  return {
    models,
    loading,
    loaded,
    defaultModelId,
    imageModels,
    videoModels,
    availableModels,
    getModelById,
    ensureModelId,
    getDisplayName,
    loadModels,
    reset
  }
})
