<script setup lang="ts">
// Shows a user's photo when available, else their initial. Used everywhere a
// user is represented (self, friends, post/comment authors).
const props = defineProps<{ url?: string | null; name?: string; size?: number; font?: number }>()
const initial = computed(() => (props.name?.trim()?.[0] ?? '?').toUpperCase())
const px = computed(() => props.size ?? 40)
</script>

<template>
  <div
    class="avatar"
    :style="{
      width: px + 'px',
      height: px + 'px',
      fontSize: (font ?? Math.round(px * 0.4)) + 'px',
      ...(url ? { backgroundImage: `url(${url})` } : {}),
    }"
  >
    <template v-if="!url">{{ initial }}</template>
  </div>
</template>
