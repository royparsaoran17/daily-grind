<script setup lang="ts">
import type { QuestFormValue } from '~/components/QuestForm.vue'

const route = useRoute()
const { quests, load, update, remove } = useQuests()
const { show } = useToast()
const { t } = useI18n()

const id = route.params.id as string
const saving = ref(false)
const deleting = ref(false)
const error = ref('')
const pending = ref(true)

const quest = computed(() => quests.value.find((q) => q.id === id))
const initial = computed<Partial<QuestFormValue> | undefined>(() => {
  const q = quest.value
  if (!q) return undefined
  return {
    name: q.name,
    categoryId: q.categoryId,
    frequency: q.frequency as QuestFormValue['frequency'],
    difficulty: q.difficulty as QuestFormValue['difficulty'],
    reminder: q.reminder ?? '',
    weekday: q.weekday ?? 1,
    dayOfMonth: q.dayOfMonth ?? 1,
  }
})

onMounted(async () => {
  try { await load() } finally { pending.value = false }
})

async function onSubmit(v: QuestFormValue) {
  error.value = ''
  if (!v.name || !v.categoryId) {
    error.value = t('questForm.required')
    return
  }
  saving.value = true
  try {
    await update(id, v)
    show(t('quests.updated'), 'ph-fill ph-check-circle')
    await navigateTo('/quests')
  } catch (e: any) {
    error.value = e?.data?.error ?? t('auth.genericErr')
  } finally {
    saving.value = false
  }
}

async function destroy() {
  if (!confirm(t('quests.deleteConfirm'))) return
  deleting.value = true
  try {
    await remove(id)
    show(t('quests.deleted'), 'ph-fill ph-trash')
    await navigateTo('/quests')
  } finally {
    deleting.value = false
  }
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac gap12">
      <button class="backbtn" @click="navigateTo('/quests')"><i class="ph-bold ph-caret-left" /></button>
      <span class="h">{{ t('quests.editTitle') }}</span>
    </div>

    <div v-if="pending" class="spinner" />
    <div v-else-if="!quest" class="card pad mut" style="text-align:center;font-size:13px">
      {{ t('quests.notFound') }}
    </div>
    <template v-else>
      <QuestForm :initial="initial" :saving="saving" :error="error" :submit-label="t('quests.saveChanges')" @submit="onSubmit" />
      <button class="btn btno" style="color:var(--str);border-color:var(--str)" :disabled="deleting" @click="destroy">
        <i class="ph-bold ph-trash" />{{ deleting ? t('quests.deleting') : t('quests.deleteQuest') }}
      </button>
    </template>
  </div>
</template>
