import test from 'node:test'
import assert from 'node:assert/strict'

import {
  buildGenerateSessionTimeline,
  clearGenerateSessionResults,
  hasGenerateSessionContent
} from '../src/utils/generateSessionView.js'

test('generate page timeline stays empty when there are no current session results', () => {
  const timeline = buildGenerateSessionTimeline([], '当前会话')

  assert.deepEqual(timeline, [])
})

test('generate page timeline only includes current session results', () => {
  const currentResults = [
    { id: 'now-1', prompt: '山景油画', status: 'success' },
    { id: 'now-2', prompt: '人物转油画', status: 'generating' }
  ]

  const timeline = buildGenerateSessionTimeline(currentResults, '当前会话')

  assert.equal(timeline.length, 1)
  assert.equal(timeline[0].label, '当前会话')
  assert.deepEqual(timeline[0].items, currentResults)
})

test('clearing the visible generate session does not mutate original results', () => {
  const currentResults = [{ id: 'keep-1', prompt: '保留在资产页', status: 'success' }]

  const cleared = clearGenerateSessionResults(currentResults)

  assert.deepEqual(cleared, [])
  assert.deepEqual(currentResults, [{ id: 'keep-1', prompt: '保留在资产页', status: 'success' }])
})

test('generate page content flag depends only on current session results', () => {
  assert.equal(hasGenerateSessionContent([]), false)
  assert.equal(hasGenerateSessionContent([{ id: 'x' }]), true)
})
