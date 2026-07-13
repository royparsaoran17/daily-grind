<script setup lang="ts">
interface Verse { verse: number; textId: string; textEn?: string; meaning?: string }
interface BookEntry { id: string; name: string; ordinal: number; testament: 'OT' | 'NT'; chapters: number[] }
interface Chapter { bookId: string; book: string; chapter: number; verses: Verse[] }
interface Mark { bookId: string; chapter: number; verse: number; kind: string }
interface BookmarkItem { bookId: string; book: string; chapter: number; verse: number; text: string }

const { show } = useToast()
const { t, locale } = useI18n()

const data = ref<Chapter | null>(null)
const books = ref<BookEntry[]>([])
const pending = ref(true)
const pickerOpen = ref(false)
const pickBook = ref<BookEntry | null>(null)
const activeMeaning = ref<Verse | null>(null)

const selectedVerse = ref<number | null>(null)
const highlights = ref<number[]>([])
const bookmarks = ref<number[]>([])

const savedOpen = ref(false)
const savedList = ref<BookmarkItem[]>([])

async function loadMarks() {
  if (!data.value) return
  const marks = await useApi()<Mark[]>(`/bible/marks?book_id=${data.value.bookId}&chapter=${data.value.chapter}`)
  highlights.value = marks.filter((m) => m.kind === 'highlight').map((m) => m.verse)
  bookmarks.value = marks.filter((m) => m.kind === 'bookmark').map((m) => m.verse)
}

async function loadChapter(bookId?: string, chapter?: number) {
  pending.value = true
  pickerOpen.value = false
  selectedVerse.value = null
  try {
    const q = bookId ? `?bookId=${encodeURIComponent(bookId)}&chapter=${chapter}` : ''
    data.value = await useApi()<Chapter>(`/bible${q}`)
    activeMeaning.value = data.value.verses.find((v) => v.meaning) ?? null
    await loadMarks()
  } finally {
    pending.value = false
  }
}

function openPicker() {
  pickerOpen.value = !pickerOpen.value
  savedOpen.value = false
  pickBook.value = books.value.find((b) => b.id === data.value?.bookId) ?? books.value[0] ?? null
}

async function openSaved() {
  savedOpen.value = !savedOpen.value
  pickerOpen.value = false
  if (savedOpen.value) savedList.value = await useApi()<BookmarkItem[]>('/bible/bookmarks')
}

onMounted(async () => {
  books.value = await useApi()<BookEntry[]>('/bible/books')
  await loadChapter()
})

function verseText(v: Verse): string {
  return locale.value === 'en' && v.textEn ? v.textEn : v.textId
}
const selected = computed(() => data.value?.verses.find((v) => v.verse === selectedVerse.value) ?? null)

async function toggleMark(kind: 'highlight' | 'bookmark') {
  if (!data.value || selectedVerse.value == null) {
    show(t('bible.tapFirst'), 'ph-fill ph-hand-tap')
    return
  }
  const verse = selectedVerse.value
  const res = await useApi()<{ marked: boolean }>('/bible/marks', {
    method: 'POST',
    body: { bookId: data.value.bookId, chapter: data.value.chapter, verse, kind },
  })
  const arr = kind === 'highlight' ? highlights : bookmarks
  arr.value = res.marked ? [...arr.value, verse] : arr.value.filter((v) => v !== verse)
  show(
    res.marked ? (kind === 'highlight' ? t('bible.verseMarked') : t('bible.verseSaved')) : t('bible.markRemoved'),
    kind === 'highlight' ? 'ph-fill ph-highlighter' : 'ph-fill ph-bookmark-simple',
  )
}

async function shareVerse() {
  if (!data.value || !selected.value) {
    show(t('bible.tapFirst'), 'ph-fill ph-hand-tap')
    return
  }
  const ref = `${data.value.book} ${data.value.chapter}:${selected.value.verse}`
  const text = `"${verseText(selected.value)}" — ${ref}`
  if (import.meta.client && navigator.share) {
    try { await navigator.share({ text }) } catch { /* dismissed */ }
  } else if (import.meta.client) {
    await navigator.clipboard?.writeText(text)
    show(t('bible.verseCopied'), 'ph-fill ph-copy')
  }
}

