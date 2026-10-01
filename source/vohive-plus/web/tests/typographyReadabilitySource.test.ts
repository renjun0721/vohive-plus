import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'
const source = (path: string) => readFileSync(new URL('../src/' + path, import.meta.url), 'utf8')

test('classic typography and responsive device card grid are retained', () => {
  const shell=source('layouts/AuthenticatedShell.vue'); assert.match(shell,/Space Grotesk/); assert.match(shell,/Inter/); const dashboard=source('views/Dashboard.vue'); assert.match(dashboard,/grid-cols-1 sm:grid-cols-2 lg:grid-cols-4/); assert.match(dashboard,/lg:grid-cols-3 2xl:grid-cols-4/); assert.match(source('views/Login.vue'),/max-w-md/); assert.doesNotMatch(source('views/Login.vue'),/login-identity/);
})
