import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'
const source = (path: string) => readFileSync(new URL('../src/' + path, import.meta.url), 'utf8')

test('classic loading panel preserves the VoHive gradient and title', () => {
  const s=source('components/LoadingScreen.vue'); assert.match(s,/from-indigo-500 to-purple-600/); assert.match(s,/\{\{ title \}\}/); assert.match(s,/loader-shimmer/);
})