function showMeaning() {
  if (selected.value?.meaning) activeMeaning.value = selected.value
  else show(t('bible.noMeaning'), 'ph-fill ph-sparkle')
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac jb">
      <span class="h">{{ t('bible.title') }}</span>
    </div>

    <div class="fx ac jb">
      <button class="chipv chip" style="font-size:12.5px;padding:8px 13px;border:none" @click="openPicker">
        <i class="ph-fill ph-book-open" />{{ data ? `${data.book} ${data.chapter}` : '…' }}
        <i :class="pickerOpen ? 'ph-bold ph-caret-up' : 'ph-bold ph-caret-down'" />
      </button>
      <button class="chip chipa" style="border:none" @click="openSaved"><i class="ph-fill ph-bookmark-simple" />{{ t('bible.saved') }}</button>
    </div>

    <!-- quick links -->
    <div class="fx gap8" style="overflow-x:auto;padding-bottom:2px">
      <NuxtLink to="/plans" class="chip" style="flex:none"><i class="ph ph-list-checks" />{{ t('plans.title') }}</NuxtLink>
      <NuxtLink to="/prayers" class="chip" style="flex:none"><i class="ph ph-hands-praying" />{{ t('prayers.title') }}</NuxtLink>
      <NuxtLink to="/devotional" class="chip chipa" style="flex:none"><i class="ph-fill ph-sun-horizon" />{{ t('devotional.title') }}</NuxtLink>
      <NuxtLink to="/journal" class="chip" style="flex:none"><i class="ph ph-notebook" />{{ t('bible.journal') }}</NuxtLink>
    </div>

    <!-- saved / bookmarks panel -->
    <div v-if="savedOpen" class="card pad">
      <span class="tny" style="display:block;margin-bottom:10px">{{ t('bible.savedVerses') }}</span>
      <div v-if="!savedList.length" class="mut" style="font-size:12px">{{ t('bible.noSaved') }}</div>
      <div v-else class="fx col gap10">
        <button
          v-for="b in savedList" :key="`${b.bookId}-${b.chapter}-${b.verse}`"
          style="text-align:left;background:none;border:none;padding:0" @click="loadChapter(b.bookId, b.chapter)"
        >
          <span style="font:700 12px 'Space Grotesk';color:var(--primary)">{{ b.book }} {{ b.chapter }}:{{ b.verse }}</span>
          <span class="mut" style="display:block;font-size:11.5px;line-height:1.5">{{ b.text.slice(0, 80) }}…</span>
        </button>
      </div>
    </div>

    <!-- book / chapter picker -->
    <div v-if="pickerOpen" class="card pad">
      <span class="tny" style="display:block;margin-bottom:10px">{{ t('bible.chooseBook') }}</span>
      <div class="fx wrap gap8" style="max-height:132px;overflow:auto">
        <button
          v-for="b in books" :key="b.id" class="chip"
          :class="{ chipv: pickBook?.id === b.id }" style="border:none" @click="pickBook = b"
        >{{ b.name }}</button>
      </div>
      <template v-if="pickBook">
        <span class="tny" style="display:block;margin:14px 0 10px">{{ t('bible.chooseChapter') }} — {{ pickBook.name }}</span>
        <div class="fx wrap gap8" style="max-height:150px;overflow:auto">
          <button
            v-for="c in pickBook.chapters" :key="c" class="chip"
            :class="{ chipv: data && pickBook.id === data.bookId && c === data.chapter }"
            style="border:none;min-width:38px;justify-content:center" @click="loadChapter(pickBook.id, c)"
          >{{ c }}</button>
        </div>
      </template>
    </div>

    <div v-if="pending" class="spinner" />
    <template v-else-if="data">
      <div class="card pad" style="padding:20px">
        <p
          v-for="(v, i) in data.verses" :key="v.verse"
          class="verse" style="cursor:pointer;border-radius:6px;transition:background .15s"
          :style="{
            margin: i === data.verses.length - 1 ? '0' : '0 0 14px',
            outline: selectedVerse === v.verse ? '2px solid var(--primary)' : 'none',
            outlineOffset: '3px',
          }"
          @click="selectedVerse = selectedVerse === v.verse ? null : v.verse"
        >
          <span class="vnum">{{ v.verse }}</span
          ><span :class="{ hl: highlights.includes(v.verse) }">{{ verseText(v) }}</span>
          <i v-if="bookmarks.includes(v.verse)" class="ph-fill ph-bookmark-simple" style="color:var(--primary);font-size:12px;margin-left:4px" />
        </p>
        <p v-if="!data.verses.length" class="mut" style="font-size:12.5px;margin:0">{{ t('bible.notAvailable') }}</p>
      </div>

      <div v-if="activeMeaning" class="card pad meaning">
        <div class="fx ac gap8" style="margin-bottom:9px">
          <i class="ph-fill ph-sparkle" style="color:var(--primary);font-size:17px" />
          <span style="font:700 12px 'Space Grotesk';color:var(--pink)">{{ t('bible.meaning') }} {{ activeMeaning.verse }}</span>
        </div>
        <p style="font:400 12.5px/1.65 'Plus Jakarta Sans';color:var(--ink);margin:0">{{ activeMeaning.meaning }}</p>
      </div>

      <div class="actbar" style="margin-top:auto">
        <button class="actbtn" :style="{ opacity: selectedVerse ? 1 : 0.5 }" @click="toggleMark('highlight')">
          <i class="ph-fill ph-highlighter" />{{ t('bible.mark') }}
        </button>
        <button class="actbtn" :style="{ opacity: selectedVerse ? 1 : 0.5 }" @click="toggleMark('bookmark')">
          <i class="ph ph-bookmark-simple" />{{ t('bible.save') }}
        </button>
        <button class="actbtn" :style="{ opacity: selectedVerse ? 1 : 0.5 }" @click="showMeaning">
          <i class="ph ph-sparkle" />{{ t('bible.meaning').split(' ')[0] }}
        </button>
        <button class="actbtn" :style="{ opacity: selectedVerse ? 1 : 0.5 }" @click="shareVerse">
          <i class="ph ph-share-network" />{{ t('bible.share') }}
        </button>
      </div>
      <p class="mut" style="text-align:center;font-size:10.5px;margin:0">
        {{ selectedVerse ? t('bible.selected', { n: selectedVerse }) : t('bible.tapHint') }}
      </p>
    </template>
  </div>
</template>
