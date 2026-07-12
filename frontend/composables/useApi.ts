import type { $Fetch } from 'nitropack'

const toSnake = (s: string) => s.replace(/[A-Z]/g, (m) => '_' + m.toLowerCase())
const toCamel = (s: string) => s.replace(/_([a-z0-9])/g, (_, c) => c.toUpperCase())

/** Deep-convert object keys with `fn`. Arrays/primitives pass through. */
function convertKeys(input: any, fn: (k: string) => string): any {
  if (Array.isArray(input)) return input.map((v) => convertKeys(v, fn))
  if (input && typeof input === 'object' && !(input instanceof Date)) {
    const out: Record<string, any> = {}
    for (const [k, v] of Object.entries(input)) out[fn(k)] = convertKeys(v, fn)
    return out
  }
  return input
}

/**
 * Preconfigured $fetch for the Go API. Attaches the bearer token and bridges
 * naming conventions: the wire uses snake_case (matching the database), while
 * Vue code keeps camelCase — request bodies are snake_cased on the way out and
 * responses are camelCased on the way in.
 */
export function useApi(): $Fetch {
  const config = useRuntimeConfig()
  const auth = useAuthStore()

  return $fetch.create({
    baseURL: config.public.apiBase,
    onRequest({ options }) {
      if (auth.token) {
        options.headers = new Headers(options.headers)
        options.headers.set('Authorization', `Bearer ${auth.token}`)
      }
      if (options.body && typeof options.body === 'object') {
        options.body = convertKeys(options.body, toSnake)
      }
    },
    onResponse({ response }) {
      if (response._data && typeof response._data === 'object') {
        response._data = convertKeys(response._data, toCamel)
      }
    },
    onResponseError({ response }) {
      if (response.status === 401 && auth.token) {
        auth.logout()
        if (import.meta.client) navigateTo('/login')
      }
    },
  })
}
