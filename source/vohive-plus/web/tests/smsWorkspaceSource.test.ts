import assert from 'node:assert/strict'
import test from 'node:test'
import { readFileSync } from 'node:fs'
const source = (path: string) => readFileSync(new URL('../src/' + path, import.meta.url), 'utf8')

test('original SMS panes keep current SIM identity, unread and historical navigation', () => {
  const s=source('views/Sms.vue'); assert.match(s,/<RecycleScroller/); for(const name of ['showDeviceSidebar','showListPane','showDetailPane']) assert.ok(s.includes('v-if="'+name+'"')); assert.match(s,/smsStore\.markThreadRead/); assert.match(s,/params\.iccid = t\.iccid/); assert.match(s,/mergeSmsThreadPages/); assert.match(s,/data-message-id="m\.id"/); assert.match(s,/@click="showLatestMessages"/); assert.match(s,/<PhoneContactsDrawer/); assert.doesNotMatch(s,/sms_thread_last_seen|localStorage/);
})
