// Guards every route. Unauthenticated users are sent to /login; authenticated
// users hitting /login are sent home. Auth state is only reliable on the
// client (token lives in localStorage), so we skip the guard during SSR.
export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) return

  const auth = useAuthStore()
  const isAuthRoute = to.path === '/login'

  if (!auth.isAuthed && !isAuthRoute) {
    return navigateTo('/login')
  }
  if (auth.isAuthed && isAuthRoute) {
    return navigateTo('/')
  }
})
