<script setup lang="ts">
interface Category { id: string; label: string; icon: string; attribute: string }
export interface QuestFormValue {
  name: string; categoryId: string
  frequency: 'daily' | 'weekly' | 'monthly'
  difficulty: 'easy' | 'medium' | 'hard'
  reminder: string; weekday: number | null; dayOfMonth: number | null
}

const props = defineProps<{
  initial?: Partial<QuestFormValue>
  submitLabel?: string
  saving?: boolean
  error?: string
}>()
const emit = defineEmits<{ (e: 'submit', v: QuestFormValue): void }>()
const { t, locale } = useI18n()

const rewardTable = { easy: [20, 8], medium: [40, 15], hard: [60, 25] } as const
const WEEKDAYS_ID = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab']
const WEEKDAYS_EN = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
const WEEKDAYS = computed(() => (locale.value === 'en' ? WEEKDAYS_EN : WEEKDAYS_ID))

const categories = ref<Category[]>([])
const name = ref(props.initial?.name ?? '')
const categoryId = ref(props.initial?.categoryId ?? '')
const frequency = ref<QuestFormValue['frequency']>(props.initial?.frequency ?? 'daily')
const difficulty = ref<QuestFormValue['difficulty']>(props.initial?.difficulty ?? 'medium')
const reminder = ref(props.initial?.reminder ?? '')
const weekday = ref(props.initial?.weekday ?? 1)
const dayOfMonth = ref(props.initial?.dayOfMonth ?? 1)

const selected = computed(() => categories.value.find((c) => c.id === categoryId.value))
const reward = computed(() => rewardTable[difficulty.value])

onMounted(async () => {
  categories.value = await useApi()<Category[]>('/categories')
  if (!categoryId.value && categories.value.length) categoryId.value = categories.value[0].id
})

function submit() {
  emit('submit', {
    name: name.value.trim(),
    categoryId: categoryId.value,
    frequency: frequency.value,
    difficulty: difficulty.value,
    reminder: reminder.value.trim(),
    weekday: frequency.value === 'weekly' ? weekday.value : null,
    dayOfMonth: frequency.value === 'monthly' ? dayOfMonth.value : null,
  })
}
</script>

<template>
  <div class="fx col gap16">
    <div>
      <span class="flabel">{{ t('questForm.name') }}</span>
      <div class="field">
        <i class="ph ph-pencil-simple" />
        <input v-model="name" type="text" :placeholder="t('questForm.namePh')">
      </div>
    </div>

    <div>
      <span class="flabel">{{ t('questForm.frequency') }}</span>
      <div class="seg">
        <button class="segi" :class="{ segon: frequency === 'daily' }" @click="frequency = 'daily'">{{ t('quests.daily') }}</button>
        <button class="segi" :class="{ segon: frequency === 'weekly' }" @click="frequency = 'weekly'">{{ t('quests.weekly') }}</button>
        <button class="segi" :class="{ segon: frequency === 'monthly' }" @click="frequency = 'monthly'">{{ t('quests.monthly') }}</button>
      </div>

      <div v-if="frequency === 'weekly'" style="margin-top:10px">
        <span class="flabel">{{ t('questForm.everyDay') }}</span>
        <div class="fx wrap gap8">
          <button
            v-for="(d, i) in WEEKDAYS" :key="i" class="chip"
            :class="{ chipv: weekday === i }" style="border:none;min-width:44px;justify-content:center"
            @click="weekday = i"
          >{{ d }}</button>
        </div>
      </div>

      <div v-if="frequency === 'monthly'" style="margin-top:10px">
        <span class="flabel">{{ t('questForm.everyDate') }}</span>
        <div class="fx wrap gap8" style="max-height:120px;overflow:auto">
          <button
            v-for="n in 28" :key="n" class="chip"
            :class="{ chipv: dayOfMonth === n }" style="border:none;min-width:38px;justify-content:center"
            @click="dayOfMonth = n"
          >{{ n }}</button>
        </div>
      </div>
    </div>

    <div>
      <span class="flabel">{{ t('questForm.category') }}</span>
      <div class="fx wrap gap8">
        <button
          v-for="c in categories" :key="c.id" class="chip"
          :class="{ chipv: c.id === categoryId }" style="border:none" @click="categoryId = c.id"
        >
          <i :class="c.id === categoryId ? `ph-fill ${c.icon}` : `ph ${c.icon}`" />{{ c.label }}
        </button>
      </div>
      <div v-if="selected" class="mapbox">
        <i class="ph-fill ph-arrow-bend-down-right" style="color:var(--primary);font-size:16px" />
        <span style="font:600 12px 'Plus Jakarta Sans';color:var(--pink)">
          {{ t('questForm.raises') }} <b>{{ t('attr.' + selected.attribute) }}</b>
        </span>
      </div>
    </div>

    <div>
      <span class="flabel">{{ t('questForm.difficulty') }}</span>
      <div class="seg">
        <button class="segi" :class="{ segon: difficulty === 'easy' }" @click="difficulty = 'easy'">{{ t('questForm.easy') }}</button>
        <button class="segi" :class="{ segon: difficulty === 'medium' }" @click="difficulty = 'medium'">{{ t('questForm.medium') }}</button>
        <button class="segi" :class="{ segon: difficulty === 'hard' }" @click="difficulty = 'hard'">{{ t('questForm.hard') }}</button>
      </div>
      <div class="fx ac jb" style="margin-top:10px">
        <span class="mut" style="font:600 11.5px 'Plus Jakarta Sans'">{{ t('questForm.rewardPer') }}</span>
        <div class="fx gap8">
          <span class="chipa pill"><i class="ph-fill ph-lightning" />+{{ reward[0] }} EXP</span>
          <span class="chipa pill"><i class="ph-fill ph-coins" />+{{ reward[1] }}</span>
        </div>
      </div>
    </div>

    <div>
      <span class="flabel">{{ t('questForm.reminder') }}</span>
      <div class="field">
        <i class="ph ph-bell" />
        <input v-model="reminder" type="text" :placeholder="t('questForm.reminderPh')">
      </div>
    </div>

    <p v-if="error" style="color:var(--str);font:600 12px 'Plus Jakarta Sans';margin:0;text-align:center">{{ error }}</p>

    <button class="btn" :disabled="saving" @click="submit">
      <i class="ph-bold ph-check" />{{ saving ? t('common.saving') : (submitLabel ?? t('questForm.save')) }}
    </button>
  </div>
</template>
