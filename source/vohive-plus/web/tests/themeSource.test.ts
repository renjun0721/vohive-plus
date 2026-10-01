import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'
const source = (path: string) => readFileSync(new URL('../src/' + path, import.meta.url), 'utf8')

test('classic bootstrap and CSS use original light and dark canvases', () => {
  const css=source('style.css'); assert.match(css,/--ui-bg: #f9fafb/); assert.match(css,/--ui-bg: #101014/); assert.match(css,/--ui-radius-lg: 16px/); const shell=source('layouts/AuthenticatedShell.vue'); assert.match(shell,/232px/); assert.match(shell,/52px/); assert.match(shell,/max-width: 767px/); assert.doesNotMatch(shell,/flow-shell|topbar-product/);
})
