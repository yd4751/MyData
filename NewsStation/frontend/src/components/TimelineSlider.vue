<template>
  <div class="timeline-container">
    <div class="timeline-header">
      <span class="timeline-label">{{ t('timeRange') }}</span>
      <div class="date-display">
        <span>{{ startDate }}</span>
        <span class="separator">~</span>
        <span>{{ endDate }}</span>
      </div>
    </div>
    <div class="timeline-track">
      <div class="track-progress" :style="{ width: progress + '%' }"></div>
      <input
        type="range"
        class="timeline-slider"
        min="0"
        max="100"
        v-model="progress"
      />
      <div class="track-markers">
        <span v-for="i in 10" :key="i" class="marker"></span>
      </div>
    </div>
    <div class="timeline-labels">
      <span>7{{ t('daysAgo') }}</span>
      <span>3{{ t('daysAgo') }}</span>
      <span>{{ t('today') }}</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { getTranslation } from '../i18n/languages'

const props = defineProps({
  locale: {
    type: String,
    default: 'zh-CN'
  }
})

const progress = ref(50)

function t(key) {
  return getTranslation(props.locale, key)
}

const startDate = computed(() => {
  const date = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000)
  return date.toLocaleDateString(props.locale.startsWith('zh') ? 'zh-CN' : props.locale, { month: 'short', day: 'numeric' })
})

const endDate = computed(() => {
  return new Date().toLocaleDateString(props.locale.startsWith('zh') ? 'zh-CN' : props.locale, { month: 'short', day: 'numeric' })
})
</script>

<style scoped>
.timeline-container {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 16px;
  padding: 20px 28px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
  z-index: 100;
  min-width: 400px;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.timeline-label {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.date-display {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: #6b7280;
}

.separator {
  color: #d1d5db;
}

.timeline-track {
  position: relative;
  height: 8px;
  background: #e5e7eb;
  border-radius: 4px;
  cursor: pointer;
}

.track-progress {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  background: linear-gradient(90deg, #3b82f6 0%, #8b5cf6 100%);
  border-radius: 4px;
  transition: width 0.15s ease-out;
}

.timeline-slider {
  position: absolute;
  left: 0;
  top: 50%;
  width: 100%;
  height: 8px;
  margin: 0;
  -webkit-appearance: none;
  appearance: none;
  background: transparent;
  cursor: pointer;
  transform: translateY(-50%);
  z-index: 10;
}

.timeline-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 22px;
  height: 22px;
  background: white;
  border: 3px solid #3b82f6;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
  transition: transform 0.2s, box-shadow 0.2s;
}

.timeline-slider::-webkit-slider-thumb:hover {
  transform: scale(1.15);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.timeline-slider::-moz-range-thumb {
  width: 22px;
  height: 22px;
  background: white;
  border: 3px solid #3b82f6;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
}

.track-markers {
  position: absolute;
  left: 0;
  right: 0;
  top: 50%;
  display: flex;
  justify-content: space-between;
  padding: 0 2px;
  transform: translateY(-50%);
}

.marker {
  width: 2px;
  height: 10px;
  background: #d1d5db;
  border-radius: 1px;
}

.timeline-labels {
  display: flex;
  justify-content: space-between;
  margin-top: 12px;
  font-size: 11px;
  color: #9ca3af;
}
</style>