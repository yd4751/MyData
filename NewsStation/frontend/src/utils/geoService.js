export async function getGeoLocation() {
  try {
    const response = await fetch('https://ipapi.co/json/')
    const data = await response.json()
    
    return {
      success: true,
      country: data.country_name || '',
      countryCode: data.country_code || '',
      region: data.region || '',
      city: data.city || '',
      latitude: parseFloat(data.latitude) || 39.9042,
      longitude: parseFloat(data.longitude) || 116.4074,
      timezone: data.timezone || '',
      languages: data.languages || ''
    }
  } catch (error) {
    console.warn('Failed to get geolocation from IP:', error)
    return {
      success: false,
      country: 'China',
      countryCode: 'CN',
      region: 'Beijing',
      city: 'Beijing',
      latitude: 39.9042,
      longitude: 116.4074,
      timezone: 'Asia/Shanghai',
      languages: 'zh-CN'
    }
  }
}

export function getLocaleFromCountry(countryCode) {
  const localeMap = {
    'CN': 'zh-CN',
    'TW': 'zh-TW',
    'HK': 'zh-HK',
    'JP': 'ja-JP',
    'KR': 'ko-KR',
    'US': 'en-US',
    'GB': 'en-GB',
    'DE': 'de-DE',
    'FR': 'fr-FR',
    'ES': 'es-ES',
    'RU': 'ru-RU',
    'IN': 'hi-IN'
  }
  return localeMap[countryCode] || 'en-US'
}

export function getLanguageName(locale) {
  const langMap = {
    'zh-CN': '简体中文',
    'zh-TW': '繁体中文',
    'zh-HK': '繁体中文',
    'ja-JP': '日本語',
    'ko-KR': '한국어',
    'en-US': 'English',
    'en-GB': 'English',
    'de-DE': 'Deutsch',
    'fr-FR': 'Français',
    'es-ES': 'Español',
    'ru-RU': 'Русский',
    'hi-IN': 'हिन्दी'
  }
  return langMap[locale] || locale
}