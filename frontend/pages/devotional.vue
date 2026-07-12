<script setup lang="ts">
import type { User } from '~/stores/auth'

interface Devotional {
  id: string; date: string; title: string; passage: string; verseText: string
  reflection: string; prayer: string; faithReward: number; completed: boolean
}

const auth = useAuthStore()
const { show } = useToast()

const lang = ref<'EN' | 'ID'>('ID')
const devo = ref<Devotional | null>(null)
const pending = ref(true)
const missing = ref(false)
const saving = ref(false)

const dateLabel = computed(() =>
  devo.value ? new Date(devo.value.date).toLocaleDateString('id-ID', { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' }) : '',
)

onMounted(async () => {
  try {
    devo.value = await useApi()<Devotional>('/devotional/today')
  } catch {
    missing.value = true
  } finally {
    pending.value = false
  }
})

async function complete() {
  if (!devo.value || devo.value.completed) return
  saving.value = true
  try {
    const user = await useApi()<User>(`/devotional/${devo.value.id}/complete`, { method: 'POST' })
    auth.setUser(user)
    devo.value.completed = true
    show(`+${devo.value.faithReward} FAITH 🙏`, 'ph-fill ph-hands-praying')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac jb">
      <span class="h">Renungan Harian</span>
      <span v-if="auth.user" class="chip chipa"><i class="ph-fill ph-fire" />{{ auth.user.streak }} hari</span>
    </div>

    <div class="fx ac jb">
      <span class="chip"><i class="ph ph-calendar-blank" />{{ dateLabel || '—' }}</span>
      <span class="lang">
        <button class="langi" :class="{ langon: lang === 'EN' }" @click="lang = 'EN'">EN</button>
        <button class="langi" :class="{ langon: lang === 'ID' }" @click="lang = 'ID'">ID</button>
      </span>
    </div>

    <div v-if="pending" class="spinner" />
    <div v-else-if="missing" class="card pad mut" style="text-align:center;font-size:13px">
      Belum ada renungan untuk hari ini. Kembali lagi besok, ya.
    </div>

    <template v-else-if="devo">
      <div class="devohero">
        <div class="devolabel">RENUNGAN HARI INI</div>
        <div style="font:700 20px/1.25 'Space Grotesk';margin:8px 0 12px">{{ devo.title }}</div>
        <span class="chip" style="background:rgba(255,255,255,.18);color:#fff"><i class="ph-fill ph-book-open" />{{ devo.passage }}</span>
      </div>

      <div class="card pad" style="padding:18px">
        <p class="verse" style="margin:0;font-size:14px;font-style:italic">{{ devo.verseText }}</p>
      </div>

      <div>
        <span class="sec" style="display:block;margin-bottom:8px">Perenungan</span>
        <p style="font:400 13px/1.7 'Plus Jakarta Sans';color:var(--ink);margin:0">{{ devo.reflection }}</p>
      </div>

      <div class="card pad meaning">
        <div class="prayer">
          <div class="devolabel" style="color:var(--pink)">DOA HARI INI</div>
          <p style="font:400 12.5px/1.6 'Plus Jakarta Sans';color:var(--ink);margin:6px 0 0">{{ devo.prayer }}</p>
        </div>
      </div>

      <NuxtLink :to="`/journal?prompt=${encodeURIComponent(devo.passage)}`" class="card pad fx ac jb" style="text-decoration:none">
        <span class="fx ac gap12">
          <span class="qi" style="background:var(--psoft);color:var(--primary)"><i class="ph-fill ph-notebook" /></span>
          <span>
            <span style="display:block;font:700 13px 'Plus Jakarta Sans';color:var(--ink)">Tulis jurnal</span>
            <span class="mut" style="font-size:11px">Refleksikan renungan hari ini</span>
          </span>
        </span>
        <i class="ph ph-caret-right mut" style="font-size:18px" />
      </NuxtLink>

      <div class="fx ac gap10" style="margin-top:auto">
        <button class="btn f1" :disabled="devo.completed || saving" @click="complete">
          <i class="ph-bold ph-check" />
          <template v-if="devo.completed">Selesai hari ini</template>
          <template v-else>
            {{ saving ? 'Menyimpan…' : 'Selesai' }}
            <span class="pill" style="background:rgba(255,255,255,.2);color:#fff;padding:2px 7px"><i class="ph-fill ph-hands-praying" />+{{ devo.faithReward }} FAITH</span>
          </template>
        </button>
        <button class="btn btno" style="width:auto;padding:15px;flex:none"><i class="ph ph-bookmark-simple" /></button>
        <button class="btn btno" style="width:auto;padding:15px;flex:none"><i class="ph ph-share-network" /></button>
      </div>
    </template>
  </div>
</template>
