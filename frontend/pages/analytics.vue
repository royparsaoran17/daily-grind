<script setup lang="ts">
import { idNum, ATTR_COLOR } from '~/utils/format'

interface Day { date: string; label: string; completions: number; exp: number }
interface CatCount { category: string; attribute: string; count: number }
interface Analytics {
  daily: Day[]
  byCategory: CatCount[]
  attributes: Record<string, number>
  totalCompletions: number
  activeDays: number
  currentStreak: number
  thisWeekExp: number
}

interface HeatDay { date: string; count: number }

const { t, locale } = useI18n()
const data = ref<Analytics | null>(null)
const heat = ref<HeatDay[]>([])
const pending = ref(true)
const metric = ref<'completions' | 'exp'>('completions')
const active = ref<number | null>(null) // hovered/tapped bar index

onMounted(async () => {
  try {
    const [a, h] = await Promise.all([
      useApi()<Analytics>('/analytics'),
      useApi()<HeatDay[]>('/analytics/heatmap'),
    ])
    data.value = a
    heat.value = h
  } finally {
    pending.value = false
  }
})

// --- heatmap: pad the start to a Sunday so columns == weeks, rows == weekday ---
const heatCells = computed<(HeatDay | null)[]>(() => {
  if (!heat.value.length) return []
  const first = new Date(heat.value[0].date + 'T00:00:00')
  const pad = first.getDay() // 0=Sun..6
  return [...Array(pad).fill(null), ...heat.value]
})
function heatColor(count: number): string {
  if (count <= 0) return 'var(--track)'
  const pct = count >= 4 ? 100 : 25 * count // 25/50/75/100
  return `color-mix(in srgb, var(--grn-ink) ${pct}%, transparent)`
}
function heatTitle(c: HeatDay | null): string {
  if (!c) return ''
  const d = new Date(c.date + 'T00:00:00').toLocaleDateString(locale.value === 'en' ? 'en-US' : 'id-ID', { day: 'numeric', month: 'short' })
  return `${d}: ${c.count}`
}

// --- bar chart geometry (viewBox 320x150) ---
const W = 320, H = 150, padB = 22, padT = 14
const vals = computed(() => (data.value?.daily ?? []).map((d) => (metric.value === 'exp' ? d.exp : d.completions)))
const maxVal = computed(() => Math.max(1, ...vals.value))
const bars = computed(() => {
  const days = data.value?.daily ?? []
  const n = days.length || 1
  const gap = 5
  const bw = (W - gap * (n - 1)) / n
  const usableH = H - padB - padT
  return days.map((d, i) => {
    const v = metric.value === 'exp' ? d.exp : d.completions
    const h = v === 0 ? 0 : Math.max(3, (v / maxVal.value) * usableH)
    return { x: i * (bw + gap), y: H - padB - h, w: bw, h, v, label: d.label, date: d.date }
  })
})

const catMax = computed(() => Math.max(1, ...(data.value?.byCategory ?? []).map((c) => c.count)))

