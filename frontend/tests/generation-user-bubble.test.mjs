import test from 'node:test'
import assert from 'node:assert/strict'
import { createSSRApp, h } from 'vue'
import { renderToString } from 'vue/server-renderer'
import GenerationUserBubble from '../src/components/GenerationUserBubble.js'

async function renderBubble(props) {
  const app = createSSRApp({
    render: () => h(GenerationUserBubble, props)
  })

  return renderToString(app)
}

test('renders uploaded reference image previews in the user bubble', async () => {
  const html = await renderBubble({
    prompt: '转换成油画风格',
    tag: '图片生成',
    referenceImages: ['https://example.com/ref-1.png', 'https://example.com/ref-2.png']
  })

  assert.match(html, /转换成油画风格/)
  assert.match(html, /https:\/\/example\.com\/ref-1\.png/)
  assert.match(html, /https:\/\/example\.com\/ref-2\.png/)
  assert.match(html, /generation-user-bubble__images/)
})

test('does not render preview images when no reference images are provided', async () => {
  const html = await renderBubble({
    prompt: '只生成文字',
    tag: '图片生成',
    referenceImages: []
  })

  assert.match(html, /只生成文字/)
  assert.doesNotMatch(html, /generation-user-bubble__images/)
  assert.doesNotMatch(html, /<img/)
})
