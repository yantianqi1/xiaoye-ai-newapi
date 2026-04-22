import { ref } from 'vue'
import axios from 'axios'
import { useUserStore } from '../stores/user'

const IMAGE_UPLOAD_FIELD = 'file'
const DEFAULT_IMAGE_FILENAME = 'upload.png'

export function useGenerate() {
  const userStore = useUserStore()
  const activePolls = ref({})

  function buildHeaders() {
    const headers = { Authorization: `Bearer ${userStore.token}` }
    if (userStore.apiKey) {
      headers['X-User-Api-Key'] = userStore.apiKey
    }
    return headers
  }

  // 统一生成接口（支持 image/video/ecommerce）
  async function generate(type, payload) {
    if (!userStore.apiKey) {
      userStore.openApiKeyModal()
      throw new Error('请先设置 API Key')
    }
    const response = await axios.post('/api/generate', {
      type,
      ...payload
    }, {
      headers: buildHeaders(),
      timeout: 30000
    })
    return { task_id: response.data.task_id }
  }

  // 兼容旧接口
  async function generateImage(payload) {
    return generate('image', payload)
  }

  async function generateEcommerce(payload) {
    return generate('ecommerce', payload)
  }

  async function generateVideo(payload) {
    return generate('video', payload)
  }

  // 获取单个生成记录详情（统一查询接口）
  async function getGeneration(id) {
    const { data } = await axios.get(`/api/generations/${id}`, {
      headers: buildHeaders()
    })
    return data
  }

  // 统一的轮询任务状态（支持 image/ecommerce/video）
  async function pollTask(taskId, onUpdate) {
    if (activePolls.value[taskId]) return
    activePolls.value[taskId] = true

    while (activePolls.value[taskId]) {
      try {
        // 使用 generations API 查询状态（通用的生成记录查询）
        const { data: gen } = await axios.get(`/api/generations/${taskId}`, {
          headers: buildHeaders()
        })
        if (!activePolls.value[taskId]) break

        // 状态映射：success -> 完成，failed -> 失败，其他 -> 进行中
        if (gen.status === 'success') {
          delete activePolls.value[taskId]
          onUpdate({ 
            status: 'success', 
            images: gen.images || [], 
            video_url: gen.video_url,
            credits_spent: gen.credits_cost 
          })
          return
        }
        if (gen.status === 'failed') {
          delete activePolls.value[taskId]
          onUpdate({ status: 'failed', error_msg: gen.error_msg || '生成失败' })
          return
        }
        // generating, queued, running 都显示为进行中
        onUpdate({ status: gen.status || 'generating' })
      } catch (e) {
        console.error(`[Poll] 网络错误 task=${taskId}:`, e?.message || e)
      }
      await new Promise(r => setTimeout(r, 3000))  // 3秒轮询一次
    }
  }

  // 兼容旧接口
  async function pollVideoTask(taskId, onUpdate) {
    return pollTask(taskId, onUpdate)
  }

  function stopAllPolls() {
    Object.keys(activePolls.value).forEach(k => delete activePolls.value[k])
  }

  function dataUrlToBlob(dataUrl) {
    const [meta, data] = dataUrl.split(',')
    if (!meta || !data) throw new Error('无效的图片数据')
    const mimeType = meta.match(/^data:(.*?);base64$/)?.[1] || 'image/png'
    const binary = atob(data)
    const bytes = new Uint8Array(binary.length)
    for (let i = 0; i < binary.length; i += 1) bytes[i] = binary.charCodeAt(i)
    return new Blob([bytes], { type: mimeType })
  }

  function buildImageUploadFormData(source) {
    const form = new FormData()
    if (typeof File !== 'undefined' && source instanceof File) {
      form.append(IMAGE_UPLOAD_FIELD, source, source.name)
      return form
    }
    if (typeof Blob !== 'undefined' && source instanceof Blob) {
      form.append(IMAGE_UPLOAD_FIELD, source, DEFAULT_IMAGE_FILENAME)
      return form
    }
    if (typeof source === 'string' && source.startsWith('data:')) {
      const blob = dataUrlToBlob(source)
      const extension = blob.type.split('/')[1] || 'png'
      form.append(IMAGE_UPLOAD_FIELD, blob, `upload.${extension}`)
      return form
    }
    throw new Error('不支持的图片上传数据')
  }

  async function uploadImageToOSS(source) {
    const formData = buildImageUploadFormData(source)
    const response = await axios.post('/api/user/upload/image', formData, {
      headers: buildHeaders()
    })
    return response.data.url
  }

  async function optimizePrompt(payload) {
    const response = await axios.post('/api/prompt/optimize', payload, {
      headers: buildHeaders(),
      timeout: 45000
    })
    return response.data
  }

  async function reversePrompt(payload) {
    const response = await axios.post('/api/tools/reverse-prompt', payload, {
      headers: buildHeaders(),
      timeout: 60000
    })
    return response.data
  }

  return {
    generate,
    generateImage,
    generateEcommerce,
    generateVideo,
    pollTask,
    pollVideoTask,
    getGeneration,
    stopAllPolls,
    uploadImageToOSS,
    optimizePrompt,
    reversePrompt,
    activePolls
  }
}
