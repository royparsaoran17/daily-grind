<script setup lang="ts">
interface Verse { verse: number; textId: string; textEn?: string; meaning?: string }
interface BookEntry { id: string; name: string; ordinal: number; testament: 'OT' | 'NT'; chapters: number[] }
interface Chapter { bookId: string; book: string; chapter: number; verses: Verse[] }

const lang = ref<'EN' | 'ID'>('ID')
const data = ref<Chapter | null>(null)
const books = ref<BookEntry[]>([])
const pending = ref(true)
const pickerOpen = ref(false)
const pickBook = ref<BookEntry | null>(null) // book being expanded in the picker
const activeMeaning = ref<Verse | null>(null)

async function loadChapter(bookId?: string, chapter?: number) {
  pending.value = true
  pickerOpen.value = false
  try {
    const q = bookId ? `?bookId=${encodeURIComponent(bookId)}&chapter=${chapter}` : ''
    data.value = await useApi()<Chapter>(`/bible${q}`)
    activeMeaning.value = data.value.verses.find((v) => v.meaning) ?? null
  } finally {
    pending.value = false
  }
}

function openPicker() {
  pickerOpen.value = !pickerOpen.value
  // Default the expanded book to the one currently open.
  pickBook.value = books.value.find((b) => b.id === data.value?.bookId) ?? books.value[0] ?? null
}

onMounted(async () => {
  books.value = await useApi()<BookEntry[]>('/bible/books')
  await loadChapter()
})

function verseText(v: Verse): string {
  return lang.value === 'EN' && v.textEn ? v.textEn : v.textId
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac jb">
      <span class="h">Alkitab</span>
      <span class="lang">
        <button class="langi" :class="{ langon: lang === 'EN' }" @click="lang = 'EN'">EN</button>
        <button class="langi" :class="{ langon: lang === 'ID' }" @click="lang = 'ID'">ID</button>
      </span>
    </div>

    <div class="fx ac jb">
      <button class="chipv chip" style="font-size:12.5px;padding:8px 13px;border:none" @click="openPicker">
        <i class="ph-fill ph-book-open" />{{ data ? `${data.book} ${data.chapter}` : '…' }}
        <i :class="pickerOpen ? 'ph-bold ph-caret-up' : 'ph-bold ph-caret-down'" />
      </button>
      <div class="fx gap8">
        <NuxtLink to="/journal" class="chip"><i class="ph ph-notebook" />Jurnal</NuxtLink>
        <NuxtLink to="/devotional" class="chip chipa"><i class="ph-fill ph-sun-horizon" />Renungan</NuxtLink>
      </div>
    </div>

    <!-- book / chapter picker -->
    <div v-if="pickerOpen" class="card pad">
      <span class="tny" style="display:block;margin-bottom:10px">Kitab</span>
      <div class="fx wrap gap8" style="max-height:132px;overflow:auto">
        <button
          v-for="b in books" :key="b.id" class="chip"
          :class="{ chipv: pickBook?.id === b.id }" style="border:none" @click="pickBook = b"
        >
          {{ b.name }}
        </button>
      </div>
      <template v-if="pickBook">
        <span class="tny" style="display:block;margin:14px 0 10px">Pasal — {{ pickBook.name }}</span>
        <div class="fx wrap gap8" style="max-height:150px;overflow:auto">
          <button
            v-for="c in pickBook.chapters" :key="c" class="chip"
            :class="{ chipv: data && pickBook.id === data.bookId && c === data.chapter }"
            style="border:none;min-width:38px;justify-content:center"
            @click="loadChapter(pickBook.id, c)"
          >
            {{ c }}
          </button>
        </div>
      </template>
    </div>

    <div v-if="pending" class="spinner" />
    <template v-else-if="data">
      <div class="card pad" style="padding:20px">
        <p
          v-for="(v, i) in data.verses" :key="v.verse" class="verse"
          :style="i === data.verses.length - 1 ? 'margin:0' : 'margin:0 0 14px'"
        >
          <span class="vnum">{{ v.verse }}</span>{{ verseText(v) }}
        </p>
        <p v-if="!data.verses.length" class="mut" style="font-size:12.5px;margin:0">Pasal ini belum tersedia.</p>
      </div>

      <div v-if="activeMeaning" class="card pad meaning">
        <div class="fx ac gap8" style="margin-bottom:9px">
          <i class="ph-fill ph-sparkle" style="color:var(--primary);font-size:17px" />
          <span style="font:700 12px 'Space Grotesk';color:var(--pink)">Makna ayat {{ activeMeaning.verse }}</span>
        </div>
        <p style="font:400 12.5px/1.65 'Plus Jakarta Sans';color:var(--ink);margin:0">{{ activeMeaning.meaning }}</p>
      </div>

      <div class="actbar" style="margin-top:auto">
        <button class="actbtn"><i class="ph-fill ph-highlighter" />Tandai</button>
        <button class="actbtn"><i class="ph ph-bookmark-simple" />Simpan</button>
        <NuxtLink class="actbtn" to="/journal"><i class="ph ph-notebook" />Jurnal</NuxtLink>
        <button class="actbtn"><i class="ph ph-share-network" />Bagikan</button>
      </div>
    </template>
  </div>
</template>
