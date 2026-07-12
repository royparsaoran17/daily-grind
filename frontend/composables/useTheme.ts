const isDark = ref(false)

/** Global light/dark theme with localStorage persistence. */
export function useTheme() {
  function apply() {
    if (import.meta.client) {
      document.body.classList.toggle('dark', isDark.value)
    }
  }
  function toggle() {
    isDark.value = !isDark.value
    if (import.meta.client) localStorage.setItem('dg_theme', isDark.value ? 'dark' : 'light')
    apply()
  }
  function init() {
    if (import.meta.client) {
      const saved = localStorage.getItem('dg_theme')
      isDark.value = saved
        ? saved === 'dark'
        : window.matchMedia('(prefers-color-scheme: dark)').matches
      apply()
    }
  }
  return { isDark, toggle, init, apply }
}
