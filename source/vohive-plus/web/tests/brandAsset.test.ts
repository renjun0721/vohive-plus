import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'
const source = (path: string) => readFileSync(new URL('../src/' + path, import.meta.url), 'utf8')

test('personal title keeps the original VoHive icon', () => {
  const html=readFileSync(new URL('../index.html',import.meta.url),'utf8');const icon=readFileSync(new URL('../public/favicon.svg',import.meta.url),'utf8');assert.match(html,/<title>VoHive Plus<\/title>/);assert.match(icon,/<svg/);assert.match(html,/Inter/);
})
