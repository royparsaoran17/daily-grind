<script setup lang="ts">
import { timeAgo } from '~/utils/format'

interface Comment { id: string; author: string; body: string }
interface Post {
  id: string; author: string; authorLevel: number; body: string
  photoUrl?: string; badge?: string; likes: number; likedByMe: boolean
  comments: Comment[]; createdAt: string
}

const auth = useAuthStore()
const { show } = useToast()

const posts = ref<Post[]>([])
const pending = ref(true)
const draft = ref('')
const replyDraft = reactive<Record<string, string>>({})

onMounted(async () => {
  try { posts.value = await useApi()<Post[]>('/feed') } finally { pending.value = false }
})

async function publish() {
  const body = draft.value.trim()
  if (!body) return
  draft.value = ''
  await useApi()('/feed', { method: 'POST', body: { body } })
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
      <div class="avatar" style="width:40px;height:40px">{{ auth.initial }}</div>
      <div class="f1">
        <textarea v-model="draft" class="compin" rows="2" placeholder="Bagikan progres atau statusmu…" />
        <div class="fx ac jb" style="margin-top:8px">
          <div class="fx gap8">
            <span class="pact" style="width:32px;height:32px;justify-content:center;border-radius:10px;background:var(--soft)"><i class="ph-fill ph-image" style="color:var(--primary)" /></span>
            <span class="pact" style="width:32px;height:32px;justify-content:center;border-radius:10px;background:var(--soft)"><i class="ph-fill ph-chart-line-up" style="color:var(--primary)" /></span>
            <span class="pact" style="width:32px;height:32px;justify-content:center;border-radius:10px;background:var(--soft)"><i class="ph-fill ph-medal" style="color:var(--primary)" /></span>
          </div>
          <button class="btn" style="width:auto;padding:8px 16px;font-size:12.5px" :disabled="!draft.trim()" @click="publish">Bagikan</button>
        </div>
      </div>
    </div>

    <div v-if="pending" class="spinner" />

    <div v-for="p in posts" :key="p.id" class="post">
      <div class="fx ac" style="gap:11px">
        <div class="avatar" style="width:40px;height:40px">{{ p.author[0] }}</div>
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
          <div class="avatar" style="width:28px;height:28px;font-size:11px">{{ c.author[0] }}</div>
          <div class="replyt">
            <span style="font:700 11.5px 'Plus Jakarta Sans';color:var(--ink)">{{ c.author }}</span>
            <span style="font:400 12px 'Plus Jakarta Sans';color:var(--ink)"> {{ c.body }}</span>
          </div>
        </div>
      </div>

      <div class="replyin">
        <div class="avatar" style="width:26px;height:26px;font-size:10px">{{ auth.initial }}</div>
        <input v-model="replyDraft[p.id]" type="text" placeholder="Tulis balasan…" @keyup.enter="reply(p)">
        <button class="fx ac jc" style="width:32px;height:32px;border-radius:50%;background:var(--primary);color:#fff;border:none" @click="reply(p)">
          <i class="ph-fill ph-paper-plane-right" />
        </button>
      </div>
    </div>
  </div>
</template>
