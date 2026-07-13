<script setup lang="ts">
const route = useRoute()
const { t } = useI18n()

const tabs = computed(() => [
  { to: '/', label: t('nav.home'), icon: 'ph-house-simple', match: ['/'] },
  { to: '/quests', label: t('nav.quests'), icon: 'ph-target', match: ['/quests'] },
  { to: '/friends', label: t('nav.friends'), icon: 'ph-users-three', match: ['/friends', '/feed'] },
  { to: '/bible', label: t('nav.bible'), icon: 'ph-book-bookmark', match: ['/bible', '/devotional'] },
  { to: '/profile', label: t('nav.profile'), icon: 'ph-user-circle', match: ['/profile'] },
])

function active(match: string[]): boolean {
  return match.some((m) => (m === '/' ? route.path === '/' : route.path.startsWith(m)))
}
</script>

<template>
  <nav class="tabs">
    <NuxtLink v-for="tab in tabs" :key="tab.to" :to="tab.to" class="tab" :class="{ tabon: active(tab.match) }">
      <i :class="active(tab.match) ? `ph-fill ${tab.icon}` : `ph ${tab.icon}`" />
      {{ tab.label }}
    </NuxtLink>
  </nav>
</template>
