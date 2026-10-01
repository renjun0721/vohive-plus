export const THEME_STORAGE_KEY = 'theme'
export const THEME_MODES = ['light', 'dark'] as const
export type ThemeMode = (typeof THEME_MODES)[number]
export type NavyThemeMode = ThemeMode
export type ThemeClassNames = { dark: boolean; classic: boolean }
export function isThemeMode(value: string | null | undefined): value is ThemeMode { return value === 'light' || value === 'dark' }
export function resolveStoredTheme(value: string | null | undefined): ThemeMode { return value === 'dark' || value === 'navy-night' || value === 'classic' ? 'dark' : 'light' }
export function readStoredTheme(storage: Pick<Storage, 'getItem'> | null | undefined = defaultStorage()): ThemeMode { try { return resolveStoredTheme(storage?.getItem(THEME_STORAGE_KEY)) } catch { return 'light' } }
export function persistTheme(mode: ThemeMode, storage: Pick<Storage, 'setItem'> | null | undefined = defaultStorage()): void { try { storage?.setItem(THEME_STORAGE_KEY, mode) } catch {} }
export function isDarkTheme(mode: ThemeMode): boolean { return mode === 'dark' }
export function themeClassNames(mode: ThemeMode): ThemeClassNames { return { dark: isDarkTheme(mode), classic: false } }
export function nextNavyTheme(mode: ThemeMode): ThemeMode { return mode === 'light' ? 'dark' : 'light' }
export function applyThemeClass(mode: ThemeMode, root: Pick<HTMLElement, 'classList'> | null | undefined = defaultRoot()): ThemeClassNames { const names = themeClassNames(mode); root?.classList.toggle('dark', names.dark); root?.classList.toggle('classic', false); return names }
function defaultStorage(): Storage | null { try { return typeof localStorage === 'undefined' ? null : localStorage } catch { return null } }
function defaultRoot(): HTMLElement | null { return typeof document === 'undefined' ? null : document.documentElement }
