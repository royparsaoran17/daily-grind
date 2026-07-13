<script setup lang="ts">
const { isDark, toggle } = useTheme()
const { t } = useI18n()
</script>

<template>
  <button class="themebtn" type="button" @click="toggle">
    <i :class="isDark ? 'ph-fill ph-sun' : 'ph-fill ph-moon-stars'" />
    {{ isDark ? t('theme.light') : t('theme.dark') }}
  </button>
</template>
