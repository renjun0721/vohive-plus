import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'
const source = (path: string) => readFileSync(new URL('../src/' + path, import.meta.url), 'utf8')

test('classic proxy tabs retain outbound and upstream configuration', () => {
  const s=source('views/Proxy.vue'); assert.match(s,/PageHeader title="代理管理"/); assert.match(s,/activeTab/); assert.match(s,/instancesWithStatus/); assert.match(s,/upstreamProxiesWithRuleCount/); assert.match(s,/createLatestRequestGate/); assert.match(s,/showUpstreamSaveResult/);
})
