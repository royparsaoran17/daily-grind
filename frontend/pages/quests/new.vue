<script setup lang="ts">
import type { QuestFormValue } from '~/components/QuestForm.vue'

const { create } = useQuests()
const { show } = useToast()
const { t } = useI18n()
const saving = ref(false)
const error = ref('')

async function onSubmit(v: QuestFormValue) {
  error.value = ''
  if (!v.name || !v.categoryId) {
    error.value = t('questForm.required')
    return
  }
  saving.value = true
  try {
    await create(v)
    show(t('quests.created'), 'ph-fill ph-target')
    await navigateTo('/quests')
  } catch (e: any) {
    error.value = e?.data?.error ?? t('auth.genericErr')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac gap12">
      <button class="backbtn" @click="navigateTo('/quests')"><i class="ph-bold ph-caret-left" /></button>
      <span class="h">{{ t('quests.newTitle') }}</span>
    </div>
    <QuestForm :saving="saving" :error="error" :submit-label="t('questForm.save')" @submit="onSubmit" />
  </div>
</template>
