export function duration(sec: number): string {
  if (!sec) return ''
  const s = Math.round(sec)
  const h = Math.floor(s / 3600)
  const m = Math.floor((s % 3600) / 60)
  const ss = String(s % 60).padStart(2, '0')
  return h ? `${h}:${String(m).padStart(2, '0')}:${ss}` : `${m}:${ss}`
}

export function bytes(n: number): string {
  if (!n) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const i = Math.min(Math.floor(Math.log(n) / Math.log(1024)), units.length - 1)
  return `${(n / 1024 ** i).toFixed(i ? 1 : 0)} ${units[i]}`
}

export function eta(sec: number): string {
  return sec >= 0 ? duration(sec) || '0:00' : '—'
}

// "47,2 %" style, Spanish decimals.
export function pct(n: number): string {
  return `${n.toLocaleString('es', { minimumFractionDigits: 1, maximumFractionDigits: 1 })} %`
}

export function modeLabel(r: { mode: string; audioFormat: string; audioBitrate: number; height: number; videoFormat: string }): string {
  return r.mode === 'audio' ? `${r.audioFormat.toUpperCase()} ${r.audioBitrate} kbps` : `${r.videoFormat.toUpperCase()} ${r.height ? r.height + 'p' : 'máx.'}`
}

// "Hoy", "Ayer" or "martes 21 de octubre".
export function dayLabel(iso: string): string {
  const d = new Date(iso)
  const day = (x: Date) => new Date(x.getFullYear(), x.getMonth(), x.getDate()).getTime()
  const diff = Math.round((day(new Date()) - day(d)) / 86400000)
  if (diff === 0) return 'Hoy'
  if (diff === 1) return 'Ayer'
  return d.toLocaleDateString('es', { weekday: 'long', day: 'numeric', month: 'long' })
}

// "74 mil vistas", "8,2 M de vistas"
export function views(n: number | undefined): string {
  if (!n) return ''
  if (n >= 1_000_000) return `${(n / 1_000_000).toLocaleString('es', { maximumFractionDigits: 1 })} M de vistas`
  if (n >= 10_000) return `${Math.round(n / 1000).toLocaleString('es')} mil vistas`
  return `${n.toLocaleString('es')} vistas`
}
