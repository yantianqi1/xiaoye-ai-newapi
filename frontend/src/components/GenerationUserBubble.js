import { computed, defineComponent, h } from 'vue'

const REFERENCE_IMAGE_ALT_PREFIX = 'reference preview'

function normalizeReferenceImages(referenceImages) {
  if (!Array.isArray(referenceImages)) return []
  return referenceImages.filter((image) => typeof image === 'string' && image.trim() !== '')
}

export default defineComponent({
  name: 'GenerationUserBubble',
  props: {
    prompt: { type: String, default: '' },
    tag: { type: String, default: '' },
    referenceImages: { type: Array, default: () => [] }
  },
  setup(props) {
    const previewImages = computed(() => normalizeReferenceImages(props.referenceImages))

    return () => h('div', { class: 'bubble user-bubble generation-user-bubble' }, [
      previewImages.value.length
        ? h('div', {
            class: [
              'generation-user-bubble__images',
              { 'generation-user-bubble__images--single': previewImages.value.length === 1 }
            ]
          }, previewImages.value.map((src, index) => h('div', {
            key: `${src}-${index}`,
            class: 'generation-user-bubble__image-item'
          }, [
            h('img', {
              class: 'generation-user-bubble__image',
              src,
              alt: `${REFERENCE_IMAGE_ALT_PREFIX} ${index + 1}`
            })
          ])))
        : null,
      props.tag ? h('span', { class: 'generation-user-bubble__tag' }, props.tag) : null,
      props.prompt ? h('p', { class: 'generation-user-bubble__text' }, props.prompt) : null
    ])
  }
})
