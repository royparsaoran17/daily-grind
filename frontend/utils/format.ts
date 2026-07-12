/** Indonesian thousands separator: 1240 -> "1.240". */
export function idNum(n: number): string {
  return new Intl.NumberFormat('id-ID').format(n)
}

/** Relative time in Bahasa Indonesia: "2 jam lalu". */
export function timeAgo(iso: string): string {
  const then = new Date(iso).getTime()
  const diff = Math.max(0, Date.now() - then)
  const m = Math.floor(diff / 60000)
  if (m < 1) return 'baru saja'
  if (m < 60) return `${m} menit lalu`
  const h = Math.floor(m / 60)
  if (h < 24) return `${h} jam lalu`
  const d = Math.floor(h / 24)
  return `${d} hari lalu`
}

/** Attribute display colour token. */
export const ATTR_COLOR: Record<string, string> = {
  str: 'var(--str)',
  vit: 'var(--vit)',
  int: 'var(--int)',
  wis: 'var(--wis)',
  faith: 'var(--faith)',
}

export const ATTR_ICON: Record<string, string> = {
  str: 'ph-fill ph-barbell',
  vit: 'ph-fill ph-heartbeat',
  int: 'ph-fill ph-brain',
  wis: 'ph-fill ph-book-open',
  faith: 'ph-fill ph-hands-praying',
}
