import { messages, type Locale } from '~/i18n/messages'

const locale = ref<Locale>('id')

function resolve(dict: any, key: string): any {
  return key.split('.').reduce((o, k) => (o != null ? o[k] : undefined), dict)
}

/** Minimal reactive i18n: `t(key)` + per-user persisted locale. */
export function useI18n() {
  function t(key: string, params?: Record<string, string | number>): string {
    let val = resolve(messages[locale.value], key)
    if (val == null) val = resolve(messages.id, key) // fall back to Indonesian
    if (typeof val !== 'string') return key
    if (params) {
      for (const [k, v] of Object.entries(params)) {
        val = val.replace(new RegExp(`\\{${k}\\}`, 'g'), String(v))
      }
    }
    return val
  }

  /** Set the active locale, persisting to localStorage and (if logged in) the API. */
  async function setLocale(l: Locale, persist = true) {
    locale.value = l
    if (import.meta.client) localStorage.setItem('dg_locale', l)
    if (persist) {
      const auth = useAuthStore()
      if (auth.isAuthed) {
        try { await useApi()('/me/locale', { method: 'PUT', body: { locale: l } }) } catch { /* non-fatal */ }
      }
    }
  }

  /** Initialise the locale. Pass a preferred value (e.g. the user's), else fall
   *  back to localStorage, then the browser language, then Indonesian. */
  function initLocale(preferred?: Locale) {
    if (preferred === 'id' || preferred === 'en') { locale.value = preferred; return }
    if (import.meta.client) {
      const saved = localStorage.getItem('dg_locale')
      if (saved === 'id' || saved === 'en') locale.value = saved
      else locale.value = navigator.language.startsWith('en') ? 'en' : 'id'
    }
  }

  return { locale, t, setLocale, initLocale }
}
