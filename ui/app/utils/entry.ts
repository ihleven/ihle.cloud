import { cloneDeep, tap, set } from 'lodash'

export function cloneSetPath(o: object, path: string, value: string | boolean | number | object | string[]): object {
  return tap(cloneDeep(o), e => set(e, path, value))
}

export function setPath(o: object, path: string, value: string | boolean | number | object): object {
  // return tap(o, e => set(e, path, value))
  return set(o, path, value)
}

export function entrylink(path: string): object {
  return { name: 'entries-slug', params: { slug: path.split('/').filter(s => s !== '') } }
}

export function basename(str: string): string {
  return str.substring(str.lastIndexOf('/') + 1)
}

export function dirname(path: string): string {
  if (!path) return ''
  return path.match(/.*\//) ? path.match(/.*\//)![0] : ''
}
