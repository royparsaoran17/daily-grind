<script setup lang="ts">
import { idNum, ATTR_COLOR, ATTR_ICON } from '~/utils/format'

const auth = useAuthStore()
const { quests, load, toggle, doneCount } = useQuests()

interface FeedPost {
  id: string; author: string; authorLevel: number; badge?: string
}
const activity = ref<FeedPost[]>([])
const pending = ref(true)

const attrs = computed(() => {
  const a = auth.user?.attributes
  return [
    { key: 'STR', val: a?.str ?? 0, id: 'str' },
    { key: 'VIT', val: a?.vit ?? 0, id: 'vit' },
    { key: 'INT', val: a?.int ?? 0, id: 'int' },
    { key: 'WIS', val: a?.wis ?? 0, id: 'wis' },
    { key: 'FAITH', val: a?.faith ?? 0, id: 'faith' },
  ]
})
const expPct = computed(() => {
  const u = auth.user
  if (!u || !u.nextExp) return 0
  return Math.min(100, Math.round((u.exp / u.nextExp) * 100))
})
const todayQuests = computed(() => quests.value.slice(0, 3))

onMounted(async () => {
  try {
    await Promise.all([
      auth.fetchMe(),
      load(true),
      useApi()<FeedPost[]>('/feed').then((f) => (activity.value = f.slice(0, 2))),
    ])
  } finally {
    pending.value = false
  }
})
</script>

<template>
  <div v-if="auth.user" class="fx col gap16">
    <!-- greeting -->
    <div class="fx ac jb">
      <div>
        <div class="mut" style="font-size:12.5px;font-weight:600">Selamat datang,</div>
        <div class="h">{{ auth.firstName }} ✦</div>
      </div>
      <div class="fx ac gap8">
        <span class="chip chipa"><i class="ph-fill ph-fire" />{{ auth.user.streak }}</span>
        <div class="avatar" style="width:42px;height:42px">{{ auth.initial }}</div>
      </div>
    </div>

    <!-- level / EXP card -->
    <div class="card pad" style="background:linear-gradient(135deg,var(--primary),var(--primary2));border:none;color:#fff;box-shadow:0 12px 26px -10px rgba(140,47,58,.55)">
      <div class="fx ac gap16">
        <div class="ring" :style="{ background: `conic-gradient(#fff ${expPct}%, rgba(255,255,255,.28) 0)` }">
          <div class="avatar" style="width:60px;height:60px;background:#efe9e0">{{ auth.initial }}</div>
          <span class="lvlbadge" style="background:var(--amber);border-color:transparent">LVL {{ auth.user.level }}</span>
        </div>
        <div class="f1">
          <div class="fx ac jb">
            <span style="font:700 16px 'Space Grotesk'">{{ auth.user.name }}</span>
            <span class="pill" style="background:rgba(255,255,255,.2);color:#fff"><i class="ph-fill ph-coins" />{{ idNum(auth.user.coins) }}</span>
          </div>
          <div style="font:600 11.5px 'Plus Jakarta Sans';color:rgba(255,255,255,.82);margin:2px 0 10px">{{ auth.user.title }}</div>
          <div class="xp" style="background:rgba(255,255,255,.25)"><div class="xpf" :style="{ width: expPct + '%', background: '#fff' }" /></div>
          <div style="font:600 10.5px 'Plus Jakarta Sans';color:rgba(255,255,255,.88);margin-top:5px">
            {{ auth.user.exp }} / {{ auth.user.nextExp }} EXP menuju Lvl {{ auth.user.level + 1 }}
          </div>
        </div>
      </div>
    </div>

    <!-- attributes -->
    <div class="fx" style="gap:2px">
      <div v-for="a in attrs" :key="a.key" class="statmini">
        <div class="si" :style="{ background: `color-mix(in srgb, ${ATTR_COLOR[a.id]} 14%, transparent)`, color: ATTR_COLOR[a.id] }">
          <i :class="ATTR_ICON[a.id]" />
        </div>
        <span class="sv">{{ a.val }}</span>
        <span class="sk">{{ a.key }}</span>
      </div>
    </div>

    <!-- today's quests -->
    <div class="fx ac jb">
      <span class="sec">Quest hari ini</span>
      <NuxtLink to="/quests" class="tny">{{ doneCount }} / {{ quests.length }} selesai</NuxtLink>
    </div>
    <div class="fx col gap8">
      <QuestRow v-for="q in todayQuests" :key="q.id" :quest="q" @toggle="toggle" />
      <NuxtLink v-if="quests.length > 3" to="/quests" class="mut" style="text-align:center;font:600 12px 'Plus Jakarta Sans';padding:6px">
        Lihat semua quest
      </NuxtLink>
    </div>

    <!-- friend activity -->
    <div class="fx ac jb">
      <span class="sec">Aktivitas teman</span>
      <NuxtLink to="/feed" class="tny" style="color:var(--primary)">Semua</NuxtLink>
    </div>
    <div>
      <div v-for="(p, i) in activity" :key="p.id" class="feed" :style="i === activity.length - 1 ? 'border:none' : ''">
        <div class="avatar" style="width:36px;height:36px">{{ p.author[0] }}</div>
        <div class="f1">
          <span style="font:600 12.5px 'Plus Jakarta Sans';color:var(--ink)"><b>{{ p.author }}</b> memperbarui progres</span>
        </div>
        <span class="chip chipv">Lvl {{ p.authorLevel }}</span>
      </div>
      <div v-if="!activity.length && !pending" class="mut" style="font-size:12px;padding:8px 0">Belum ada aktivitas.</div>
    </div>
  </div>

  <div v-else class="spinner" />
</template>
