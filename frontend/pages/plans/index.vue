<script setup lang="ts">
interface Plan {
  id: string; title: string; description: string; icon: string
  totalDays: number; enrolled: boolean; completed: number; faithReward: number
}
const { t } = useI18n()
const plans = ref<Plan[]>([])
const pending = ref(true)

onMounted(async () => {
  try { plans.value = await useApi()<Plan[]>('/reading-plans') } finally { pending.value = false }
})
function pct(p: Plan) { return p.totalDays ? Math.round((p.completed / p.totalDays) * 100) : 0 }
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac gap12">
      <button class="backbtn" @click="navigateTo('/bible')"><i class="ph-bold ph-caret-left" /></button>
      <div>
        <span class="h">{{ t('plans.title') }}</span>
        <div class="mut" style="font-size:11.5px;margin-top:2px">{{ t('plans.subtitle') }}</div>
      </div>
    </div>

    <div v-if="pending" class="spinner" />
    <div v-else-if="!plans.length" class="mut" style="text-align:center;font-size:12.5px;padding:20px">{{ t('plans.empty') }}</div>
    <div v-else class="fx col gap12">
      <NuxtLink v-for="p in plans" :key="p.id" :to="`/plans/${p.id}`" class="card pad fx ac gap12">
        <div class="ring" style="flex:none" :style="{ background: `conic-gradient(var(--primary) ${pct(p)}%, var(--track) 0)` }">
          <div class="avatar" style="width:46px;height:46px;background:var(--psoft);color:var(--primary);font-size:22px">
            <i :class="`ph-fill ${p.icon}`" />
          </div>
        </div>
        <div class="f1">
          <div style="font:700 14px 'Space Grotesk';color:var(--ink)">{{ p.title }}</div>
          <div class="mut" style="font-size:11px;line-height:1.4;margin:2px 0 6px">{{ p.description }}</div>
          <div class="fx ac gap8">
            <span class="chip chipa" style="padding:2px 8px;font-size:10px"><i class="ph-fill ph-hands-praying" />+{{ p.faithReward }} FAITH</span>
            <span class="mut" style="font-size:10.5px">{{ t('plans.progress', { done: p.completed, total: p.totalDays }) }}</span>
            <span v-if="p.enrolled" class="chip chipg" style="padding:2px 8px;font-size:10px"><i class="ph-fill ph-check" />{{ t('plans.joined') }}</span>
          </div>
        </div>
        <i class="ph ph-caret-right mut" />
      </NuxtLink>
    </div>
  </div>
</template>
