export class SSEService {
  constructor(url) {
    this.url = url
    this.eventSource = null
    this.listeners = []
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
  }

  connect() {
    if (this.eventSource) {
      this.eventSource.close()
    }

    this.eventSource = new EventSource(this.url)

    this.eventSource.onmessage = (event) => {
      this.reconnectAttempts = 0
      try {
        const data = JSON.parse(event.data)
        this.listeners.forEach(listener => listener(data))
      } catch (error) {
        console.error('Failed to parse SSE message:', error)
      }
    }

    this.eventSource.onerror = (error) => {
      console.error('SSE error:', error)
      this.eventSource.close()
      
      if (this.reconnectAttempts < this.maxReconnectAttempts) {
        const delay = Math.pow(2, this.reconnectAttempts) * 1000
        setTimeout(() => {
          this.reconnectAttempts++
          this.connect()
        }, delay)
      }
    }

    this.eventSource.onopen = () => {
      console.log('SSE connection opened')
    }
  }

  addListener(listener) {
    this.listeners.push(listener)
  }

  removeListener(listener) {
    this.listeners = this.listeners.filter(l => l !== listener)
  }

  disconnect() {
    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
    }
    this.listeners = []
  }
}