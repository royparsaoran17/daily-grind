<script setup lang="ts">
interface Entry {
  id: string; date: string; title: string; body: string
  mood?: string; prompt?: string; updatedAt: string
}

const route = useRoute()
const { show } = useToast()

const MOODS = [
  { key: 'bersyukur', label: 'Bersyukur', icon: 'ph-fill ph-hands-praying' },
  { key: 'senang', label: 'Senang', icon: 'ph-fill ph-smiley' },
  { key: 'biasa', label: 'Biasa', icon: 'ph-fill ph-smiley-meh' },
  { key: 'lelah', label: 'Lelah', icon: 'ph-fill ph-battery-low' },
  { key: 'sedih', label: 'Sedih', icon: 'ph-fill ph-cloud-rain' },
]

const entries = ref<Entry[]>([])
const pending = ref(true)
const saving = ref(false)

// The editor always targets a specific date; "today" by default.
const selectedDate = ref<string>('today')
const title = ref('')
const body = ref('')
const mood = ref('')
const prompt = ref('')

const todayLabel = computed(() =>
  new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long' }),
)
const editingLabel = computed(() =>
  selectedDate.value === 'today'
    ? todayLabel.value
    : new Date(selectedDate.value).toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }),
)
const canSave = computed(() => (title.value.trim() || body.value.trim()) && !saving.value)

async function loadEntry(date: string) {
  const e = await useApi()<Entry | null>(`/journal/${date}`)
  title.value = e?.title ?? ''
  body.value = e?.body ?? ''
  mood.value = e?.mood ?? ''
  prompt.value = e?.prompt ?? (date === 'today' ? (route.query.prompt as string) ?? '' : '')
}

async function refresh() {
  entries.value = await useApi()<Entry[]>('/journal')
}

onMounted(async () => {
  try {
    await Promise.all([loadEntry('today'), refresh()])
  } finally {
    pending.value = false
  }
})

async function save() {
  saving.value = true
  try {
    await useApi()(`/journal/${selectedDate.value}`, {
      method: 'PUT',
      body: { title: title.value, body: body.value, mood: mood.value, prompt: prompt.value },
    })
    show('Jurnal tersimpan ✍️', 'ph-fill ph-notebook')
    await refresh()
  } catch (e: any) {
    show(e?.data?.error ?? 'Gagal menyimpan.', 'ph-fill ph-warning')
  } finally {
    saving.value = false
  }
}

function edit(e: Entry) {
  selectedDate.value = e.date
  title.value = e.title
  body.value = e.body
  mood.value = e.mood ?? ''
  prompt.value = e.prompt ?? ''
  if (import.meta.client) window.scrollTo({ top: 0, behavior: 'smooth' })
}

function newToday() {
  selectedDate.value = 'today'
  title.value = ''
  body.value = ''
  mood.value = ''
  prompt.value = (route.query.prompt as string) ?? ''
}

async function remove(e: Entry) {
  await useApi()(`/journal/${e.date}`, { method: 'DELETE' })
  if (e.date === selectedDate.value) newToday()
  await refresh()
  show('Jurnal dihapus', 'ph-fill ph-trash')
}

function moodIcon(key?: string) {
  return MOODS.find((m) => m.key === key)?.icon ?? 'ph-fill ph-note'
}
function preview(text: string) {
  return text.length > 90 ? text.slice(0, 90) + '…' : text
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac gap12">
      <button class="backbtn" @click="navigateTo('/bible')"><i class="ph-bold ph-caret-left" /></button>
      <span class="h">Jurnal Harian</span>
    </div>

    <!-- editor -->
    <div class="card pad fx col gap12">
      <div class="fx ac jb">
        <span class="sec"><i class="ph-fill ph-calendar-check" style="color:var(--primary)" /> {{ editingLabel }}</span>
        <button v-if="selectedDate !== 'today'" class="chip" @click="newToday">
          <i class="ph ph-plus" />Hari ini
        </button>
      </div>

      <input v-model="title" class="field" style="font:700 15px 'Space Grotesk';color:var(--ink)"
             placeholder="Judul (mis. Bersyukur hari ini)">

      <textarea v-model="body" class="compin" rows="6"
                style="border:1.5px solid var(--fline);border-radius:14px;padding:13px 15px"
                placeholder="Tulis refleksi, doa, atau apa pun yang kamu syukuri hari ini…" />

      <div v-if="prompt" class="mapbox">
        <i class="ph-fill ph-book-open" style="color:var(--primary);font-size:16px" />
        <span style="font:600 12px 'Plus Jakarta Sans';color:var(--pink)">Terinspirasi dari {{ prompt }}</span>
      </div>

      <div>
        <span class="flabel">Perasaan</span>
        <div class="fx wrap gap8">
          <button
            v-for="m in MOODS" :key="m.key" class="chip"
            :class="{ chipv: mood === m.key }" style="border:none"
            @click="mood = mood === m.key ? '' : m.key"
          >
            <i :class="m.icon" />{{ m.label }}
          </button>
        </div>
      </div>

      <button class="btn" :disabled="!canSave" @click="save">
        <i class="ph-bold ph-floppy-disk" />{{ saving ? 'Menyimpan…' : 'Simpan Jurnal' }}
      </button>
    </div>

    <!-- history -->
    <div class="fx ac jb">
      <span class="sec">Riwayat</span>
      <span class="tny">{{ entries.length }} catatan</span>
    </div>

    <div v-if="pending" class="spinner" />
    <div v-else-if="!entries.length" class="mut" style="text-align:center;font-size:12.5px;padding:20px">
      Belum ada catatan. Mulai tulis jurnal pertamamu di atas.
    </div>
    <div v-else class="fx col gap10">
      <div v-for="e in entries" :key="e.id" class="quest" style="align-items:flex-start">
        <span class="qi" style="background:var(--psoft);color:var(--primary)"><i :class="moodIcon(e.mood)" /></span>
        <button class="f1" style="text-align:left;background:none;border:none;padding:0" @click="edit(e)">
          <span style="display:block;font:700 13.5px 'Plus Jakarta Sans';color:var(--ink)">
            {{ e.title || '(tanpa judul)' }}
          </span>
          <span class="mut" style="display:block;font-size:10.5px;margin:2px 0 4px">
            {{ new Date(e.date).toLocaleDateString('id-ID', { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' }) }}
          </span>
          <span class="mut" style="font-size:11.5px;line-height:1.5">{{ preview(e.body) }}</span>
        </button>
        <button class="pact" style="padding:0" @click.stop="remove(e)"><i class="ph ph-trash" style="font-size:17px" /></button>
      </div>
    </div>
  </div>
</template>
