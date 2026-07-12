<script setup lang="ts">
const route = useRoute()

const tabs = [
  { to: '/', label: 'Beranda', icon: 'ph-house-simple', match: ['/'] },
  { to: '/quests', label: 'Quest', icon: 'ph-target', match: ['/quests'] },
  { to: '/friends', label: 'Teman', icon: 'ph-users-three', match: ['/friends', '/feed'] },
  { to: '/bible', label: 'Alkitab', icon: 'ph-book-bookmark', match: ['/bible', '/devotional'] },
  { to: '/profile', label: 'Profil', icon: 'ph-user-circle', match: ['/profile'] },
]

function active(match: string[]): boolean {
  return match.some((m) => (m === '/' ? route.path === '/' : route.path.startsWith(m)))
}
</script>

<template>
  <nav class="tabs">
    <NuxtLink v-for="t in tabs" :key="t.to" :to="t.to" class="tab" :class="{ tabon: active(t.match) }">
      <i :class="active(t.match) ? `ph-fill ${t.icon}` : `ph ${t.icon}`" />
      {{ t.label }}
    </NuxtLink>
  </nav>
</template>
