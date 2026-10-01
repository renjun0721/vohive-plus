import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'
const source = (path: string) => readFileSync(new URL('../src/' + path, import.meta.url), 'utf8')

test('original overview keeps its three panels and live device facts', () => {
  const s=source('components/DeviceOverviewTab.vue'); assert.match(s,/grid-cols-1 lg:grid-cols-3 gap-4/); assert.match(s,/运行状态/); assert.match(s,/useSensitiveVisibility/); assert.match(s,/e911_setup_available/); assert.match(s,/hasValidSignalDbm/); assert.match(source('views/Devices.vue'),/retrySIMPin/);
})
