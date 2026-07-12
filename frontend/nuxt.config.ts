// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-07-01',
  devtools: { enabled: false },
  ssr: true,
  modules: ['@pinia/nuxt'],
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    public: {
      // Override with NUXT_PUBLIC_API_BASE at runtime.
      apiBase: 'http://localhost:8080/api',
    },
  },
  app: {
    head: {
      title: 'DailyGrind',
      viewport: 'width=device-width, initial-scale=1, maximum-scale=1, viewport-fit=cover',
      meta: [
        { name: 'theme-color', content: '#8c2f3a' },
        { name: 'mobile-web-app-capable', content: 'yes' },
      ],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@500;600;700&family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap',
        },
        { rel: 'stylesheet', href: 'https://unpkg.com/@phosphor-icons/web@2.1.1/src/regular/style.css' },
        { rel: 'stylesheet', href: 'https://unpkg.com/@phosphor-icons/web@2.1.1/src/fill/style.css' },
        { rel: 'stylesheet', href: 'https://unpkg.com/@phosphor-icons/web@2.1.1/src/bold/style.css' },
      ],
    },
  },
})
