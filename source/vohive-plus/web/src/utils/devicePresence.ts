import { ref } from 'vue'
import type { AxiosInstance } from 'axios'

export interface PresenceDevice {
  id?: string
  running?: boolean
  worker_running?: boolean
  lifecycle_phase?: string
  backend_mode?: string
  usb_path?: string
  control_device?: string
  interface?: string
  at_port?: string
  modem?: { imei?: string }
  config?: { modem_imei?: string; device_backend?: string }
}

export interface DiscoveredHardware {
  configured_id?: string
  imei?: string
  usb_path?: string
  control_path?: string
  net_interface?: string
  at_port?: string
}

function sameHardware(device: PresenceDevice, hardware: DiscoveredHardware) {
  if (hardware.configured_id && hardware.configured_id === device.id) return true
  const imei = device.modem?.imei || device.config?.modem_imei
  if (imei && hardware.imei && imei === hardware.imei) return true
  return [
    [device.usb_path, hardware.usb_path],
    [device.control_device, hardware.control_path],
    [device.interface, hardware.net_interface],
    [device.at_port, hardware.at_port],
  ].some(([a, b]) => typeof a === 'string' && a !== '' && a === b)
}

// Hardware discovery is read-only. Unknown/stale results must not label a
// modem as unplugged while it is initializing or recovering from a reset.
export function createDevicePresenceTracker(now = Date.now) {
  const epoch = ref(0)
  let hardware: DiscoveredHardware[] | null = null
  let hardwareAt = 0
  let pcscError = false
  const managed = new Map<string, PresenceDevice>()
  const lastRunning = new Map<string, number>()
  const recoveryStarted = new Map<string, number>()

  function remember(devices: PresenceDevice[]) {
    for (const device of devices) {
      if (!device?.id) continue
      managed.set(device.id, { ...managed.get(device.id), ...device })
      if (device.running || device.worker_running) {
        lastRunning.set(device.id, now())
        recoveryStarted.delete(device.id)
      } else if (['rebooting', 'usb_wait'].includes(device.lifecycle_phase || '')) {
        if (!recoveryStarted.has(device.id)) recoveryStarted.set(device.id, now())
      } else recoveryStarted.delete(device.id)
    }
    epoch.value++
  }

  function updateHardware(devices: DiscoveredHardware[] | null, error = false) {
    hardware = devices
    hardwareAt = now()
    pcscError = error
    epoch.value++
  }

  function isAbsent(device: PresenceDevice | null | undefined) {
    void epoch.value
    if (!device?.id || device.running || device.worker_running || !hardware) return false
    if (now() - hardwareAt > 45000) return false
    const lastSeen = lastRunning.get(device.id)
    if (lastSeen !== undefined && now() - lastSeen < 15000) return false
    const recoveryAt = recoveryStarted.get(device.id)
    if (recoveryAt !== undefined && now() - recoveryAt < 30000) return false
    const current = { ...managed.get(device.id), ...device }
    if (pcscError && (current.backend_mode === 'pcsc' || current.config?.device_backend === 'pcsc')) return false
    if (hardware.some(item => sameHardware(current, item))) return false
    const known = [...managed.values()]
    if (hardware.some(item => !known.some(entry =>
      (entry.running || entry.worker_running) && sameHardware(entry, item)))) return false
    return true
  }

  return { remember, updateHardware, isAbsent }
}

const tracker = createDevicePresenceTracker()
export const rememberDevicePresence = tracker.remember
export const isDeviceAbsent = tracker.isAbsent

export function installDevicePresenceTracking(api: AxiosInstance) {
  let pollAt = 0
  let pending = false
  async function refresh() {
    if (pending || Date.now() - pollAt < 15000 || document.hidden ||
      !api.defaults.headers.common.Authorization) return
    pending = true
    pollAt = Date.now()
    try {
      const response = await api.get('/devices/discovered', { params: { with_imei: 0 }, timeout: 5000 })
      if (!Array.isArray(response.data?.devices)) throw Error('硬件检测暂不可用')
      tracker.updateHardware(response.data.devices, !!response.data.pcsc_error)
    } catch {
      tracker.updateHardware(null)
    } finally {
      pending = false
    }
  }
  api.interceptors.response.use(response => {
    const path = String(response.config?.url || '').split('?')[0]
    if ((path === '/devices' || /^\/devices\/[^/]+\/overview$/.test(path)) &&
      Array.isArray(response.data?.devices)) {
      tracker.remember(response.data.devices)
      void refresh()
    }
    return response
  })
  window.setInterval(() => { void refresh() }, 15000)
  document.addEventListener('visibilitychange', () => {
    if (!document.hidden) { pollAt = 0; void refresh() }
  })
}
