<script setup lang="ts">
import { idNum, ATTR_COLOR, ATTR_ICON } from '~/utils/format'

interface Achievement {
  id: string; name: string; icon: string; color: string
  hint: string; unlocked: boolean; progress: number; target: number
}

const auth = useAuthStore()
const { show } = useToast()
const { t, locale } = useI18n()

const joined = computed(() => {
  if (!auth.user) return ''
  return new Date(auth.user.createdAt).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'id-ID', { month: 'short', year: 'numeric' })
})
const expPct = computed(() => {
  const u = auth.user
  return u && u.nextExp ? Math.min(100, Math.round((u.exp / u.nextExp) * 100)) : 0
})
// Attribute bar fill is scaled against a soft cap of 30 for display.
const bars = computed(() => {
  const a = auth.user?.attributes
  return [
    { key: 'STR', id: 'str', val: a?.str ?? 0 },
    { key: 'VIT', id: 'vit', val: a?.vit ?? 0 },
    { key: 'INT', id: 'int', val: a?.int ?? 0 },
    { key: 'WIS', id: 'wis', val: a?.wis ?? 0 },
    { key: 'FAITH', id: 'faith', val: a?.faith ?? 0 },
  ]
})

// --- achievements (live) ---
const achievements = ref<Achievement[]>([])
const unlockedCount = computed(() => achievements.value.filter((a) => a.unlocked).length)

// --- profile editing ---
const editing = ref(false)
const editName = ref('')
const editTitle = ref('')
const savingProfile = ref(false)
const editMsg = ref('')

const FREEZE_COST = 100
const buying = ref(false)
const freezeMsg = ref('')

// avatar upload
const fileInput = ref<HTMLInputElement | null>(null)
const uploadingAvatar = ref(false)
async function onPickAvatar(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  uploadingAvatar.value = true
  try {
    await auth.uploadAvatar(file)
    show(t('profile.updated'), 'ph-fill ph-image')
  } catch (err: any) {
    show(err?.data?.error ?? t('auth.genericErr'), 'ph-fill ph-warning')
  } finally {
    uploadingAvatar.value = false
    if (fileInput.value) fileInput.value.value = ''
  }
}

async function loadAchievements() {
  achievements.value = await useApi()<Achievement[]>('/achievements')
}

onMounted(async () => {
  await auth.fetchMe()
  await loadAchievements()
})

function openEdit() {
  editName.value = auth.user?.name ?? ''
  editTitle.value = auth.user?.title ?? ''
  editMsg.value = ''
  editing.value = true
}

async function saveProfile() {
  editMsg.value = ''
  if (!editName.value.trim()) {
    editMsg.value = t('profile.nameEmpty')
    return
  }
  savingProfile.value = true
  try {
    await auth.updateProfile({ name: editName.value.trim(), title: editTitle.value.trim() })
    editing.value = false
    show(t('profile.updated'), 'ph-fill ph-user-circle')
  } catch (e: any) {
    editMsg.value = e?.data?.error ?? t('auth.genericErr')
  } finally {
    savingProfile.value = false
  }
}

async function buyFreeze() {
  freezeMsg.value = ''
  buying.value = true
  try {
    await auth.buyFreeze()
    show(t('profile.freezeBought'), 'ph-fill ph-snowflake')
  } catch (e: any) {
    freezeMsg.value = e?.data?.error ?? t('auth.genericErr')
  } finally {
    buying.value = false
  }
}

async function logout() {
  auth.logout()
  await navigateTo('/login')
}
</script>

