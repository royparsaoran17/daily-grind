<script setup lang="ts">
import { ATTR_COLOR } from '~/utils/format'

interface Quest {
  id: string; name: string; icon: string; attribute: string
  expReward: number; streak: number; done: boolean
  schedule?: string; dueToday?: boolean; frequency?: string
}
const props = defineProps<{ quest: Quest; showStreak?: boolean; showSchedule?: boolean; editable?: boolean }>()
const emit = defineEmits<{ (e: 'toggle', quest: Quest): void; (e: 'edit', quest: Quest): void }>()

const { t } = useI18n()
const tint = computed(() => ATTR_COLOR[props.quest.attribute] ?? 'var(--primary)')

// Play the celebration animation only when the tap *completes* the quest.
const celebrate = ref(false)
function onToggle() {
  if (!props.quest.done) {
    celebrate.value = true
    setTimeout(() => (celebrate.value = false), 550)
  }
  emit('toggle', props.quest)
}
</script>

<template>
  <div class="quest" role="button" tabindex="0" style="cursor:pointer" @click="onToggle">
    <span class="chk" :class="{ chkon: quest.done, celebrate }">
      <i v-if="quest.done" class="ph-bold ph-check" />
    </span>
    <span class="qi" :style="{ background: `color-mix(in srgb, ${tint} 15%, transparent)`, color: tint }">
      <i :class="quest.icon" />
    </span>
    <span class="f1">
      <span style="display:block;font:700 13.5px 'Plus Jakarta Sans';color:var(--ink)">{{ quest.name }}</span>
      <span class="fx ac gap8" style="margin-top:3px">
        <span class="mut" style="font-size:10.5px">→ {{ quest.attribute.toUpperCase() }}</span>
        <span v-if="showSchedule && quest.frequency !== 'daily'" class="mut" style="font-size:10.5px">
          · {{ quest.schedule }}
        </span>
        <span v-if="showSchedule && quest.dueToday && quest.frequency !== 'daily'" class="chipg pill" style="padding:1px 6px;font-size:9px">
          {{ t('common.today') }}
        </span>
        <span v-if="showStreak && quest.streak > 0" class="chipa pill" style="padding:1px 6px;font-size:9px">
          <i class="ph-fill ph-fire" />{{ quest.streak }}
        </span>
      </span>
    </span>
    <span class="chipa pill">+{{ quest.expReward }}</span>
    <button
      v-if="editable" class="pact" style="padding:0 0 0 4px" aria-label="Edit quest"
      @click.stop="emit('edit', quest)"
    >
      <i class="ph ph-dots-three-vertical" style="font-size:18px" />
    </button>
  </div>
</template>
