<template>
  <div
    v-if="visible"
    class="info-window"
    :style="{
      left: position.x + 'px',
      top: position.y + 'px'
    }"
  >
    <div class="window-header">
      <h3 class="window-title">{{ news.title }}</h3>
      <button class="close-btn" @click="$emit('close')">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
        </svg>
      </button>
    </div>
    <div class="window-body">
      <p class="window-summary">{{ news.summary }}</p>
      <div class="window-meta">
        <span class="meta-item">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
            <path d="M12 8a2 2 0 10-4 0v4a2 2 0 004 0V8zM2 8a2 2 0 114 0v4a2 2 0 01-4 0V8zM20 8a2 2 0 10-4 0v4a2 2 0 004 0V8z" />
          </svg>
          {{ formatTime(news.publish_time) }}
        </span>
        <span class="meta-item">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-11a1 1 0 10-2 0v2H7a1 1 0 100 2h2v2a1 1 0 102 0v-2h2a1 1 0 100-2h-2V7z" clip-rule="evenodd" />
          </svg>
          {{ t('source') }}: {{ news.source }}
        </span>
        <span v-if="news.is_breaking" class="breaking-badge">
          <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
          {{ t('breaking') }}
        </span>
      </div>
    </div>
    <div class="window-pointer"></div>
  </div>
</template>

<script setup>
import { watch } from 'vue'
import { getTranslation } from '../i18n/languages'

const props = defineProps({
  visible: Boolean,
  news: Object,
  position: Object,
  locale: {
    type: String,
    default: 'zh-CN'
  }
})

defineEmits(['close'])

watch(() => props.news, (newNews) => {
  console.log('InfoWindow news updated:', newNews?.title || 'null')
}, { immediate: true })

watch(() => props.visible, (visible) => {
  console.log('InfoWindow visible:', visible)
})

function t(key) {
  return getTranslation(props.locale, key)
}

function formatTime(dateString) {
  const date = new Date(dateString)
  return date.toLocaleString(props.locale.startsWith('zh') ? 'zh-CN' : props.locale, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.info-window {
  position: absolute;
  z-index: 1000;
  background: rgba(255, 255, 255, 0.98);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 16px;
  width: 340px;
  box-shadow: 0 12px 48px rgba(0, 0, 0, 0.12),
              0 0 0 1px rgba(0, 0, 0, 0.04);
  transform: translate(-50%, -100%) translateY(-15px);
  overflow: hidden;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translate(-50%, -100%) translateY(-5px);
  }
  to {
    opacity: 1;
    transform: translate(-50%, -100%) translateY(-15px);
  }
}

.window-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 16px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
}

.window-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
  line-height: 1.4;
  flex: 1;
  padding-right: 12px;
}

.close-btn {
  flex-shrink: 0;
  background: rgba(0, 0, 0, 0.06);
  border: none;
  border-radius: 8px;
  padding: 6px;
  color: #6b7280;
  cursor: pointer;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(0, 0, 0, 0.1);
  color: #1f2937;
}

.window-body {
  padding: 16px;
}

.window-summary {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #4b5563;
  line-height: 1.6;
}

.window-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  align-items: center;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #6b7280;
  background: rgba(0, 0, 0, 0.04);
  padding: 4px 10px;
  border-radius: 12px;
}

.breaking-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  font-weight: 600;
  color: #ef4444;
  background: rgba(239, 68, 68, 0.08);
  padding: 4px 10px;
  border-radius: 12px;
}

.window-pointer {
  position: absolute;
  bottom: -8px;
  left: 50%;
  transform: translateX(-50%);
  width: 0;
  height: 0;
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
  border-top: 8px solid white;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.08));
}
</style>