<template>
  <div v-if="auth.user" class="fx col gap16">
    <div class="fx ac jb">
      <span class="h">{{ t('profile.title') }}</span>
      <div class="fx gap12 ac" style="color:var(--muted);font-size:20px">
        <NuxtLink to="/analytics" class="pact" style="padding:0"><i class="ph ph-chart-line-up" style="font-size:20px" /></NuxtLink>
        <button class="pact" style="padding:0" @click="openEdit"><i class="ph ph-pencil-simple" style="font-size:20px" /></button>
        <NuxtLink to="/settings" class="pact" style="padding:0"><i class="ph ph-gear-six" style="font-size:20px" /></NuxtLink>
      </div>
    </div>

    <!-- edit form -->
    <div v-if="editing" class="card pad fx col gap12">
      <span class="sec">{{ t('profile.editTitle') }}</span>
      <div>
        <span class="flabel">{{ t('auth.name') }}</span>
        <div class="field"><i class="ph ph-user" /><input v-model="editName" type="text" :placeholder="t('auth.namePh')"></div>
      </div>
      <div>
        <span class="flabel">{{ t('profile.gelar') }}</span>
        <div class="field"><i class="ph ph-seal" /><input v-model="editTitle" type="text" :placeholder="t('profile.gelarPh')"></div>
      </div>
      <p v-if="editMsg" style="color:var(--str);font:600 11.5px 'Plus Jakarta Sans';margin:0">{{ editMsg }}</p>
      <div class="fx gap10">
        <button class="btn btno" style="flex:1" @click="editing = false">{{ t('common.cancel') }}</button>
        <button class="btn" style="flex:1" :disabled="savingProfile" @click="saveProfile">
          <i class="ph-bold ph-check" />{{ savingProfile ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </div>

    <div class="fx col ac" style="text-align:center">
      <div class="ring" style="position:relative" :style="{ background: `conic-gradient(var(--primary) ${expPct}%, var(--track) 0)` }">
        <div
          class="avatar" style="width:82px;height:82px;font-size:26px"
          :style="auth.user.avatarUrl ? { backgroundImage: `url(${auth.user.avatarUrl})` } : {}"
        >
          <template v-if="!auth.user.avatarUrl">{{ auth.initial }}</template>
        </div>
        <span class="lvlbadge">LEVEL {{ auth.user.level }}</span>
        <button
          type="button" aria-label="Upload avatar" :disabled="uploadingAvatar"
          style="position:absolute;top:0;right:0;width:30px;height:30px;border-radius:50%;background:var(--primary);color:#fff;border:2px solid var(--surface);display:flex;align-items:center;justify-content:center;font-size:14px"
          @click="fileInput?.click()"
        >
          <i :class="uploadingAvatar ? 'ph ph-spinner' : 'ph-fill ph-camera'" />
        </button>
        <input ref="fileInput" type="file" accept="image/*" style="display:none" @change="onPickAvatar">
      </div>
      <div class="h" style="margin-top:14px">{{ auth.user.name }}</div>
      <div class="mut" style="font-size:12.5px;margin-top:3px">{{ auth.user.title }} · {{ t('profile.joined') }} {{ joined }}</div>
      <div class="fx gap8 wrap jc" style="margin-top:12px">
        <span class="chip chipa"><i class="ph-fill ph-fire" />{{ auth.user.streak }} {{ t('profile.days') }}</span>
        <span class="chip chipa"><i class="ph-fill ph-coins" />{{ idNum(auth.user.coins) }}</span>
      </div>
    </div>

    <div class="card pad">
      <div class="fx ac jb" style="margin-bottom:6px">
        <span class="tny">EXP</span>
        <span style="font:700 11px 'Space Grotesk';color:var(--primary)">{{ auth.user.exp }} / {{ auth.user.nextExp }}</span>
      </div>
      <div class="xp"><div class="xpf" :style="{ width: expPct + '%' }" /></div>
    </div>

    <!-- streak + freeze -->
    <div class="card pad">
      <div class="fx ac jb">
        <div class="fx ac gap12">
          <div class="si" style="width:40px;height:40px;border-radius:13px;background:rgba(242,166,59,.16);color:var(--amber-ink);display:flex;align-items:center;justify-content:center;font-size:20px">
            <i class="ph-fill ph-fire" />
          </div>
          <div>
            <div style="font:700 15px 'Space Grotesk';color:var(--ink)">{{ t('profile.streakDays', { n: auth.user.streak }) }}</div>
            <div class="mut" style="font-size:11px">
              <i class="ph-fill ph-snowflake" style="color:var(--int)" />
              {{ t('profile.freezes', { n: auth.user.streakFreezes }) }}
            </div>
          </div>
        </div>
        <button class="chip chipa" style="border:none" :disabled="buying || auth.user.coins < FREEZE_COST" @click="buyFreeze">
          <i class="ph-fill ph-plus" />{{ t('profile.buy') }} · {{ FREEZE_COST }}
        </button>
      </div>
      <p class="mut" style="font:500 10.5px/1.5 'Plus Jakarta Sans';margin:10px 0 0">
        {{ t('profile.freezeHint') }}
      </p>
      <p v-if="freezeMsg" style="color:var(--str);font:600 11px 'Plus Jakarta Sans';margin:6px 0 0">{{ freezeMsg }}</p>
    </div>

    <div class="card pad">
      <span class="sec" style="display:block;margin-bottom:14px">{{ t('profile.attributes') }}</span>
      <div class="fx col gap12">
        <div v-for="b in bars" :key="b.key" class="statbar">
          <div class="si" :style="{ background: `color-mix(in srgb, ${ATTR_COLOR[b.id]} 14%, transparent)`, color: ATTR_COLOR[b.id] }">
            <i :class="ATTR_ICON[b.id]" />
          </div>
          <span class="sk">{{ b.key }}</span>
          <div class="sbtrack"><div class="sbfill" :style="{ width: Math.min(100, (b.val / 30) * 100) + '%', background: ATTR_COLOR[b.id] }" /></div>
          <span class="sv2">{{ b.val }}</span>
        </div>
      </div>
    </div>

    <!-- achievements (live) -->
    <div class="card pad">
      <div class="fx ac jb" style="margin-bottom:14px">
        <span class="sec">{{ t('profile.achievements') }}</span>
        <span class="tny">{{ unlockedCount }} / {{ achievements.length }} {{ t('profile.unlocked') }}</span>
      </div>
      <div class="badgegrid">
        <div v-for="a in achievements" :key="a.id" class="badge" :title="a.hint">
          <div
            class="bicon" :class="{ blocked: !a.unlocked }"
            :style="a.unlocked ? { background: `color-mix(in srgb, ${a.color} 16%, transparent)`, color: a.color } : {}"
          >
            <i :class="a.unlocked ? a.icon : 'ph-fill ph-lock-simple'" />
          </div>
          <span class="bname">{{ a.name }}</span>
          <span v-if="!a.unlocked" class="bname" style="font-size:8px;opacity:.8">{{ Math.min(a.progress, a.target) }}/{{ a.target }}</span>
        </div>
      </div>
    </div>
  </div>

  <div v-else class="spinner" />
</template>
