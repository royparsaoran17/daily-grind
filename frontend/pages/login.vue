<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const auth = useAuthStore()
const { t, locale, setLocale } = useI18n()
const mode = ref<'login' | 'register'>('login')
const showPw = ref(false)

const name = ref('')
const email = ref('nadia@email.com')
const password = ref('password123')
const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  loading.value = true
  try {
    if (mode.value === 'login') {
      await auth.login(email.value, password.value)
    } else {
      await auth.register(name.value, email.value, password.value)
    }
    // Adopt the account's saved language + sync device timezone.
    if (auth.user?.locale) setLocale(auth.user.locale as any, false)
    auth.syncTimezone()
    await navigateTo('/')
  } catch (e: any) {
    error.value = e?.data?.error ?? t('auth.genericErr')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div style="gap:0">
    <div class="fx jb ac" style="margin-bottom:26px">
      <span class="lang">
        <button class="langi" :class="{ langon: locale === 'en' }" @click="setLocale('en')">EN</button>
        <button class="langi" :class="{ langon: locale === 'id' }" @click="setLocale('id')">ID</button>
      </span>
      <ThemeToggle />
    </div>

    <div class="fx col ac" style="text-align:center;margin-bottom:28px">
      <div class="logo" style="margin-bottom:16px"><i class="ph-fill ph-lightning" /></div>
      <div class="h" style="font-size:24px">DailyGrind</div>
      <div class="mut" style="font-size:13px;margin-top:6px">{{ t('auth.tagline') }}</div>
    </div>

    <div class="seg" style="margin-bottom:20px">
      <button class="segi" :class="{ segon: mode === 'login' }" @click="mode = 'login'">{{ t('auth.signin') }}</button>
      <button class="segi" :class="{ segon: mode === 'register' }" @click="mode = 'register'">{{ t('auth.signup') }}</button>
    </div>

    <form @submit.prevent="submit">
      <div v-if="mode === 'register'" style="margin-bottom:14px">
        <span class="flabel">{{ t('auth.name') }}</span>
        <div class="field">
          <i class="ph ph-user" />
          <input v-model="name" type="text" :placeholder="t('auth.namePh')" required>
        </div>
      </div>

      <div style="margin-bottom:14px">
        <span class="flabel">{{ t('auth.email') }}</span>
        <div class="field">
          <i class="ph ph-envelope-simple" />
          <input v-model="email" type="email" placeholder="nama@email.com" required>
        </div>
      </div>

      <div style="margin-bottom:10px">
        <span class="flabel">{{ t('auth.password') }}</span>
        <div class="field">
          <i class="ph ph-lock-simple" />
          <input v-model="password" :type="showPw ? 'text' : 'password'" placeholder="••••••••" required>
          <i :class="showPw ? 'ph ph-eye-slash' : 'ph ph-eye'" style="cursor:pointer" @click="showPw = !showPw" />
        </div>
      </div>

      <div v-if="mode === 'login'" class="fx jc" style="margin-bottom:22px">
        <span class="mut" style="font:600 12px 'Plus Jakarta Sans'">{{ t('auth.forgot') }}</span>
      </div>
      <div v-else style="height:22px" />

      <p v-if="error" style="color:var(--str);font:600 12px 'Plus Jakarta Sans';margin:0 0 14px;text-align:center">
        {{ error }}
      </p>

      <button class="btn" type="submit" :disabled="loading" style="margin-bottom:16px">
        <template v-if="!loading">{{ mode === 'login' ? t('auth.signin') : t('auth.signup') }} <i class="ph-bold ph-arrow-right" /></template>
        <template v-else>{{ t('auth.processing') }}</template>
      </button>
    </form>

    <div class="divider"><span class="divline" />{{ t('auth.continueWith') }}<span class="divline" /></div>
    <button class="btn btno" style="margin-top:16px" @click="error = t('auth.googleSoon')">
      <i class="ph-bold ph-google-logo" />{{ t('auth.google') }}
    </button>

    <p class="mut" style="font:500 11px 'Plus Jakarta Sans';text-align:center;margin-top:20px">
      {{ t('auth.demoHint') }}
    </p>
  </div>
</template>
