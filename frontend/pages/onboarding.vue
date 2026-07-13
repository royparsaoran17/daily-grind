<script setup lang="ts">
definePageMeta({ layout: 'auth' })

const auth = useAuthStore()
const { show } = useToast()
const { t } = useI18n()

const step = ref(0)
const saving = ref(false)

// Suggested starter quests mapped to the seeded category slugs.
const suggestions = [
  { name: 'Olahraga 20 menit', categoryId: 'olahraga', icon: 'ph-fill ph-barbell', attr: 'STR', difficulty: 'medium' },
  { name: 'Baca buku 10 halaman', categoryId: 'belajar', icon: 'ph-fill ph-brain', attr: 'INT', difficulty: 'easy' },
  { name: 'Saat teduh / renungan', categoryId: 'rohani', icon: 'ph-fill ph-hands-praying', attr: 'FAITH', difficulty: 'easy' },
  { name: 'Minum air 2 liter', categoryId: 'kesehatan', icon: 'ph-fill ph-drop', attr: 'VIT', difficulty: 'easy' },
  { name: 'Rapikan meja kerja', categoryId: 'kerja', icon: 'ph-fill ph-briefcase', attr: 'WIS', difficulty: 'easy' },
]
const picked = ref<number[]>([0, 2, 3]) // sensible defaults

function togglePick(i: number) {
  picked.value = picked.value.includes(i) ? picked.value.filter((x) => x !== i) : [...picked.value, i]
}

async function finish(skip = false) {
  saving.value = true
  try {
    if (!skip) {
      for (const i of picked.value) {
        const s = suggestions[i]
        await useApi()('/quests', {
          method: 'POST',
          body: { name: s.name, categoryId: s.categoryId, frequency: 'daily', difficulty: s.difficulty },
        })
      }
    }
    await auth.completeOnboarding()
    show(t('onboarding.welcomeToast'), 'ph-fill ph-lightning')
    await navigateTo('/')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="fx col" style="gap:0;width:100%">
    <!-- step 0: welcome -->
    <div v-if="step === 0" class="fx col ac" style="text-align:center">
      <div class="logo" style="margin-bottom:18px;width:64px;height:64px;font-size:32px"><i class="ph-fill ph-lightning" /></div>
      <div class="h" style="font-size:24px">{{ t('onboarding.hi', { name: auth.firstName }) }}</div>
      <p class="mut" style="font-size:13.5px;line-height:1.6;margin:12px 0 24px;max-width:300px">
        {{ t('onboarding.intro') }}
      </p>
      <div class="card pad" style="width:100%;text-align:left;margin-bottom:24px">
        <div class="fx col gap12">
          <div class="fx ac gap12"><span class="qi" style="background:rgba(224,87,79,.14);color:var(--str)"><i class="ph-fill ph-barbell" /></span><span style="font:600 12.5px 'Plus Jakarta Sans'">{{ t('onboarding.strLine') }}</span></div>
          <div class="fx ac gap12"><span class="qi" style="background:rgba(58,123,213,.14);color:var(--int)"><i class="ph-fill ph-brain" /></span><span style="font:600 12.5px 'Plus Jakarta Sans'">{{ t('onboarding.intLine') }}</span></div>
          <div class="fx ac gap12"><span class="qi" style="background:rgba(200,138,28,.16);color:var(--faith)"><i class="ph-fill ph-hands-praying" /></span><span style="font:600 12.5px 'Plus Jakarta Sans'">{{ t('onboarding.faithLine') }}</span></div>
        </div>
      </div>
      <button class="btn" @click="step = 1">{{ t('onboarding.start') }} <i class="ph-bold ph-arrow-right" /></button>
    </div>

    <!-- step 1: pick starter quests -->
    <div v-else class="fx col" style="width:100%">
      <div class="h" style="font-size:22px">{{ t('onboarding.pickTitle') }}</div>
      <p class="mut" style="font-size:13px;margin:8px 0 18px">{{ t('onboarding.pickSub') }}</p>

      <div class="fx col gap10" style="margin-bottom:20px">
        <button
          v-for="(s, i) in suggestions" :key="i" class="quest" style="cursor:pointer;text-align:left"
          @click="togglePick(i)"
        >
          <span class="chk" :class="{ chkon: picked.includes(i) }"><i v-if="picked.includes(i)" class="ph-bold ph-check" /></span>
          <span class="qi" style="background:var(--psoft);color:var(--primary)"><i :class="s.icon" /></span>
          <span class="f1">
            <span style="display:block;font:700 13px 'Plus Jakarta Sans';color:var(--ink)">{{ s.name }}</span>
            <span class="mut" style="font-size:10.5px">→ {{ s.attr }}</span>
          </span>
        </button>
      </div>

      <button class="btn" :disabled="saving" @click="finish(false)">
        <i class="ph-bold ph-check" />{{ saving ? t('onboarding.preparing') : t('onboarding.createStart', { n: picked.length }) }}
      </button>
      <button class="btn btno" style="margin-top:10px" :disabled="saving" @click="finish(true)">{{ t('onboarding.skip') }}</button>
    </div>
  </div>
</template>
