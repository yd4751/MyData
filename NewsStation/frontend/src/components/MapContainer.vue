<template>
  <div class="map-container">
    <div ref="mapContainer" class="map-wrapper"></div>
    <div v-if="isLoading" class="loading-overlay">
      <div class="loading-spinner"></div>
      <span class="loading-text">{{ t('autoLocation') }}</span>
    </div>
    <InfoWindow
      v-if="selectedNews"
      :visible="!!selectedNews"
      :news="selectedNews"
      :position="infoWindowPosition"
      :locale="locale"
      @close="closeInfoWindow"
    />
    <div class="map-controls">
      <div class="control-btn" @click="zoomIn">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clip-rule="evenodd" />
        </svg>
      </div>
      <div class="control-btn" @click="zoomOut">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M3 10a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
        </svg>
      </div>
      <div class="control-btn" @click="resetView">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M10 2a8 8 0 100 16 8 8 0 000-16zm0 14a6 6 0 110-12 6 6 0 010 12zm1-8a1 1 0 10-2 0v3a1 1 0 00.293.707l2 2a1 1 0 101.414-1.414L11 10.586V8z" clip-rule="evenodd" />
        </svg>
      </div>
    </div>
    <div class="legend">
      <div class="legend-title">{{ t('newsCount') }}</div>
      <div class="legend-items">
        <div class="legend-item">
          <div class="legend-dot small"></div>
          <span>1-5</span>
        </div>
        <div class="legend-item">
          <div class="legend-dot medium"></div>
          <span>5-20</span>
        </div>
        <div class="legend-item">
          <div class="legend-dot large"></div>
          <span>20+</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import 'leaflet.markercluster/dist/MarkerCluster.css'
import 'leaflet.markercluster/dist/MarkerCluster.Default.css'
import 'leaflet.markercluster'
import InfoWindow from './InfoWindow.vue'
import { getTranslation } from '../i18n/languages'

