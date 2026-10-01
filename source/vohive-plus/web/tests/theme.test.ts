import assert from 'node:assert/strict'
import test from 'node:test'
import { resolveStoredTheme, nextNavyTheme, themeClassNames, applyThemeClass, persistTheme, readStoredTheme } from '../src/utils/theme'
test('original light/dark preferences survive migration', () => {
 for (const [value, expected] of [['light','light'],['dark','dark'],['navy-light','light'],['navy-night','dark'],['classic','dark'],['unknown','light']] as const) assert.equal(resolveStoredTheme(value),expected)
 assert.equal(resolveStoredTheme(null),'light')
})
test('theme button toggles the original pair', () => { assert.equal(nextNavyTheme('light'),'dark'); assert.equal(nextNavyTheme('dark'),'light') })
test('theme restores dark class and clears obsolete classic class', () => {
 const classes=new Set(['classic']); const root={ classList:{toggle(name:string, force?:boolean){ if(force) classes.add(name);else classes.delete(name);return !!force}} }
 applyThemeClass('dark', root as any);assert.deepEqual([...classes],['dark']);applyThemeClass('light',root as any);assert.deepEqual([...classes],[])
 assert.deepEqual(themeClassNames('light'),{dark:false,classic:false})
})
test('persistence and unavailable storage preserve an original theme', () => {
 const values=new Map<string,string>();persistTheme('dark',{setItem:(k,v)=>values.set(k,v)});assert.equal(readStoredTheme({getItem:k=>values.get(k)||null}),'dark')
 assert.equal(readStoredTheme({getItem:()=>{throw new Error('unavailable')}}),'light')
})
