# 前端地图组件设计

## 组件结构

```
src/
├── components/
│   ├── MapContainer.vue      # 地图容器组件
│   ├── NewsMarker.vue        # 新闻标记点组件
│   ├── InfoWindow.vue        # 信息弹窗组件
│   ├── BreakingBanner.vue    # 突发新闻横幅
│   └── HistoryPanel.vue      # 历史新闻面板
├── services/
│   └── api.ts                # API 调用服务
└── types/
    └── news.ts               # 类型定义
```

## 类型定义 (types/news.ts)

```typescript
export interface News {
  id: number;
  title: string;
  summary: string;
  source: string;
  publish_time: string;
  latitude: number;
  longitude: number;
  geo_level: 'world' | 'continent' | 'country' | 'city';
  is_breaking: boolean;
  priority: number;
  created_at: string;
}

export interface NewsRequest {
  title: string;
  summary: string;
  source: string;
  publish_time: string;
  geo_level: string;
  latitude: number;
  longitude: number;
  is_breaking: boolean;
  priority: number;
}

export interface HistoryResponse {
  data: News[];
  total: number;
  page: number;
  limit: number;
}
```

## API 服务 (services/api.ts)

```typescript
import axios from 'axios';

const API_BASE = '/api/v1';

export const newsApi = {
  createNews: (data: NewsRequest) => 
    axios.post(`${API_BASE}/news`, data),
  
  getNewsForMap: (level: string, code?: string) => 
    axios.get(`${API_BASE}/news/map`, { 
      params: { level, code } 
    }),
  
  getBreakingNews: () => 
    axios.get(`${API_BASE}/news/breaking`),
  
  getHistoryNews: (start: string, end: string, page = 1, limit = 20) => 
    axios.get(`${API_BASE}/news/history`, {
      params: { start, end, page, limit }
    })
};
```

## 地图组件逻辑

### MapContainer 组件职责

1. **初始化地图**: 使用 Mapbox/Leaflet/Cesium 初始化地图实例
2. **层级钻取**:
   - 默认展示全球视图（热力图/聚合点）
   - 点击大洲 → 展示该大洲 Top N 新闻
   - 点击国家 → 展示该国 Top N 新闻
   - 点击城市 → 展示该市详情新闻
3. **Marker 渲染**: 根据新闻数据在地图上渲染标记点
4. **交互处理**: 处理 Marker 点击事件，弹出 InfoWindow

### BreakingBanner 组件职责

1. **轮询更新**: 每 30 秒调用 `/api/v1/news/breaking`
2. **横幅展示**: 若有突发新闻，显示红色闪烁横幅
3. **高亮定位**: 点击横幅可定位到对应地图位置

### HistoryPanel 组件职责

1. **时间筛选**: 提供 Date Picker 选择时间范围
2. **列表展示**: 展示历史新闻列表，支持分页
3. **地图联动**: 点击列表项可在地图上定位

## 数据流

```
┌─────────────────────────────────────────────────────────────────┐
│                      GeoNews Frontend                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐    │
│  │ Breaking    │    │ Map         │    │ History        │    │
│  │ Banner      │    │ Container   │    │ Panel          │    │
│  └──────┬──────┘    └──────┬──────┘    └────────┬────────┘    │
│         │                  │                     │              │
│         ▼                  ▼                     ▼              │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                      API Service                        │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                  │
│                              ▼                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Backend API                           │   │
│  │  POST /api/v1/news                                      │   │
│  │  GET  /api/v1/news/map                                  │   │
│  │  GET  /api/v1/news/breaking                             │   │
│  │  GET  /api/v1/news/history                              │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## 样式规范

### 颜色方案

| 元素 | 颜色 | 说明 |
|------|------|------|
| 主色调 | #3B82F6 | 蓝色系 |
| 突发新闻 | #EF4444 | 红色警示 |
| 背景 | #FFFFFF | 白色背景 |
| 文字 | #1F2937 | 深色文字 |

### 动画效果

- **突发新闻 Marker**: 脉冲动画（红色警报图标）
- **信息弹窗**: 淡入淡出效果
- **横幅闪烁**: 红色闪烁提醒

## 性能优化

1. **点位聚合**: 使用 Supercluster 进行点位聚合，优化大量 Marker 渲染
2. **按需加载**: 根据缩放级别动态加载不同精度的新闻数据
3. **缓存策略**: 缓存已加载的新闻数据，减少重复请求