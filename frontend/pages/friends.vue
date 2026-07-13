<script setup lang="ts">
interface Friend { id: string; name: string; avatar?: string; level: number; streak: number; weeklyExp: number; isMe: boolean }
interface UserResult { id: string; name: string; avatar?: string; level: number; title: string; status: 'none' | 'outgoing' | 'incoming' | 'friend' }
interface Request { id: string; name: string; avatar?: string; level: number; title: string }

const { show } = useToast()
const { t } = useI18n()

const friends = ref<Friend[]>([])
const requests = ref<Request[]>([])
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
  const [f, r] = await Promise.all([
    useApi()<Friend[]>('/friends'),
    useApi()<Request[]>('/friends/requests'),
  ])
  friends.value = f
  requests.value = r
}

onMounted(async () => {
  try { await refresh() } finally { pending.value = false }
})

async function accept(req: Request) {
  await useApi()(`/friends/requests/${req.id}/accept`, { method: 'POST' })
  show(t('friends.accepted', { name: req.name }), 'ph-fill ph-user-plus')
  await refresh()
}
async function reject(req: Request) {
  await useApi()(`/friends/requests/${req.id}/reject`, { method: 'POST' })
  requests.value = requests.value.filter((r) => r.id !== req.id)
  show(t('friends.rejected'), 'ph-fill ph-x-circle')
}

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
  const res = await useApi()<{ status: 'outgoing' | 'friend' }>(`/friends/${u.id}`, { method: 'POST' })
  u.status = res.status
  show(res.status === 'friend' ? t('friends.added', { name: u.name }) : t('friends.requestSent', { name: u.name }), 'ph-fill ph-user-plus')
  await refresh()
}

async function removeFriend(f: Friend) {
  await useApi()(`/friends/${f.id}`, { method: 'DELETE' })
  await refresh()
  show(t('friends.removed', { name: f.name }), 'ph-fill ph-user-minus')
}

