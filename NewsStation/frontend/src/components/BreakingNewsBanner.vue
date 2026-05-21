<template>
  <div class="breaking-news-banner">
    <div class="banner-content">
      <div class="alert-icon">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
        </svg>
      </div>
      <div class="news-text">
        <span class="label">{{ t('breakingNews') }}</span>
        <span class="marquee">
          <span class="marquee-content">
            🚨 {{ t('alert') }} | {{ t('techSummit') }} | {{ t('aiBreakthrough') }} | {{ t('climateAgreement') }}
          </span>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { getTranslation } from '../i18n/languages'

defineProps({
  locale: {
    type: String,
    default: 'zh-CN'
  }
})

function t(key) {
  return getTranslation(this.locale, key)
}
</script>

<style scoped>
.breaking-news-banner {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.9) 0%, rgba(220, 38, 38, 0.9) 100%);
  backdrop-filter: blur(10px);
  padding: 10px 20px;
  box-shadow: 0 4px 20px rgba(239, 68, 68, 0.4);
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { box-shadow: 0 4px 20px rgba(239, 68, 68, 0.4); }
  50% { box-shadow: 0 4px 30px rgba(239, 68, 68, 0.6); }
}

.banner-content {
  display: flex;
  align-items: center;
  gap: 16px;
  max-width: 100%;
  overflow: hidden;
}

.alert-icon {
  flex-shrink: 0;
  color: white;
  animation: bounce 1s ease-in-out infinite;
}

@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-3px); }
}

.news-text {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  overflow: hidden;
}

.label {
  flex-shrink: 0;
  background: rgba(255, 255, 255, 0.2);
  color: white;
  padding: 3px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.marquee {
  flex: 1;
  overflow: hidden;
  white-space: nowrap;
}

.marquee-content {
  display: inline-block;
  color: white;
  font-size: 14px;
  font-weight: 500;
  animation: marquee 20s linear infinite;
}

@keyframes marquee {
  0% { transform: translateX(0); }
  100% { transform: translateX(-50%); }
}
</style>