<script setup lang="ts">
interface Prayer { id: string; title: string; body: string; answered: boolean; answeredAt?: string; createdAt: string }

const { show } = useToast()
const { t, locale } = useI18n()

const prayers = ref<Prayer[]>([])
const pending = ref(true)
const adding = ref(false)
const title = ref('')
const body = ref('')
const saving = ref(false)

const active = computed(() => prayers.value.filter((p) => !p.answered))
const answered = computed(() => prayers.value.filter((p) => p.answered))

async function refresh() { prayers.value = await useApi()<Prayer[]>('/prayers') }
onMounted(async () => { try { await refresh() } finally { pending.value = false } })

async function save() {
  if (!title.value.trim()) return
  saving.value = true
  try {
    await useApi()('/prayers', { method: 'POST', body: { title: title.value, body: body.value } })
    title.value = ''; body.value = ''; adding.value = false
    await refresh()
    show(t('prayers.added'), 'ph-fill ph-hands-praying')
  } finally {
    saving.value = false
  }
}
async function toggle(p: Prayer) {
  const res = await useApi()<{ answered: boolean }>(`/prayers/${p.id}/answer`, { method: 'POST' })
  p.answered = res.answered
  p.answeredAt = res.answered ? new Date().toISOString() : undefined
  await refresh()
  if (res.answered) show(t('prayers.answeredToast'), 'ph-fill ph-sparkle')
}
async function remove(p: Prayer) {
  await useApi()(`/prayers/${p.id}`, { method: 'DELETE' })
  prayers.value = prayers.value.filter((x) => x.id !== p.id)
}
function fmt(iso?: string) {
  return iso ? new Date(iso).toLocaleDateString(locale.value === 'en' ? 'en-US' : 'id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) : ''
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac jb">
      <div class="fx ac gap12">
        <button class="backbtn" @click="navigateTo('/bible')"><i class="ph-bold ph-caret-left" /></button>
        <span class="h">{{ t('prayers.title') }}</span>
      </div>
      <button class="avatar" style="width:38px;height:38px;background:var(--psoft);color:var(--primary);font-size:18px;border:none" @click="adding = !adding">
        <i :class="adding ? 'ph-bold ph-x' : 'ph-bold ph-plus'" />
      </button>
    </div>

    <!-- add form -->
    <div v-if="adding" class="card pad fx col gap10">
      <div class="field"><i class="ph ph-pencil-simple" /><input v-model="title" type="text" :placeholder="t('prayers.titlePh')"></div>
      <textarea v-model="body" class="compin" rows="3" style="border:1.5px solid var(--fline);border-radius:14px;padding:12px 14px" :placeholder="t('prayers.bodyPh')" />
      <button class="btn" :disabled="saving || !title.trim()" @click="save"><i class="ph-bold ph-check" />{{ saving ? t('common.saving') : t('prayers.save') }}</button>
    </div>

    <div v-if="pending" class="spinner" />
    <template v-else>
      <div v-if="!prayers.length" class="mut" style="text-align:center;font-size:12.5px;padding:20px">{{ t('prayers.empty') }}</div>

      <!-- active -->
      <template v-if="active.length">
        <span class="sec">{{ t('prayers.active') }} ({{ active.length }})</span>
        <div class="fx col gap10">
          <div v-for="p in active" :key="p.id" class="card pad fx ac gap12">
            <button class="chk" @click="toggle(p)" />
            <div class="f1">
              <div style="font:700 13px 'Plus Jakarta Sans';color:var(--ink)">{{ p.title }}</div>
              <div v-if="p.body" class="mut" style="font-size:11.5px;line-height:1.5;margin-top:2px">{{ p.body }}</div>
            </div>
            <button class="pact" style="padding:0" @click="remove(p)"><i class="ph ph-trash" style="font-size:16px" /></button>
          </div>
        </div>
      </template>

      <!-- answered -->
      <template v-if="answered.length">
        <span class="sec">{{ t('prayers.answered') }} ({{ answered.length }})</span>
        <div class="fx col gap10">
          <div v-for="p in answered" :key="p.id" class="card pad fx ac gap12" style="opacity:.85">
            <button class="chk chkon" @click="toggle(p)"><i class="ph-bold ph-check" /></button>
            <div class="f1">
              <div style="font:700 13px 'Plus Jakarta Sans';color:var(--ink);text-decoration:line-through">{{ p.title }}</div>
              <div class="chipg chip" style="padding:2px 8px;font-size:10px;margin-top:4px"><i class="ph-fill ph-sparkle" />{{ t('prayers.answeredOn', { date: fmt(p.answeredAt) }) }}</div>
            </div>
            <button class="pact" style="padding:0" @click="remove(p)"><i class="ph ph-trash" style="font-size:16px" /></button>
          </div>
        </div>
      </template>
    </template>
  </div>
</template>
