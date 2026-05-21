export const languages = {
  'zh-CN': {
    name: '简体中文',
    translations: {
      breakingNews: '突发新闻',
      newsCount: '新闻数量',
      breakingCount: '突发新闻',
      timeRange: '时间范围',
      daysAgo: '天前',
      today: '今天',
      source: '来源',
      breaking: '突发',
      close: '关闭',
      loading: '加载中...',
      autoLocation: '自动定位中...',
      alert: '紧急预警',
      techSummit: '科技峰会',
      aiBreakthrough: '人工智能新突破',
      climateAgreement: '全球气候协议达成'
    }
  },
  'zh-TW': {
    name: '繁體中文',
    translations: {
      breakingNews: '突發新聞',
      newsCount: '新聞數量',
      breakingCount: '突發新聞',
      timeRange: '時間範圍',
      daysAgo: '天前',
      today: '今天',
      source: '來源',
      breaking: '突發',
      close: '關閉',
      loading: '載入中...',
      autoLocation: '自動定位中...',
      alert: '緊急預警',
      techSummit: '科技峰會',
      aiBreakthrough: '人工智慧新突破',
      climateAgreement: '全球氣候協議達成'
    }
  },
  'ja-JP': {
    name: '日本語',
    translations: {
      breakingNews: '速報ニュース',
      newsCount: 'ニュース数',
      breakingCount: '速報',
      timeRange: '時間範囲',
      daysAgo: '日前',
      today: '今日',
      source: '出典',
      breaking: '速報',
      close: '閉じる',
      loading: '読み込み中...',
      autoLocation: '自動位置特定中...',
      alert: '緊急警報',
      techSummit: 'テクノロジーサミット',
      aiBreakthrough: 'AIの新たな飛躍',
      climateAgreement: '地球温暖化協定成立'
    }
  },
  'ko-KR': {
    name: '한국어',
    translations: {
      breakingNews: '속보 뉴스',
      newsCount: '뉴스 수',
      breakingCount: '속보',
      timeRange: '시간 범위',
      daysAgo: '일 전',
      today: '오늘',
      source: '출처',
      breaking: '속보',
      close: '닫기',
      loading: '로딩 중...',
      autoLocation: '자동 위치 설정 중...',
      alert: '긴급 경보',
      techSummit: '테크 서밋',
      aiBreakthrough: '인공지능 새로운 돌파',
      climateAgreement: '글로벌 기후 협정 체결'
    }
  },
  'en-US': {
    name: 'English',
    translations: {
      breakingNews: 'Breaking News',
      newsCount: 'News Count',
      breakingCount: 'Breaking',
      timeRange: 'Time Range',
      daysAgo: 'days ago',
      today: 'Today',
      source: 'Source',
      breaking: 'Breaking',
      close: 'Close',
      loading: 'Loading...',
      autoLocation: 'Auto locating...',
      alert: 'Emergency Alert',
      techSummit: 'Tech Summit',
      aiBreakthrough: 'AI Breakthrough',
      climateAgreement: 'Global Climate Agreement'
    }
  },
  'de-DE': {
    name: 'Deutsch',
    translations: {
      breakingNews: 'Breaking News',
      newsCount: 'Nachrichtenanzahl',
      breakingCount: 'Breaking',
      timeRange: 'Zeitbereich',
      daysAgo: 'Tage zuvor',
      today: 'Heute',
      source: 'Quelle',
      breaking: 'Breaking',
      close: 'Schließen',
      loading: 'Laden...',
      autoLocation: 'Automatische Ortung...',
      alert: 'Notfallwarnung',
      techSummit: 'Tech Summit',
      aiBreakthrough: 'KI-Durchbruch',
      climateAgreement: 'Globaler Klimavertrag'
    }
  },
  'fr-FR': {
    name: 'Français',
    translations: {
      breakingNews: 'Infos en direct',
      newsCount: 'Nombre de nouvelles',
      breakingCount: 'Flash',
      timeRange: 'Plage horaire',
      daysAgo: 'jours ago',
      today: 'Aujourd\'hui',
      source: 'Source',
      breaking: 'Flash',
      close: 'Fermer',
      loading: 'Chargement...',
      autoLocation: 'Localisation automatique...',
      alert: 'Alerte d\'urgence',
      techSummit: 'Sommet tech',
      aiBreakthrough: 'Percée IA',
      climateAgreement: 'Accord climatique mondial'
    }
  },
  'ru-RU': {
    name: 'Русский',
    translations: {
      breakingNews: 'Важные новости',
      newsCount: 'Количество новостей',
      breakingCount: 'Важное',
      timeRange: 'Диапазон времени',
      daysAgo: 'дней назад',
      today: 'Сегодня',
      source: 'Источник',
      breaking: 'Важное',
      close: 'Закрыть',
      loading: 'Загрузка...',
      autoLocation: 'Автоматическое определение...',
      alert: 'Экстренное предупреждение',
      techSummit: 'Технологический саммит',
      aiBreakthrough: 'Прорыв в ИИ',
      climateAgreement: 'Глобальное климатическое соглашение'
    }
  }
}

export function getTranslation(locale, key) {
  const lang = languages[locale] || languages['en-US']
  return lang.translations[key] || key
}