<script setup lang="ts">
const auth = useAuthStore()
const { isDark, toggle: toggleTheme } = useTheme()
const { t, locale, setLocale } = useI18n()
const { show } = useToast()

const pwOpen = ref(false)
const curPw = ref('')
const newPw = ref('')
const pwSaving = ref(false)
const pwMsg = ref('')

const delOpen = ref(false)
const delPw = ref('')
const deleting = ref(false)
const delMsg = ref('')

async function changePassword() {
  pwMsg.value = ''
  if (newPw.value.length < 6) { pwMsg.value = t('settings.pwShort'); return }
  pwSaving.value = true
  try {
    await auth.changePassword(curPw.value, newPw.value)
    curPw.value = ''; newPw.value = ''; pwOpen.value = false
    show(t('settings.pwUpdated'), 'ph-fill ph-lock-key')
  } catch (e: any) {
    pwMsg.value = e?.data?.error ?? t('auth.genericErr')
  } finally {
    pwSaving.value = false
  }
}

async function deleteAccount() {
  delMsg.value = ''
  deleting.value = true
  try {
    await auth.deleteAccount(delPw.value)
    await navigateTo('/login')
  } catch (e: any) {
    delMsg.value = e?.data?.error ?? t('auth.genericErr')
  } finally {
    deleting.value = false
  }
}

async function logout() {
  auth.logout()
  await navigateTo('/login')
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac gap12">
      <button class="backbtn" @click="navigateTo('/profile')"><i class="ph-bold ph-caret-left" /></button>
      <span class="h">{{ t('settings.title') }}</span>
    </div>

    <!-- account -->
    <span class="sec">{{ t('settings.account') }}</span>
    <div class="card" style="overflow:hidden">
      <NuxtLink to="/profile" class="fx ac jb pad" style="border-bottom:1px solid var(--line)">
        <span class="fx ac gap12">
          <i class="ph ph-user-circle" style="font-size:20px;color:var(--primary)" />
          <span style="font:600 13px 'Plus Jakarta Sans';color:var(--ink)">{{ t('settings.editProfile') }}</span>
        </span>
        <i class="ph ph-caret-right mut" />
      </NuxtLink>
      <button class="fx ac jb pad" style="width:100%;background:none;border:none;border-bottom:1px solid var(--line)" @click="pwOpen = !pwOpen">
        <span class="fx ac gap12">
          <i class="ph ph-lock-key" style="font-size:20px;color:var(--primary)" />
          <span style="font:600 13px 'Plus Jakarta Sans';color:var(--ink)">{{ t('settings.changePassword') }}</span>
        </span>
        <i :class="pwOpen ? 'ph ph-caret-up' : 'ph ph-caret-down'" class="mut" />
      </button>
      <div v-if="pwOpen" class="pad fx col gap10" style="border-bottom:1px solid var(--line);background:var(--soft)">
        <div class="field"><i class="ph ph-lock-simple" /><input v-model="curPw" type="password" :placeholder="t('settings.curPw')"></div>
        <div class="field"><i class="ph ph-lock" /><input v-model="newPw" type="password" :placeholder="t('settings.newPw')"></div>
        <p v-if="pwMsg" style="color:var(--str);font:600 11px 'Plus Jakarta Sans';margin:0">{{ pwMsg }}</p>
        <button class="btn" :disabled="pwSaving" @click="changePassword">{{ pwSaving ? t('common.saving') : t('settings.savePw') }}</button>
      </div>
      <button class="fx ac jb pad" style="width:100%;background:none;border:none" @click="logout">
        <span class="fx ac gap12">
          <i class="ph ph-sign-out" style="font-size:20px;color:var(--primary)" />
          <span style="font:600 13px 'Plus Jakarta Sans';color:var(--ink)">{{ t('settings.logout') }}</span>
        </span>
        <i class="ph ph-caret-right mut" />
      </button>
    </div>

    <!-- preferences -->
    <span class="sec">{{ t('settings.preferences') }}</span>
    <div class="card">
      <!-- language -->
      <div class="fx ac jb pad" style="border-bottom:1px solid var(--line)">
        <span class="fx ac gap12">
          <i class="ph ph-translate" style="font-size:20px;color:var(--primary)" />
          <span style="font:600 13px 'Plus Jakarta Sans';color:var(--ink)">{{ t('settings.language') }}</span>
        </span>
        <span class="lang">
          <button class="langi" :class="{ langon: locale === 'id' }" @click="setLocale('id')">ID</button>
          <button class="langi" :class="{ langon: locale === 'en' }" @click="setLocale('en')">EN</button>
        </span>
      </div>
      <!-- dark mode -->
      <div class="fx ac jb pad">
        <span class="fx ac gap12">
          <i :class="isDark ? 'ph-fill ph-moon-stars' : 'ph-fill ph-sun'" style="font-size:20px;color:var(--primary)" />
          <span style="font:600 13px 'Plus Jakarta Sans';color:var(--ink)">{{ t('settings.darkMode') }}</span>
        </span>
        <button
          role="switch" :aria-checked="isDark" @click="toggleTheme"
          :style="{
            width: '44px', height: '26px', borderRadius: '20px', border: 'none', position: 'relative',
            background: isDark ? 'var(--primary)' : 'var(--track)', transition: 'background .2s', cursor: 'pointer',
          }"
        >
          <span :style="{
            position: 'absolute', top: '3px', left: isDark ? '21px' : '3px', width: '20px', height: '20px',
            borderRadius: '50%', background: '#fff', transition: 'left .2s',
          }" />
        </button>
      </div>
    </div>

    <!-- danger zone -->
    <span class="sec" style="color:var(--str)">{{ t('settings.danger') }}</span>
    <div class="card pad">
      <button class="btn btno" style="color:var(--str);border-color:var(--str)" @click="delOpen = !delOpen">
        <i class="ph ph-trash" />{{ t('settings.deleteAccount') }}
      </button>
      <div v-if="delOpen" class="fx col gap10" style="margin-top:12px">
        <p class="mut" style="font-size:11.5px;line-height:1.5;margin:0">{{ t('settings.deleteWarn') }}</p>
        <div class="field"><i class="ph ph-lock-simple" /><input v-model="delPw" type="password" :placeholder="t('settings.confirmPw')"></div>
        <p v-if="delMsg" style="color:var(--str);font:600 11px 'Plus Jakarta Sans';margin:0">{{ delMsg }}</p>
        <button class="btn" style="background:var(--str);box-shadow:none" :disabled="deleting || !delPw" @click="deleteAccount">
          {{ deleting ? t('common.saving') : t('settings.deleteYes') }}
        </button>
      </div>
    </div>

    <p class="mut" style="text-align:center;font-size:10.5px;margin:8px 0 0">DailyGrind v1.0</p>
  </div>
</template>
