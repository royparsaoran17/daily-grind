<script setup lang="ts">
const { isDark, toggle } = useTheme()
</script>

<template>
  <button class="themebtn" type="button" @click="toggle">
    <i :class="isDark ? 'ph-fill ph-sun' : 'ph-fill ph-moon-stars'" />
    {{ isDark ? 'Terang' : 'Gelap' }}
  </button>
</template>
