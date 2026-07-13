<script setup lang="ts">
import { timeAgo } from '~/utils/format'

interface Comment { id: string; author: string; authorAvatar?: string; body: string }
interface Post {
  id: string; author: string; authorAvatar?: string; authorLevel: number; body: string
  photoUrl?: string; badge?: string; likes: number; likedByMe: boolean
  comments: Comment[]; createdAt: string
}

const auth = useAuthStore()
const { show } = useToast()
const { uploadImage } = useUpload()

const posts = ref<Post[]>([])
const pending = ref(true)
const draft = ref('')
const draftPhoto = ref('')
const uploadingPhoto = ref(false)
const photoInput = ref<HTMLInputElement | null>(null)
const replyDraft = reactive<Record<string, string>>({})

onMounted(async () => {
  try { posts.value = await useApi()<Post[]>('/feed') } finally { pending.value = false }
})

async function onPickPhoto(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  uploadingPhoto.value = true
  try {
    draftPhoto.value = await uploadImage(file, 'post')
  } catch {
    show('Gagal mengunggah foto.', 'ph-fill ph-warning')
  } finally {
    uploadingPhoto.value = false
    if (photoInput.value) photoInput.value.value = ''
  }
}

async function publish() {
  const body = draft.value.trim()
  if (!body && !draftPhoto.value) return
  const photoUrl = draftPhoto.value
  draft.value = ''
  draftPhoto.value = ''
  await useApi()('/feed', { method: 'POST', body: { body, photoUrl } })
  posts.value = await useApi()<Post[]>('/feed')
  show('Postingan dibagikan! ✨', 'ph-fill ph-paper-plane-right')
}

async function toggleLike(p: Post) {
  const prev = p.likedByMe
  p.likedByMe = !prev
  p.likes += prev ? -1 : 1
  try {
    const res = await useApi()<{ liked: boolean; likes: number }>(`/feed/${p.id}/like`, { method: 'POST' })
    p.likedByMe = res.liked
    p.likes = res.likes
  } catch {
    p.likedByMe = prev
    p.likes += prev ? 1 : -1
  }
}

async function reply(p: Post) {
  const body = (replyDraft[p.id] ?? '').trim()
  if (!body) return
  replyDraft[p.id] = ''
  const c = await useApi()<Comment>(`/feed/${p.id}/comments`, { method: 'POST', body: { body } })
  p.comments.push(c)
}
</script>

<template>
  <div class="fx col gap16">
    <div class="fx ac jb">
      <span class="h">Aktivitas</span>
      <i class="ph ph-bell mut" style="font-size:20px" />
    </div>

    <!-- composer -->
    <div class="composer">
      <Avatar :url="auth.user?.avatarUrl" :name="auth.user?.name" :size="40" />
      <div class="f1">
        <textarea v-model="draft" class="compin" rows="2" placeholder="Bagikan progres atau statusmu…" />
        <div v-if="draftPhoto || uploadingPhoto" style="position:relative;margin:8px 0">
          <div v-if="uploadingPhoto" class="imgph" style="height:120px">Mengunggah…</div>
          <template v-else>
            <img :src="draftPhoto" class="postphoto" style="height:150px;margin:0" alt="pratinjau">
            <button
              style="position:absolute;top:8px;right:8px;width:28px;height:28px;border-radius:50%;background:rgba(0,0,0,.6);color:#fff;border:none;display:flex;align-items:center;justify-content:center"
              @click="draftPhoto = ''"
            ><i class="ph-bold ph-x" /></button>
          </template>
        </div>
        <div class="fx ac jb" style="margin-top:8px">
          <div class="fx gap8">
            <button class="pact" style="width:32px;height:32px;justify-content:center;border-radius:10px;background:var(--soft);border:none" :disabled="uploadingPhoto" @click="photoInput?.click()">
              <i class="ph-fill ph-image" style="color:var(--primary)" />
            </button>
            <span class="pact" style="width:32px;height:32px;justify-content:center;border-radius:10px;background:var(--soft)"><i class="ph-fill ph-chart-line-up" style="color:var(--primary)" /></span>
            <span class="pact" style="width:32px;height:32px;justify-content:center;border-radius:10px;background:var(--soft)"><i class="ph-fill ph-medal" style="color:var(--primary)" /></span>
          </div>
          <button class="btn" style="width:auto;padding:8px 16px;font-size:12.5px" :disabled="(!draft.trim() && !draftPhoto) || uploadingPhoto" @click="publish">Bagikan</button>
        </div>
        <input ref="photoInput" type="file" accept="image/*" style="display:none" @change="onPickPhoto">
      </div>
    </div>

    <div v-if="pending" class="spinner" />

    <div v-for="p in posts" :key="p.id" class="post">
      <div class="fx ac" style="gap:11px">
        <Avatar :url="p.authorAvatar" :name="p.author" :size="40" />
        <div class="f1">
          <div style="font:700 13.5px 'Plus Jakarta Sans';color:var(--ink)">{{ p.author }}</div>
          <div class="mut" style="font-size:10.5px">Lvl {{ p.authorLevel }} · {{ timeAgo(p.createdAt) }}</div>
        </div>
        <span v-if="p.badge" class="chip chipv"><i class="ph-fill ph-arrow-up" />{{ p.badge }}</span>
      </div>

      <p style="font:400 13px/1.55 'Plus Jakarta Sans';color:var(--ink);margin:12px 0 0">{{ p.body }}</p>
      <img v-if="p.photoUrl" :src="p.photoUrl" class="postphoto" alt="foto">

      <div class="postbar">
        <button class="pact" :class="{ on: p.likedByMe }" @click="toggleLike(p)">
          <i :class="p.likedByMe ? 'ph-fill ph-heart' : 'ph ph-heart'" />{{ p.likes }}
        </button>
        <span class="pact"><i class="ph ph-chat-circle" />{{ p.comments.length }} balasan</span>
        <span class="pact"><i class="ph ph-share-network" />Bagikan</span>
      </div>

      <div v-if="p.comments.length" class="replybox">
        <div v-for="c in p.comments" :key="c.id" class="reply">
          <Avatar :url="c.authorAvatar" :name="c.author" :size="28" :font="11" />
          <div class="replyt">
            <span style="font:700 11.5px 'Plus Jakarta Sans';color:var(--ink)">{{ c.author }}</span>
            <span style="font:400 12px 'Plus Jakarta Sans';color:var(--ink)"> {{ c.body }}</span>
          </div>
        </div>
      </div>

      <div class="replyin">
        <Avatar :url="auth.user?.avatarUrl" :name="auth.user?.name" :size="26" :font="10" />
        <input v-model="replyDraft[p.id]" type="text" placeholder="Tulis balasan…" @keyup.enter="reply(p)">
        <button class="fx ac jc" style="width:32px;height:32px;border-radius:50%;background:var(--primary);color:#fff;border:none" @click="reply(p)">
          <i class="ph-fill ph-paper-plane-right" />
        </button>
      </div>
    </div>
  </div>
</template>
