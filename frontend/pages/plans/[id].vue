<script setup lang="ts">
import type { User } from '~/stores/auth'

interface Reading { bookId: string; chapter: number; label: string }
interface Day { day: number; label: string; readings: Reading[]; completed: boolean }
interface Plan {
  id: string; title: string; description: string; icon: string
  totalDays: number; enrolled: boolean; completed: number; faithReward: number; days: Day[]
}

const route = useRoute()
const auth = useAuthStore()
const { show } = useToast()
const { t } = useI18n()

const id = route.params.id as string
const plan = ref<Plan | null>(null)
const pending = ref(true)
const busy = ref<number | null>(null)

const completedCount = computed(() => plan.value?.days.filter((d) => d.completed).length ?? 0)
const pct = computed(() => (plan.value?.totalDays ? Math.round((completedCount.value / plan.value.totalDays) * 100) : 0))

async function load() {
  plan.value = await useApi()<Plan>(`/reading-plans/${id}`)
}
onMounted(async () => {
  try { await load() } finally { pending.value = false }
})

async function join() {
  await useApi()(`/reading-plans/${id}/enroll`, { method: 'POST' })
  if (plan.value) plan.value.enrolled = true
}
async function leave() {
  await useApi()(`/reading-plans/${id}/enroll`, { method: 'DELETE' })
  if (plan.value) plan.value.enrolled = false
}

async function toggleDay(d: Day) {
  busy.value = d.day
  const wasDone = d.completed
  try {
    const user = await useApi()<User>(`/reading-plans/${id}/days/${d.day}/complete`, {
      method: wasDone ? 'DELETE' : 'POST',
    })
    auth.setUser(user)
    d.completed = !wasDone
    if (!wasDone) {
      if (plan.value) plan.value.enrolled = true
      const allDone = plan.value?.days.every((x) => x.completed)
      show(allDone ? t('plans.finished') : `+${plan.value?.faithReward} FAITH 🙏`, 'ph-fill ph-hands-praying')
    }
  } finally {
    busy.value = null
  }
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac gap12">
      <button class="backbtn" @click="navigateTo('/plans')"><i class="ph-bold ph-caret-left" /></button>
      <span class="h">{{ plan?.title ?? '…' }}</span>
    </div>

    <div v-if="pending" class="spinner" />
    <template v-else-if="plan">
      <!-- header card -->
      <div class="card pad" style="background:linear-gradient(140deg,var(--primary),var(--primary2));border:none;color:#fff">
        <div class="fx ac gap12">
          <div class="ring" :style="{ background: `conic-gradient(#fff ${pct}%, rgba(255,255,255,.28) 0)` }">
            <div class="avatar" style="width:52px;height:52px;background:rgba(255,255,255,.15);color:#fff;font-size:24px"><i :class="`ph-fill ${plan.icon}`" /></div>
          </div>
          <div class="f1">
            <div style="font:400 12px 'Plus Jakarta Sans';color:rgba(255,255,255,.85);line-height:1.4">{{ plan.description }}</div>
            <div style="font:700 13px 'Space Grotesk';margin-top:6px">{{ t('plans.progress', { done: completedCount, total: plan.totalDays }) }} · {{ pct }}%</div>
          </div>
        </div>
      </div>

      <button v-if="!plan.enrolled" class="btn" @click="join">{{ t('plans.join') }}</button>
      <button v-else class="btn btno" @click="leave">{{ t('plans.leave') }}</button>

      <!-- days -->
      <div class="fx col gap10">
        <div v-for="d in plan.days" :key="d.day" class="quest" style="align-items:flex-start">
          <button class="chk" :class="{ chkon: d.completed }" :disabled="busy === d.day" @click="toggleDay(d)">
            <i v-if="d.completed" class="ph-bold ph-check" />
          </button>
          <div class="f1">
            <div style="font:700 12.5px 'Plus Jakarta Sans';color:var(--ink)">{{ t('plans.dayLabel') }} {{ d.day }} · {{ d.label }}</div>
            <div class="fx wrap gap8" style="margin-top:6px">
              <NuxtLink
                v-for="(rd, i) in d.readings" :key="i"
                :to="`/bible?bookId=${rd.bookId}&chapter=${rd.chapter}`"
                class="chip chipv" style="border:none;font-size:10.5px"
              >
                <i class="ph-fill ph-book-open" />{{ t('plans.read') }}
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
