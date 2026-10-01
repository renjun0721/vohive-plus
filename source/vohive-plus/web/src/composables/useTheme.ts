import { computed, ref } from 'vue'
import { applyThemeClass, persistTheme, readStoredTheme, type ThemeMode } from '../utils/theme'
const theme = ref<ThemeMode>(readStoredTheme())
applyThemeClass(theme.value)
export function useTheme() {
  const isDark = computed(() => theme.value === 'dark')
  function applyTheme(mode: ThemeMode) { theme.value = mode; persistTheme(mode); applyThemeClass(mode) }
  function toggleTheme() { applyTheme(isDark.value ? 'light' : 'dark') }
  return { theme, isDark, applyTheme, toggleTheme, isClassic: computed(() => true), applyClassic: () => applyTheme('dark'), restoreNavyTheme: (_mode?: string) => applyTheme('light') }
}
