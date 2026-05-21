<template>
  <div class="skeleton-container">
    <div class="skeleton-header">
      <div class="skeleton-banner">
        <div class="skeleton-badge"></div>
        <div class="skeleton-text-line"></div>
      </div>
    </div>
    <div class="skeleton-map">
      <div class="skeleton-map-content">
        <div class="skeleton-grid">
          <div v-for="i in 20" :key="i" class="skeleton-dot"></div>
        </div>
      </div>
    </div>
    <div class="skeleton-footer">
      <div class="skeleton-slider">
        <div class="skeleton-track"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
</script>

<style scoped>
.skeleton-container {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: #f5f7fa;
  z-index: 999;
  display: flex;
  flex-direction: column;
}

.skeleton-header {
  padding: 12px 20px;
  background: linear-gradient(90deg, #ef4444 0%, #dc2626 100%);
}

.skeleton-banner {
  display: flex;
  align-items: center;
  gap: 12px;
}

.skeleton-badge {
  width: 60px;
  height: 24px;
  background: rgba(255, 255, 255, 0.3);
  border-radius: 12px;
  animation: shimmer 1.5s infinite;
}

.skeleton-text-line {
  flex: 1;
  height: 18px;
  background: rgba(255, 255, 255, 0.4);
  border-radius: 4px;
  animation: shimmer 1.5s infinite;
}

.skeleton-map {
  flex: 1;
  background: #e5e7eb;
  position: relative;
  overflow: hidden;
}

.skeleton-map-content {
  width: 100%;
  height: 100%;
  padding: 20px;
}

.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(60px, 1fr));
  gap: 16px;
}

.skeleton-dot {
  width: 12px;
  height: 12px;
  background: #d1d5db;
  border-radius: 50%;
  margin: auto;
  animation: shimmer 1.5s infinite;
}

.skeleton-footer {
  padding: 20px;
  background: white;
  border-top: 1px solid #e5e7eb;
}

.skeleton-slider {
  width: 100%;
  max-width: 500px;
  margin: 0 auto;
}

.skeleton-track {
  height: 8px;
  background: #e5e7eb;
  border-radius: 4px;
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% {
    opacity: 0.4;
  }
  50% {
    opacity: 1;
  }
  100% {
    opacity: 0.4;
  }
}
</style>