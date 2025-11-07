export const { format: formatNumber } = Intl.NumberFormat('en-GB', {
  notation: 'compact',
  maximumFractionDigits: 1,
})

export function formatDate(d: Date) {
  const options = { year: 'numeric', month: 'long', day: 'numeric' }
  return new Date(d).toLocaleDateString('de', options)
}

export function formatTS(d: Date, options?: object) {
  const defaultoptions = { timezone: 'Europe/Berlin' }
  return new Date(d).toLocaleString('de', Object.assign(defaultoptions, options))
}

export function formatTime(d: Date) {
  const options = { year: 'numeric', month: 'long', day: 'numeric' }
  return new Date(d).toLocaleTimeString('de', {})
}

export function formatDuration(dur_in_seconds: number) {
  const minutes = Math.floor(dur_in_seconds / 60)
  const seconds = dur_in_seconds % 60
  const hours = Math.floor(dur_in_seconds / 3600)

  if (hours)
    return `${hours}h${minutes % 60}m${seconds}s`

  if (minutes)
    return `${minutes}m${seconds}s`

  return `${seconds}s`
}

export function formatBytes(bytes: number, decimals = 2) {
  if (!+bytes) return '0 Bytes'

  const k = 1024
  const dm = decimals < 0 ? 0 : decimals
  const sizes = ['Bytes', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB']

  const i = Math.floor(Math.log(bytes) / Math.log(k))

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
}
export const timeAgo = (timestamp: string | number | Date) => {
  const messages = {
    justNow: t('time.justNow'),
    past: (n: string | number | Date) =>
      n.toString().match(/\d/) ? t('time.past', { n }) : n,
    day: (n: Date | number | string) =>
      n === 1 ? t('time.yesterday') : t('time.day', { n }),
    hour: (n: string | number | Date) => t('time.hour', { n }),
    minute: (n: string | number | Date) => t('time.minute', { n }),
    second: (n: string | number | Date) => t('time.second', { n }),
  } as const
  const time = useTimeAgo(new Date(timestamp), {
    messages: messages as any,
  })
  return time.value.replace(/"/gi, '')
}

export function case_insensitive_comp(strA, strB) {
  return strA.toLowerCase().localeCompare(strB.toLowerCase())
}
