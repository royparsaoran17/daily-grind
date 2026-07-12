<script setup lang="ts">
interface Friend { id: string; name: string; level: number; streak: number; weeklyExp: number; isMe: boolean }
interface UserResult { id: string; name: string; level: number; title: string; isFriend: boolean }

const { show } = useToast()

const friends = ref<Friend[]>([])
const search = ref('')
const pending = ref(true)

// add-friend panel
const adding = ref(false)
const query = ref('')
const results = ref<UserResult[]>([])
const searching = ref(false)
let searchSeq = 0

const leaderboard = computed(() => friends.value.slice(0, 3))
const others = computed(() =>
  friends.value.filter((f) => !f.isMe && f.name.toLowerCase().includes(search.value.toLowerCase())),
)
const rankColor = ['var(--amber)', '#cfcbc2', '#c88a1c']

async function refresh() {
  friends.value = await useApi()<Friend[]>('/friends')
}

onMounted(async () => {
  try { await refresh() } finally { pending.value = false }
})

// Search users as you type (guarded against out-of-order responses).
watch(query, async (q) => {
  const term = q.trim()
  const seq = ++searchSeq
  if (!term) { results.value = []; searching.value = false; return }
  searching.value = true
  try {
    const res = await useApi()<UserResult[]>(`/users/search?q=${encodeURIComponent(term)}`)
    if (seq === searchSeq) results.value = res
  } finally {
    if (seq === searchSeq) searching.value = false
  }
})

async function addFriend(u: UserResult) {
  await useApi()(`/friends/${u.id}`, { method: 'POST' })
  u.isFriend = true
  show(`${u.name} ditambahkan sebagai teman`, 'ph-fill ph-user-plus')
  await refresh()
}

async function removeFriend(f: Friend) {
  await useApi()(`/friends/${f.id}`, { method: 'DELETE' })
  await refresh()
  show(`${f.name} dihapus dari teman`, 'ph-fill ph-user-minus')
}

function toggleAdd() {
  adding.value = !adding.value
  if (!adding.value) { query.value = ''; results.value = [] }
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac jb">
      <span class="h">Teman</span>
      <button
        class="avatar" style="width:38px;height:38px;background:var(--psoft);color:var(--primary);font-size:18px;border:none"
        @click="toggleAdd"
      >
        <i :class="adding ? 'ph-bold ph-x' : 'ph-bold ph-user-plus'" />
      </button>
    </div>

    <!-- add-friend panel -->
    <div v-if="adding" class="card pad fx col gap12">
      <span class="sec">Tambah teman</span>
      <div class="field" style="border-radius:14px;background:var(--soft);border:none">
        <i class="ph ph-magnifying-glass" />
        <input v-model="query" type="text" placeholder="Cari nama pengguna…" autofocus>
      </div>
      <div v-if="searching" class="mut" style="font-size:12px">Mencari…</div>
      <div v-else-if="query.trim() && !results.length" class="mut" style="font-size:12px">Tidak ada pengguna ditemukan.</div>
      <div v-else class="fx col gap8">
        <div v-for="u in results" :key="u.id" class="fx ac gap12">
          <div class="avatar" style="width:40px;height:40px">{{ u.name[0] }}</div>
          <div class="f1">
            <div style="font:700 13px 'Plus Jakarta Sans';color:var(--ink)">{{ u.name }}</div>
            <div class="mut" style="font-size:10.5px">{{ u.title }} · Lvl {{ u.level }}</div>
          </div>
          <span v-if="u.isFriend" class="chip chipg" style="border:none"><i class="ph-fill ph-check" />Teman</span>
          <button v-else class="chip chipv" style="border:none" @click="addFriend(u)"><i class="ph-bold ph-plus" />Tambah</button>
        </div>
      </div>
    </div>

    <div class="field" style="border-radius:14px;background:var(--soft);border:none">
      <i class="ph ph-magnifying-glass" />
      <input v-model="search" type="text" placeholder="Cari teman…">
    </div>

    <div v-if="pending" class="spinner" />
    <template v-else>
      <!-- leaderboard -->
      <div class="card pad spot" style="border-radius:20px">
        <div class="fx ac jb" style="margin-bottom:14px">
          <span style="font:700 13px 'Space Grotesk'">Papan Peringkat</span>
          <span class="chip" style="background:rgba(255,255,255,.14);color:#fff">Minggu ini</span>
        </div>
        <div
          v-for="(f, i) in leaderboard" :key="f.id" class="fx ac gap12"
          style="padding:9px 0" :style="i > 0 ? 'border-top:1px solid rgba(255,255,255,.1)' : ''"
        >
          <span class="rankn" :style="{ color: rankColor[i] }">{{ i + 1 }}</span>
          <div class="avatar" style="width:36px;height:36px">{{ f.name[0] }}</div>
          <div class="f1">
            <div style="font:700 13px 'Plus Jakarta Sans'">{{ f.isMe ? f.name + ' (kamu)' : f.name }}</div>
            <div style="font-size:10.5px;color:rgba(255,255,255,.6)">{{ f.weeklyExp }} EXP minggu ini</div>
          </div>
          <span class="pill" style="background:rgba(242,166,59,.25);color:var(--amber)">Lvl {{ f.level }}</span>
        </div>
      </div>

      <span class="sec">Semua teman</span>
      <div>
        <div v-for="(f, i) in others" :key="f.id" class="rank" :style="i === others.length - 1 ? 'border:none' : ''">
          <div class="avatar" style="width:44px;height:44px">{{ f.name[0] }}</div>
          <div class="f1">
            <div style="font:700 13.5px 'Plus Jakarta Sans';color:var(--ink)">{{ f.name }}</div>
            <div class="fx ac gap8" style="margin-top:3px">
              <span class="chip chipa" style="padding:2px 7px;font-size:10px"><i class="ph-fill ph-fire" />{{ f.streak }}</span>
              <span class="mut" style="font-size:11px">Lvl {{ f.level }}</span>
            </div>
          </div>
          <button class="pact" style="padding:0" @click="removeFriend(f)"><i class="ph ph-user-minus" style="font-size:18px" /></button>
        </div>
        <div v-if="!others.length" class="mut" style="font-size:12.5px;padding:14px 0">
          Belum ada teman. Ketuk <i class="ph-bold ph-user-plus" /> untuk menambah.
        </div>
      </div>
    </template>
  </div>
</template>
