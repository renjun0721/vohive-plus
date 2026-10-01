import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'
const source = (path: string) => readFileSync(new URL('../src/' + path, import.meta.url), 'utf8')

test('classic pages retain the original surfaces and independent cards', () => {
  const css=source('style.css'); assert.match(css,/--ui-surface: rgba\(255, 255, 255, 0\.72\)/); assert.match(css,/--ui-backdrop: blur\(18px\)/); assert.match(source('views/Settings.vue'),/grid-cols-1 lg:grid-cols-2 gap-8/); assert.match(source('views/Devices.vue'),/class="ui-card p-6"/); assert.doesNotMatch(source('views/Dashboard.vue'),/<ConnectionFocusStage/);
})
