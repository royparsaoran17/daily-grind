// Guards every route. Auth state is only reliable on the client (token lives in
// localStorage), so the guard is a no-op during SSR.
export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) return

  const auth = useAuthStore()
  const isAuthRoute = to.path === '/login'
  const isOnboarding = to.path === '/onboarding'

  if (!auth.isAuthed) {
    return isAuthRoute ? undefined : navigateTo('/login')
  }
  // Authenticated:
  if (isAuthRoute) return navigateTo('/')
  // New users must finish onboarding first.
  if (!auth.user?.onboarded && !isOnboarding) return navigateTo('/onboarding')
  if (auth.user?.onboarded && isOnboarding) return navigateTo('/')
})
