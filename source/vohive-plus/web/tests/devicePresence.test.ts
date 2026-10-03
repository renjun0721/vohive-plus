import assert from 'node:assert/strict'
import test from 'node:test'
import { createDevicePresenceTracker, type PresenceDevice } from '../src/utils/devicePresence'

test('two connected modems remain online and two missing configurations are absent', () => {
  const tracker = createDevicePresenceTracker(() => 100000)
  const devices = [
    { id: 'one', running: true, interface: 'wwan0' },
    { id: 'two', running: true, control_device: '/dev/cdc-wdm1' },
    { id: 'three', running: false, lifecycle_phase: 'worker_starting' },
    { id: 'four', running: false, lifecycle_phase: 'qmi_starting' },
  ]
  tracker.remember(devices)
  tracker.updateHardware([{ net_interface: 'wwan0' }, { control_path: '/dev/cdc-wdm1' }])
  assert.deepEqual(devices.map(tracker.isAbsent), [false, false, true, true])
  tracker.updateHardware([{ net_interface: 'wwan0' }, { control_path: '/dev/cdc-wdm1' }, { usb_path: 'new-port' }])
  assert.equal(tracker.isAbsent(devices[2]), false, 'new hardware may be initializing')
})

test('recently online and restarting modems get grace periods and reinsertions clear absence', () => {
  let now = 100000
  const tracker = createDevicePresenceTracker(() => now)
  const device: PresenceDevice = { id: 'one', running: true }
  tracker.remember([device])
  device.running = false
  tracker.remember([device])
  tracker.updateHardware([])
  assert.equal(tracker.isAbsent(device), false)
  now += 15001
  assert.equal(tracker.isAbsent(device), true)
  device.lifecycle_phase = 'usb_wait'
  tracker.remember([device])
  assert.equal(tracker.isAbsent(device), false)
  now += 30001
  tracker.updateHardware([])
  assert.equal(tracker.isAbsent(device), true)
  tracker.updateHardware([{ configured_id: device.id }])
  assert.equal(tracker.isAbsent(device), false)
  device.running = true
  tracker.remember([device])
  assert.equal(tracker.isAbsent(device), false)
})

test('failed or stale discovery is unknown, and PCSC failure does not affect USB devices', () => {
  let now = 100000
  const tracker = createDevicePresenceTracker(() => now)
  const usb = { id: 'usb', running: false }
  const pcsc = { id: 'pcsc', running: false, config: { device_backend: 'pcsc' } }
  tracker.remember([usb, pcsc])
  tracker.updateHardware([], true)
  assert.equal(tracker.isAbsent(usb), true)
  assert.equal(tracker.isAbsent(pcsc), false)
  now += 45001
  assert.equal(tracker.isAbsent(usb), false)
  tracker.updateHardware(null)
  assert.equal(tracker.isAbsent(usb), false)
})
