<template>
  <div class="app-container">
    <SkeletonLoader v-if="isLoading" />
    <BreakingNewsBanner 
      :locale="currentLocale" 
      :breakingNews="breakingNewsList"
    />
    <MapContainer 
      :initial-position="initialPosition" 
      :locale="currentLocale"
      :is-loading="isLoading"
      :news-data="newsList"
      @bounds-change="handleBoundsChange"
    />
    <TimelineSlider :locale="currentLocale" />
    <div class="logo">
      <svg class="logo-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10" />
        <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" />
        <path d="M2 12h20" />
      </svg>
      <span class="logo-text">GeoNews</span>
    </div>
    <div class="stats-panel">
      <div class="stat-item">
        <span class="stat-value">{{ newsCount }}</span>
        <span class="stat-label">{{ t('newsCount') }}</span>
      </div>
      <div class="stat-item">
        <span class="stat-value breaking">{{ breakingCount }}</span>
        <span class="stat-label">{{ t('breakingCount') }}</span>
      </div>
    </div>
    <div class="language-selector" v-if="!isLoading">
      <select v-model="currentLocale" class="language-select">
        <option v-for="(lang, code) in languages" :key="code" :value="code">
          {{ lang.name }}
        </option>
      </select>
      <div class="location-indicator" v-if="locationInfo">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" clip-rule="evenodd" />
        </svg>
        <span>{{ locationInfo.city }}, {{ locationInfo.countryCode }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import BreakingNewsBanner from './components/BreakingNewsBanner.vue'
import MapContainer from './components/MapContainer.vue'
import TimelineSlider from './components/TimelineSlider.vue'
import SkeletonLoader from './components/SkeletonLoader.vue'
import { getGeoLocation, getLocaleFromCountry } from './utils/geoService'
import { languages, getTranslation } from './i18n/languages'
import { SSEService } from './utils/sseService'

const newsList = ref([])
const newsCount = computed(() => newsList.value.length)
const breakingCount = computed(() => newsList.value.filter(item => item.is_breaking).length)

const isLoading = ref(true)
const currentLocale = ref('zh-CN')
const initialPosition = ref({
  lat: 39.9042,
  lng: 116.4074
})
const locationInfo = ref(null)
const breakingNewsList = ref([])

let sseService = null

function t(key) {
  return getTranslation(currentLocale.value, key)
}

async function initGeoLocation() {
  isLoading.value = true
  try {
    const geoData = await getGeoLocation()
    locationInfo.value = {
      city: geoData.city,
      countryCode: geoData.countryCode
    }
    initialPosition.value = {
      lat: geoData.latitude,
      lng: geoData.longitude
    }
    currentLocale.value = getLocaleFromCountry(geoData.countryCode)
  } catch (error) {
    console.warn('Geo location init failed:', error)
  } finally {
    isLoading.value = false
  }
}

function handleBoundsChange(bounds) {
  fetch(`/api/v1/news?bounds=${bounds}`)
    .then(response => response.json())
    .then(data => {
      if (data.data && data.data.length > 0) {
        newsList.value = [...newsList.value, ...data.data]
        breakingNewsList.value = [...breakingNewsList.value, ...data.data.filter(n => n.is_breaking)]
      }
    })
    .catch(error => {
      console.error('Failed to fetch news by bounds:', error)
    })
}

function initSSE() {
  sseService = new SSEService('/api/v1/stream')
  sseService.addListener((news) => {
    breakingNewsList.value = [news, ...breakingNewsList.value].slice(0, 20)
  })
  sseService.connect()
}

async function fetchNews() {
  try {
    const response = await fetch('/api/v1/news')
    console.log('Fetch news response:', response)
    const data = await response.json()
    console.log('News data received:', data)
    if (data.data && data.data.length > 0) {
      newsList.value = data.data
      breakingNewsList.value = data.data.filter(item => item.is_breaking)
      console.log('News list updated:', newsList.value.length, 'items')
    } else {
      console.log('No news data received')
    }
  } catch (error) {
    console.error('Failed to fetch news:', error)
  }
}

onMounted(() => {
  initGeoLocation()
  fetchNews()
  
  setTimeout(() => {
    initSSE()
  }, 1000)
})

onUnmounted(() => {
  if (sseService) {
    sseService.disconnect()
  }
})
</script>

<style scoped>
.app-container {
  position: relative;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background: #f5f7fa;
}

.logo {
  position: fixed;
  left: 24px;
  top: 70px;
  display: flex;
  align-items: center;
  gap: 10px;
  z-index: 100;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 12px;
  padding: 10px 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
}

.logo-icon {
  width: 32px;
  height: 32px;
  color: #3b82f6;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  letter-spacing: 1px;
}

.stats-panel {
  position: fixed;
  right: 24px;
  bottom: 120px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  z-index: 100;
}

.stat-item {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 12px;
  padding: 14px 18px;
  min-width: 100px;
  text-align: center;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
}

.stat-value {
  display: block;
  font-size: 28px;
  font-weight: 700;
  color: #3b82f6;
}

.stat-value.breaking {
  color: #ef4444;
}

.stat-label {
  display: block;
  font-size: 12px;
  color: rgba(0, 0, 0, 0.5);
  margin-top: 4px;
  font-weight: 500;
}

.language-selector {
  position: fixed;
  right: 24px;
  top: 70px;
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 8px;
  z-index: 100;
}

.language-select {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 10px;
  padding: 8px 12px;
  font-size: 13px;
  color: #374151;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  min-width: 100px;
  outline: none;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' class='h-4 w-4' viewBox='0 0 20 20' fill='%236b7280'%3E%3Cpath fill-rule='evenodd' d='M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z' clip-rule='evenodd'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 8px center;
}

.location-indicator {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 10px;
  padding: 6px 12px;
  font-size: 12px;
  color: #374151;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.location-indicator svg {
  color: #ef4444;
}
</style>