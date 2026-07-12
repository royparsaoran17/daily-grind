// Runs once on client boot: restore theme + session before the app renders.
export default defineNuxtPlugin(async () => {
  const { init: initTheme } = useTheme()
  initTheme()

  const auth = useAuthStore()
  await auth.init()
})