function toggleAdd() {
  adding.value = !adding.value
  if (!adding.value) { query.value = ''; results.value = [] }
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac jb">
      <span class="h">{{ t('friends.title') }}</span>
      <button
        class="avatar" style="width:38px;height:38px;background:var(--psoft);color:var(--primary);font-size:18px;border:none"
        @click="toggleAdd"
      >
        <i :class="adding ? 'ph-bold ph-x' : 'ph-bold ph-user-plus'" />
      </button>
    </div>

    <!-- add-friend panel -->
    <div v-if="adding" class="card pad fx col gap12">
      <span class="sec">{{ t('friends.addTitle') }}</span>
      <div class="field" style="border-radius:14px;background:var(--soft);border:none">
        <i class="ph ph-magnifying-glass" />
        <input v-model="query" type="text" :placeholder="t('friends.searchUser')" autofocus>
      </div>
      <div v-if="searching" class="mut" style="font-size:12px">{{ t('friends.searching') }}</div>
      <div v-else-if="query.trim() && !results.length" class="mut" style="font-size:12px">{{ t('friends.noUsers') }}</div>
      <div v-else class="fx col gap8">
        <div v-for="u in results" :key="u.id" class="fx ac gap12">
          <Avatar :url="u.avatar" :name="u.name" :size="40" />
          <div class="f1">
            <div style="font:700 13px 'Plus Jakarta Sans';color:var(--ink)">{{ u.name }}</div>
            <div class="mut" style="font-size:10.5px">{{ u.title }} · Lvl {{ u.level }}</div>
          </div>
          <span v-if="u.status === 'friend'" class="chip chipg" style="border:none"><i class="ph-fill ph-check" />{{ t('friends.friend') }}</span>
          <span v-else-if="u.status === 'outgoing'" class="chip" style="border:none;opacity:.7"><i class="ph ph-paper-plane-tilt" />{{ t('friends.requested') }}</span>
          <button v-else class="chip chipv" style="border:none" @click="addFriend(u)">
            <i class="ph-bold ph-plus" />{{ u.status === 'incoming' ? t('friends.accept') : t('common.add') }}
          </button>
        </div>
      </div>
    </div>

    <div class="field" style="border-radius:14px;background:var(--soft);border:none">
      <i class="ph ph-magnifying-glass" />
      <input v-model="search" type="text" :placeholder="t('friends.searchFriends')">
    </div>

    <!-- incoming friend requests -->
    <div v-if="requests.length" class="card pad">
      <span class="sec" style="display:block;margin-bottom:12px">{{ t('friends.requests') }} ({{ requests.length }})</span>
      <div class="fx col gap12">
        <div v-for="req in requests" :key="req.id" class="fx ac gap12">
          <Avatar :url="req.avatar" :name="req.name" :size="40" />
          <div class="f1">
            <div style="font:700 13px 'Plus Jakarta Sans';color:var(--ink)">{{ req.name }}</div>
            <div class="mut" style="font-size:10.5px">{{ req.title }} · {{ t('common.level') }} {{ req.level }}</div>
          </div>
          <button class="chip chipg" style="border:none" @click="accept(req)"><i class="ph-bold ph-check" />{{ t('friends.accept') }}</button>
          <button class="chip" style="border:none" @click="reject(req)"><i class="ph-bold ph-x" /></button>
        </div>
      </div>
    </div>

    <div v-if="pending" class="spinner" />
    <template v-else>
      <!-- leaderboard -->
      <div class="card pad spot" style="border-radius:20px">
        <div class="fx ac jb" style="margin-bottom:14px">
          <span style="font:700 13px 'Space Grotesk'">{{ t('friends.leaderboard') }}</span>
          <span class="chip" style="background:rgba(255,255,255,.14);color:#fff">{{ t('friends.thisWeek') }}</span>
        </div>
        <div
          v-for="(f, i) in leaderboard" :key="f.id" class="fx ac gap12"
          style="padding:9px 0" :style="i > 0 ? 'border-top:1px solid rgba(255,255,255,.1)' : ''"
        >
          <span class="rankn" :style="{ color: rankColor[i] }">{{ i + 1 }}</span>
          <Avatar :url="f.avatar" :name="f.name" :size="36" />
          <div class="f1">
            <div style="font:700 13px 'Plus Jakarta Sans'">{{ f.isMe ? `${f.name} (${t('friends.you')})` : f.name }}</div>
            <div style="font-size:10.5px;color:rgba(255,255,255,.6)">{{ t('friends.expThisWeek', { n: f.weeklyExp }) }}</div>
          </div>
          <span class="pill" style="background:rgba(242,166,59,.25);color:var(--amber)">{{ t('common.level') }} {{ f.level }}</span>
        </div>
      </div>

      <span class="sec">{{ t('friends.allFriends') }}</span>
      <div>
        <div v-for="(f, i) in others" :key="f.id" class="rank" :style="i === others.length - 1 ? 'border:none' : ''">
          <Avatar :url="f.avatar" :name="f.name" :size="44" />
          <div class="f1">
            <div style="font:700 13.5px 'Plus Jakarta Sans';color:var(--ink)">{{ f.name }}</div>
            <div class="fx ac gap8" style="margin-top:3px">
              <span class="chip chipa" style="padding:2px 7px;font-size:10px"><i class="ph-fill ph-fire" />{{ f.streak }}</span>
              <span class="mut" style="font-size:11px">{{ t('common.level') }} {{ f.level }}</span>
            </div>
          </div>
          <button class="pact" style="padding:0" @click="removeFriend(f)"><i class="ph ph-user-minus" style="font-size:18px" /></button>
        </div>
        <div v-if="!others.length" class="mut" style="font-size:12.5px;padding:14px 0">
          {{ t('friends.noneYet') }}
        </div>
      </div>
    </template>
  </div>
</template>
