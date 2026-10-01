import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

async function source(path: string) {
  return readFile(new URL(path, import.meta.url), 'utf8')
}

test('personal console extends the original VoHive identity', async () => {
  const [shell,login,header,settings]=await Promise.all(['../src/layouts/AuthenticatedShell.vue','../src/views/Login.vue','../src/components/PageHeader.vue','../src/views/Settings.vue'].map(source))
  assert.match(shell, /sidebar-brand-title">VoHive/); assert.match(shell, /classic-plus/); assert.match(login, /VoHive Plus/); assert.match(header, /title/); assert.match(settings, /title="系统设置"/);
})

test('frontend integration identifiers use the HiDeck namespace', async () => {
  const [app, systemService, websheet, sensitive, phoneSession, settingsStore] = await Promise.all([
    source('../src/App.vue'),
    source('../src/services/system.ts'),
    source('../src/components/CarrierWebsheetDialog.vue'),
    source('../src/composables/useSensitiveVisibility.ts'),
    source('../src/services/phone-session.ts'),
    source('../src/stores/settings.ts')
  ])

  assert.doesNotMatch(app, /disclaimer_agreed_at|shouldShowDisclaimer/)
  assert.doesNotMatch(app, /getDisclaimerStatus/)
  assert.match(app, /disclaimerAccepted\.value = true/)
  assert.match(systemService, /put<DisclaimerStatus>\('\/settings\/disclaimer'/)
  assert.match(websheet, /hideck-websheet-callback/)
  assert.match(websheet, /hideck-websheet-complete/)
  assert.match(websheet, /hideck-websheet/)
  assert.match(sensitive, /hideck_show_sensitive/)
  assert.match(phoneSession, /hideck_phone_control/)
  assert.match(settingsStore, /x-hideck-signature/)
  assert.doesNotMatch([app, systemService, websheet, sensitive, phoneSession, settingsStore].join('\n'), /vohive/i)
})
