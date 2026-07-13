// Runs once on client boot: restore theme, locale, and session before render.
export default defineNuxtPlugin(async () => {
  const { init: initTheme } = useTheme()
  initTheme()

  const { initLocale } = useI18n()
  initLocale() // instant: localStorage / browser language

  const auth = useAuthStore()
  await auth.init()

  // Prefer the logged-in user's saved language.
  if (auth.user?.locale) initLocale(auth.user.locale as any)

  // Keep the account's timezone in sync with this device (for correct streaks).
  if (auth.isAuthed) auth.syncTimezone()
})