function fmtDate(iso: string) {
  return new Date(iso).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'id-ID', { day: 'numeric', month: 'short' })
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac gap12">
      <button class="backbtn" @click="navigateTo('/profile')"><i class="ph-bold ph-caret-left" /></button>
      <span class="h">{{ t('analytics.title') }}</span>
    </div>

    <div v-if="pending" class="spinner" />
    <template v-else-if="data">
      <!-- stat tiles -->
      <div class="fx gap10">
        <div class="card pad f1" style="padding:14px">
          <div class="tny">{{ t('analytics.totalDone') }}</div>
          <div style="font:700 22px 'Space Grotesk';color:var(--ink);margin-top:4px">{{ idNum(data.totalCompletions) }}</div>
        </div>
        <div class="card pad f1" style="padding:14px">
          <div class="tny">{{ t('analytics.activeDays') }}</div>
          <div style="font:700 22px 'Space Grotesk';color:var(--ink);margin-top:4px">{{ data.activeDays }}</div>
        </div>
      </div>
      <div class="fx gap10">
        <div class="card pad f1" style="padding:14px">
          <div class="tny">{{ t('analytics.streak') }}</div>
          <div style="font:700 22px 'Space Grotesk';color:var(--amber-ink);margin-top:4px">
            <i class="ph-fill ph-fire" style="font-size:16px" /> {{ data.currentStreak }}
          </div>
        </div>
        <div class="card pad f1" style="padding:14px">
          <div class="tny">{{ t('analytics.weekExp') }}</div>
          <div style="font:700 22px 'Space Grotesk';color:var(--primary);margin-top:4px">{{ idNum(data.thisWeekExp) }}</div>
        </div>
      </div>

      <!-- 14-day activity bar chart -->
      <div class="card pad">
        <div class="fx ac jb" style="margin-bottom:12px">
          <span class="sec">{{ t('analytics.last14') }}</span>
          <div class="seg" style="padding:3px">
            <button class="segi" :class="{ segon: metric === 'completions' }" style="padding:5px 10px;font-size:11px" @click="metric = 'completions'">{{ t('analytics.done') }}</button>
            <button class="segi" :class="{ segon: metric === 'exp' }" style="padding:5px 10px;font-size:11px" @click="metric = 'exp'">{{ t('analytics.exp') }}</button>
          </div>
        </div>

        <!-- tap-to-inspect value -->
        <div style="height:20px;text-align:center">
          <span v-if="active !== null && bars[active]" style="font:700 12px 'Space Grotesk';color:var(--primary)">
            {{ fmtDate(bars[active].date) }} · {{ bars[active].v }} {{ metric === 'exp' ? 'EXP' : t('analytics.doneUnit') }}
          </span>
          <span v-else class="mut" style="font-size:11px">{{ t('analytics.tapBar') }}</span>
        </div>

        <svg :viewBox="`0 0 ${W} ${H}`" width="100%" :height="H" role="img" aria-label="Aktivitas 14 hari terakhir">
          <!-- baseline -->
          <line :x1="0" :y1="H - padB" :x2="W" :y2="H - padB" stroke="var(--line)" stroke-width="1" />
          <g v-for="(b, i) in bars" :key="i" @click="active = active === i ? null : i" style="cursor:pointer">
            <!-- invisible full-height hit target -->
            <rect :x="b.x" :y="padT" :width="b.w" :height="H - padB - padT" fill="transparent" />
            <rect
              v-if="b.h > 0" :x="b.x" :y="b.y" :width="b.w" :height="b.h" rx="3"
              :fill="active === i ? 'var(--primary2)' : 'var(--primary)'"
              :opacity="active === null || active === i ? 1 : 0.45"
            />
            <text
              v-if="i % 2 === 0" :x="b.x + b.w / 2" :y="H - padB + 13"
              text-anchor="middle" style="font:600 8px 'Plus Jakarta Sans'" fill="var(--muted)"
            >{{ b.label }}</text>
          </g>
        </svg>
      </div>

      <!-- consistency heatmap -->
      <div class="card pad">
        <span class="sec" style="display:block;margin-bottom:12px">{{ t('analytics.heatmap') }}</span>
        <div style="overflow-x:auto;padding-bottom:4px">
          <div
            style="display:grid;grid-auto-flow:column;grid-template-rows:repeat(7,1fr);gap:3px;width:max-content"
          >
            <div
              v-for="(c, i) in heatCells" :key="i"
              :title="heatTitle(c)"
              :style="{ width: '11px', height: '11px', borderRadius: '3px', background: c ? heatColor(c.count) : 'transparent' }"
            />
          </div>
        </div>
        <div class="fx ac gap8" style="justify-content:flex-end;margin-top:8px">
          <span class="mut" style="font-size:9.5px">{{ t('analytics.less') }}</span>
          <span v-for="n in [0,1,2,3,4]" :key="n" :style="{ width:'11px',height:'11px',borderRadius:'3px',background: heatColor(n) }" />
          <span class="mut" style="font-size:9.5px">{{ t('analytics.more') }}</span>
        </div>
      </div>

      <!-- completions by category -->
      <div class="card pad">
        <span class="sec" style="display:block;margin-bottom:14px">{{ t('analytics.byCategory') }}</span>
        <div v-if="!data.byCategory.length" class="mut" style="font-size:12px">{{ t('analytics.noData') }}</div>
        <div v-else class="fx col gap12">
          <div v-for="c in data.byCategory" :key="c.category">
            <div class="fx ac jb" style="margin-bottom:5px">
              <span style="font:600 12px 'Plus Jakarta Sans';color:var(--ink)">{{ c.category }}</span>
              <span class="sv2">{{ c.count }}</span>
            </div>
            <div class="sbtrack">
              <div class="sbfill" :style="{ width: (c.count / catMax) * 100 + '%', background: ATTR_COLOR[c.attribute] || 'var(--primary)' }" />
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