const props = defineProps({
  initialPosition: {
    type: Object,
    default: () => ({ lat: 39.9042, lng: 116.4074 })
  },
  locale: {
    type: String,
    default: 'zh-CN'
  },
  isLoading: {
    type: Boolean,
    default: false
  },
  newsData: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['bounds-change'])

const mapContainer = ref(null)
const selectedNews = ref(null)
const infoWindowPosition = ref({ x: 0, y: 0 })
let map = null
let markers = []
let markerCluster = null
let boundsTimer = null

function t(key) {
  return getTranslation(props.locale, key)
}

function createMarkers() {
  console.log('Creating markers with:', props.newsData.length, 'items')
  
  if (markerCluster) {
    map.removeLayer(markerCluster)
    markerCluster = null
  }
  
  markers.forEach(marker => {
    map.removeLayer(marker)
  })
  markers = []

  markerCluster = L.markerClusterGroup({
    maxClusterRadius: 50,
    spiderfyOnMaxZoom: false,
    showCoverageOnHover: false,
    zoomToBoundsOnClick: false,
    disableClusteringAtZoom: 18,
    iconCreateFunction: function(cluster) {
      const count = cluster.getChildCount()
      const zoom = map.getZoom()
      let size = 'small'
      let color = '#3b82f6'
      
      if (count >= 30) {
        size = 'large'
        color = '#dc2626'
      } else if (count >= 10) {
        size = 'medium'
        color = '#8b5cf6'
      } else if (count >= 5) {
        size = 'small'
        color = '#3b82f6'
      } else if (count >= 2) {
        size = 'tiny'
        color = '#10b981'
      }
      
      return L.divIcon({
        html: `<div class="cluster-marker ${size}" style="background: ${color}">${count}</div>`,
        className: '',
        iconSize: L.point(40, 40)
      })
    }
  })

  props.newsData.forEach((item, index) => {
    const icon = L.divIcon({
      html: `<div class="news-marker ${item.is_breaking ? 'breaking' : ''}" data-index="${index}"></div>`,
      className: '',
      iconSize: L.point(item.is_breaking ? 16 : 12, item.is_breaking ? 16 : 12)
    })

    const marker = L.marker([item.latitude, item.longitude], { icon })
    
    marker.bindPopup(`<div style="padding: 8px; min-width: 250px;">
      <h4 style="margin: 0 0 8px 0; font-size: 14px; font-weight: 600;">${item.title}</h4>
      <p style="margin: 0 0 8px 0; font-size: 12px; color: #666; line-height: 1.4;">${item.summary}</p>
      <div style="font-size: 11px; color: #999;">来源: ${item.source}</div>
    </div>`)
    
    marker.on('click', function(e) {
      e.stopPropagation()
      e.originalEvent.stopPropagation()
      console.log('Single marker clicked:', item.title)
      setTimeout(() => {
        this.openPopup()
      }, 0)
    })
    
    marker.on('popupopen', function(e) {
      console.log('Popup opened for:', item.title)
      const latlng = e.target.getLatLng()
      const point = map.latLngToContainerPoint(latlng)
      infoWindowPosition.value = { x: point.x, y: point.y }
      selectedNews.value = { ...item }
    })
    
    marker.on('popupclose', function(e) {
      console.log('Popup closed for:', item.title)
    })

    markers.push(marker)
    markerCluster.addLayer(marker)
  })

  markerCluster.on('clusterclick', function(e) {
      console.log('Cluster clicked, child count:', e.layer.getChildCount())
      if (e.originalEvent) {
        e.originalEvent.stopPropagation()
      }
      
      const childCount = e.layer.getChildCount()
      
      if (childCount === 1) {
        e.layer.zoomToBounds({ paddingTopLeft: [10, 10], paddingBottomRight: [10, 10] })
        setTimeout(() => {
          map.setZoom(map.getZoom() + 1)
        }, 300)
      } else if (childCount <= 5) {
        map.setZoom(map.getZoom() + 1)
        e.layer.zoomToBounds({ paddingTopLeft: [20, 20], paddingBottomRight: [20, 20] })
      } else if (childCount <= 15) {
        map.setZoom(map.getZoom() + 1)
        e.layer.zoomToBounds({ paddingTopLeft: [15, 15], paddingBottomRight: [15, 15] })
      } else {
        e.layer.zoomToBounds({ paddingTopLeft: [10, 10], paddingBottomRight: [10, 10] })
        map.setZoom(map.getZoom() + 2)
      }
      console.log('Zooming to cluster bounds, childCount:', childCount)
    })
  
  map.addLayer(markerCluster)
  console.log('MarkerCluster added to map')
  
  setTimeout(() => {
    markerCluster.eachLayer(function(marker) {
      if (marker instanceof L.Marker && !marker._isCluster) {
        marker.off('click')
        marker.on('click', function(e) {
          e.stopPropagation()
          console.log('Marker click after timeout:', marker)
          marker.openPopup()
        })
      }
    })
    console.log('Re-bound click events to markers')
  }, 100)
}

function handleBoundsChange() {
  if (boundsTimer) {
    clearTimeout(boundsTimer)
  }
  
  boundsTimer = setTimeout(() => {
    const bounds = map.getBounds()
    const sw = bounds.getSouthWest()
    const ne = bounds.getNorthEast()
    const boundsStr = `${sw.lng},${sw.lat},${ne.lng},${ne.lat}`
    emit('bounds-change', boundsStr)
  }, 500)
}

function zoomIn() {
  map.zoomIn()
}

function zoomOut() {
  map.zoomOut()
}

function resetView() {
  map.setView([props.initialPosition.lat, props.initialPosition.lng], 10)
}

function closeInfoWindow() {
  selectedNews.value = null
}

function updateMapPosition(position) {
  if (map) {
    map.setView([position.lat, position.lng], 10)
  }
}

watch(() => props.initialPosition, (newPosition) => {
  updateMapPosition(newPosition)
}, { deep: true })

watch(() => props.newsData, (newData) => {
  console.log('News data updated in MapContainer:', newData?.length || 0, 'items')
  if (map) {
    createMarkers()
  }
}, { deep: true })

onMounted(() => {
  map = L.map(mapContainer.value, {
    center: [props.initialPosition.lat, props.initialPosition.lng],
    zoom: 10,
    zoomControl: false
  })

  L.tileLayer('https://webrd0{s}.is.autonavi.com/appmaptile?lang=zh_cn&size=1&scale=1&style=8&x={x}&y={y}&z={z}', {
    subdomains: ['1', '2', '3', '4'],
    attribution: '© 高德地图',
    maxZoom: 19
  }).addTo(map)

  createMarkers()

  map.on('click', (e) => {
    console.log('Map clicked at:', e.latlng)
    
    const target = e.originalEvent.target
    const isMarker = target.classList.contains('news-marker') || 
                     target.closest && target.closest('.news-marker')
    
    if (isMarker) {
      console.log('Clicked on news marker element')
      
      map.eachLayer(function(layer) {
        if (layer instanceof L.Marker && !layer._isCluster) {
          const latlng = layer.getLatLng()
          const point = map.latLngToContainerPoint(latlng)
          const clickPoint = map.latLngToContainerPoint(e.latlng)
          
          const distance = Math.sqrt(
            Math.pow(point.x - clickPoint.x, 2) + 
            Math.pow(point.y - clickPoint.y, 2)
          )
          
          if (distance < 20) {
            console.log('Found marker at click position, opening popup')
            layer.openPopup()
          }
        }
      })
    }
    
    if (selectedNews.value && !isMarker) {
      closeInfoWindow()
    }
  })
  
  map.on('layeradd', function(e) {
    console.log('Layer added:', e.layer)
  })
  
  map.on('popupopen', function(e) {
    console.log('Global popup open event:', e.popup)
    if (e.popup._source) {
      console.log('Popup source:', e.popup._source)
    }
  })

  map.on('moveend', handleBoundsChange)
  map.on('zoomend', handleBoundsChange)
})

onUnmounted(() => {
  if (map) {
    map.remove()
  }
  if (boundsTimer) {
    clearTimeout(boundsTimer)
  }
})
</script>

<style scoped>
.map-container {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 1;
}

.map-wrapper {
  width: 100%;
  height: 100%;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.8);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 200;
  gap: 16px;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #e5e7eb;
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-text {
  font-size: 14px;
  color: #6b7280;
}

.cluster-marker {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 14px;
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  border: 2px solid rgba(255, 255, 255, 0.5);
}

.cluster-marker.tiny {
  width: 24px;
  height: 24px;
  font-size: 11px;
}

.cluster-marker.small {
  width: 28px;
  height: 28px;
  font-size: 12px;
}

.cluster-marker.medium {
  width: 34px;
  height: 34px;
  font-size: 13px;
}

.cluster-marker.large {
  width: 42px;
  height: 42px;
  font-size: 14px;
}

.news-marker {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #3b82f6;
  border: 2px solid rgba(255, 255, 255, 0.8);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

.news-marker.breaking {
  width: 16px;
  height: 16px;
  background: #ef4444;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.7);
  }
  50% {
    box-shadow: 0 0 0 8px rgba(239, 68, 68, 0);
  }
}

.map-controls {
  position: absolute;
  right: 24px;
  top: 100px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  z-index: 50;
}

.control-btn {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 10px;
  padding: 10px;
  color: #374151;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.control-btn:hover {
  background: rgba(59, 130, 246, 0.1);
  transform: translateY(-2px);
}

.legend {
  position: absolute;
  left: 24px;
  bottom: 120px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border: 1px solid rgba(0, 0, 0, 0.08);
  border-radius: 12px;
  padding: 16px;
  z-index: 50;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
}

.legend-title {
  font-size: 12px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 10px;
}

.legend-items {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  color: #6b7280;
}

.legend-dot {
  border-radius: 50%;
  border: 2px solid rgba(0, 0, 0, 0.1);
}

.legend-dot.small {
  width: 12px;
  height: 12px;
  background: #3b82f6;
}

.legend-dot.medium {
  width: 16px;
  height: 16px;
  background: #8b5cf6;
}

.legend-dot.large {
  width: 20px;
  height: 20px;
  background: #ec4899;
}

.popup-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 8px;
}

.popup-summary {
  font-size: 12px;
  color: #6b7280;
}
</style>