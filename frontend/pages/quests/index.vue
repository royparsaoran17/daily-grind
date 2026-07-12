<script setup lang="ts">
const { quests, load, toggle, doneCount, totalExpToday } = useQuests()
const freq = ref<'daily' | 'weekly' | 'monthly'>('daily')
const pending = ref(true)

const filtered = computed(() => quests.value.filter((q) => q.frequency === freq.value))
const pct = computed(() =>
  filtered.value.length ? Math.round((filtered.value.filter((q) => q.done).length / filtered.value.length) * 100) : 0,
)
const doneInTab = computed(() => filtered.value.filter((q) => q.done).length)
const periodLabel = computed(() => (freq.value === 'daily' ? 'hari ini' : freq.value === 'weekly' ? 'minggu ini' : 'bulan ini'))

const today = new Date().toLocaleDateString('id-ID', { weekday: 'short', day: 'numeric', month: 'short' })

onMounted(async () => {
  try { await load(true) } finally { pending.value = false }
})
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac jb">
      <span class="h">Quest</span>
      <span class="chip"><i class="ph ph-calendar-blank" />{{ today }}</span>
    </div>

    <div class="seg">
      <button class="segi" :class="{ segon: freq === 'daily' }" @click="freq = 'daily'">Harian</button>
      <button class="segi" :class="{ segon: freq === 'weekly' }" @click="freq = 'weekly'">Mingguan</button>
      <button class="segi" :class="{ segon: freq === 'monthly' }" @click="freq = 'monthly'">Bulanan</button>
    </div>

    <div class="card pad spot" style="border-radius:20px">
      <div class="fx ac jb">
        <span style="font:600 12px 'Plus Jakarta Sans';color:rgba(255,255,255,.7)">Progres {{ periodLabel }}</span>
        <span style="font:700 13px 'Space Grotesk';color:var(--amber)">+{{ totalExpToday }} EXP</span>
      </div>
      <div class="fx ac gap12" style="margin-top:10px">
        <div class="xp" style="flex:1;background:rgba(255,255,255,.18)"><div class="xpf" :style="{ width: pct + '%', background: 'var(--amber)' }" /></div>
        <span style="font:700 13px 'Space Grotesk'">{{ doneInTab }}/{{ filtered.length }}</span>
      </div>
    </div>

    <div v-if="pending" class="spinner" />
    <div v-else class="fx col gap10">
      <QuestRow v-for="q in filtered" :key="q.id" :quest="q" show-streak show-schedule @toggle="toggle" />
      <div v-if="!filtered.length" class="mut" style="text-align:center;font-size:12.5px;padding:20px">
        Belum ada quest {{ freq === 'daily' ? 'harian' : freq === 'weekly' ? 'mingguan' : 'bulanan' }}.
      </div>
    </div>

    <NuxtLink to="/quests/new" class="fab"><i class="ph-bold ph-plus" /></NuxtLink>
  </div>
</template>
