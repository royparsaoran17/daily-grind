<script setup lang="ts">
interface Category { id: string; label: string; icon: string; attribute: string }

const { create } = useQuests()
const { show } = useToast()

const categories = ref<Category[]>([])
const name = ref('')
const categoryId = ref('')
const frequency = ref<'daily' | 'weekly' | 'monthly'>('daily')
const difficulty = ref<'easy' | 'medium' | 'hard'>('medium')
const reminder = ref('')
const weekday = ref(1) // 0=Minggu..6=Sabtu; default Senin
const dayOfMonth = ref(1) // 1..28
const saving = ref(false)
const error = ref('')

const WEEKDAYS = ['Min', 'Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab']

const ATTR_NAME: Record<string, string> = {
  str: 'STR — Kekuatan', vit: 'VIT — Vitalitas', int: 'INT — Kecerdasan',
  wis: 'WIS — Kebijaksanaan', faith: 'FAITH — Iman',
}
const rewardTable = { easy: [20, 8], medium: [40, 15], hard: [60, 25] } as const

const selected = computed(() => categories.value.find((c) => c.id === categoryId.value))
const reward = computed(() => rewardTable[difficulty.value])

onMounted(async () => {
  categories.value = await useApi()<Category[]>('/categories')
  if (categories.value.length) categoryId.value = categories.value[0].id
})

async function save() {
  error.value = ''
  if (!name.value.trim() || !categoryId.value) {
    error.value = 'Nama dan kategori wajib diisi.'
    return
  }
  saving.value = true
  try {
    await create({
      name: name.value.trim(),
      categoryId: categoryId.value,
      frequency: frequency.value,
      difficulty: difficulty.value,
      reminder: reminder.value.trim(),
      weekday: frequency.value === 'weekly' ? weekday.value : null,
      dayOfMonth: frequency.value === 'monthly' ? dayOfMonth.value : null,
    })
    show('Quest baru dibuat! 🎯', 'ph-fill ph-target')
    await navigateTo('/quests')
  } catch (e: any) {
    error.value = e?.data?.error ?? 'Gagal menyimpan quest.'
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac gap12">
      <button class="backbtn" @click="navigateTo('/quests')"><i class="ph-bold ph-caret-left" /></button>
      <span class="h">Buat Quest Baru</span>
    </div>

    <div>
      <span class="flabel">Nama quest</span>
      <div class="field">
        <i class="ph ph-pencil-simple" />
        <input v-model="name" type="text" placeholder="mis. Lari pagi 5 km">
      </div>
    </div>

    <div>
      <span class="flabel">Frekuensi</span>
      <div class="seg">
        <button class="segi" :class="{ segon: frequency === 'daily' }" @click="frequency = 'daily'">Harian</button>
        <button class="segi" :class="{ segon: frequency === 'weekly' }" @click="frequency = 'weekly'">Mingguan</button>
        <button class="segi" :class="{ segon: frequency === 'monthly' }" @click="frequency = 'monthly'">Bulanan</button>
      </div>

      <!-- weekly: pick a day of the week -->
      <div v-if="frequency === 'weekly'" style="margin-top:10px">
        <span class="flabel">Setiap hari</span>
        <div class="fx wrap gap8">
          <button
            v-for="(d, i) in WEEKDAYS" :key="i" class="chip"
            :class="{ chipv: weekday === i }" style="border:none;min-width:44px;justify-content:center"
            @click="weekday = i"
          >{{ d }}</button>
        </div>
      </div>

      <!-- monthly: pick a day of the month -->
      <div v-if="frequency === 'monthly'" style="margin-top:10px">
        <span class="flabel">Tiap tanggal</span>
        <div class="fx wrap gap8" style="max-height:120px;overflow:auto">
          <button
            v-for="n in 28" :key="n" class="chip"
            :class="{ chipv: dayOfMonth === n }" style="border:none;min-width:38px;justify-content:center"
            @click="dayOfMonth = n"
          >{{ n }}</button>
        </div>
        <p class="mut" style="font-size:10.5px;margin:6px 0 0">Maksimal tanggal 28 agar berlaku di semua bulan.</p>
      </div>
    </div>

    <div>
      <span class="flabel">Kategori</span>
      <div class="fx wrap gap8">
        <button
          v-for="c in categories" :key="c.id" class="chip"
          :class="{ chipv: c.id === categoryId }" @click="categoryId = c.id"
        >
          <i :class="c.id === categoryId ? `ph-fill ${c.icon}` : `ph ${c.icon}`" />{{ c.label }}
        </button>
      </div>
      <div v-if="selected" class="mapbox">
        <i class="ph-fill ph-arrow-bend-down-right" style="color:var(--primary);font-size:16px" />
        <span style="font:600 12px 'Plus Jakarta Sans';color:var(--pink)">
          Menaikkan skill <b>{{ ATTR_NAME[selected.attribute] }}</b>
        </span>
      </div>
    </div>

    <div>
      <span class="flabel">Tingkat kesulitan</span>
      <div class="seg">
        <button class="segi" :class="{ segon: difficulty === 'easy' }" @click="difficulty = 'easy'">Mudah</button>
        <button class="segi" :class="{ segon: difficulty === 'medium' }" @click="difficulty = 'medium'">Sedang</button>
        <button class="segi" :class="{ segon: difficulty === 'hard' }" @click="difficulty = 'hard'">Sulit</button>
      </div>
      <div class="fx ac jb" style="margin-top:10px">
        <span class="mut" style="font:600 11.5px 'Plus Jakarta Sans'">Hadiah per selesai</span>
        <div class="fx gap8">
          <span class="chipa pill"><i class="ph-fill ph-lightning" />+{{ reward[0] }} EXP</span>
          <span class="chipa pill"><i class="ph-fill ph-coins" />+{{ reward[1] }}</span>
        </div>
      </div>
    </div>

    <div>
      <span class="flabel">Pengingat</span>
      <div class="field">
        <i class="ph ph-bell" />
        <input v-model="reminder" type="text" placeholder="mis. Setiap hari, 05.30">
        <i class="ph ph-caret-right" />
      </div>
    </div>

    <p v-if="error" style="color:var(--str);font:600 12px 'Plus Jakarta Sans';margin:0;text-align:center">{{ error }}</p>

    <button class="btn" :disabled="saving" @click="save">
      <i class="ph-bold ph-check" />{{ saving ? 'Menyimpan…' : 'Simpan Quest' }}
    </button>
  </div>
</template>
