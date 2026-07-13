import { defineStore } from 'pinia'

export interface Attributes {
  str: number; vit: number; int: number; wis: number; faith: number
}
export interface User {
  id: string
  name: string
  email: string
  title: string
  level: number
  exp: number
  nextExp: number
  coins: number
  streak: number
  streakFreezes: number
  onboarded: boolean
  locale: string
  avatarUrl: string
  timezone: string
  attributes: Attributes
  createdAt: string
}

const TOKEN_KEY = 'dg_token'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: '' as string,
    user: null as User | null,
    ready: false,
  }),
  getters: {
    isAuthed: (s) => !!s.token && !!s.user,
    firstName: (s) => s.user?.name.split(' ')[0] ?? '',
    initial: (s) => (s.user?.name.trim()[0] ?? '?').toUpperCase(),
  },
  actions: {
    /** Restore token from localStorage and fetch the profile (client only). */
    async init() {
      if (this.ready) return
      if (import.meta.client) {
        const saved = localStorage.getItem(TOKEN_KEY)
        if (saved) {
          this.token = saved
          try {
            await this.fetchMe()
          } catch {
            this.logout()
          }
        }
      }
      this.ready = true
    },
    setToken(token: string) {
      this.token = token
      if (import.meta.client) localStorage.setItem(TOKEN_KEY, token)
    },
    setUser(user: User) {
      this.user = user
    },
    async login(email: string, password: string) {
      const res = await useApi()<{ token: string; user: User }>('/auth/login', {
        method: 'POST',
        body: { email, password },
      })
      this.setToken(res.token)
      this.setUser(res.user)
    },
    async register(name: string, email: string, password: string) {
      const res = await useApi()<{ token: string; user: User }>('/auth/register', {
        method: 'POST',
        body: { name, email, password },
      })
      this.setToken(res.token)
      this.setUser(res.user)
    },
    async fetchMe() {
      const user = await useApi()<User>('/me')
      this.setUser(user)
    },
    async buyFreeze() {
      const user = await useApi()<User>('/streak/freeze', { method: 'POST' })
      this.setUser(user)
    },
    async updateProfile(payload: { name: string; title: string }) {
      const user = await useApi()<User>('/me', { method: 'PUT', body: payload })
      this.setUser(user)
    },
    async changePassword(currentPassword: string, newPassword: string) {
      await useApi()('/me/password', { method: 'PUT', body: { currentPassword, newPassword } })
    },
    async deleteAccount(password: string) {
      await useApi()('/me', { method: 'DELETE', body: { password } })
      this.logout()
    },
    async completeOnboarding() {
      const user = await useApi()<User>('/me/onboard', { method: 'POST' })
      this.setUser(user)
    },
    /** Upload an avatar image directly to Cloudinary (signed), then save its URL. */
    async uploadAvatar(file: File) {
      const url = await useUpload().uploadImage(file, 'avatar')
      const user = await useApi()<User>('/me/avatar', { method: 'PUT', body: { url } })
      this.setUser(user)
    },
    async removeAvatar() {
      const user = await useApi()<User>('/me/avatar', { method: 'DELETE' })
      this.setUser(user)
    },
    /** Sync the browser's timezone to the account if it differs from what's saved. */
    async syncTimezone() {
      if (!import.meta.client || !this.user) return
      const tz = Intl.DateTimeFormat().resolvedOptions().timeZone
      if (!tz || tz === this.user.timezone) return
      try {
        await useApi()('/me/timezone', { method: 'PUT', body: { timezone: tz } })
        this.user.timezone = tz
      } catch { /* invalid/unsupported tz — keep saved value */ }
    },
    logout() {
      this.token = ''
      this.user = null
      if (import.meta.client) localStorage.removeItem(TOKEN_KEY)
    },
  },
})
