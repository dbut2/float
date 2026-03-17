declare const self: ServiceWorkerGlobalScope
import { initializeApp } from 'firebase/app'
import { getMessaging, onBackgroundMessage } from 'firebase/messaging/sw'

;(async () => {
  const config = await fetch('/config.json', { credentials: 'same-origin' }).then(r => r.json())

  const app = initializeApp({
    apiKey: config.FIREBASE_API_KEY,
    authDomain: config.FIREBASE_AUTH_DOMAIN,
    projectId: config.FIREBASE_PROJECT_ID,
    storageBucket: config.FIREBASE_STORAGE_BUCKET,
    messagingSenderId: config.FIREBASE_MESSAGING_SENDER_ID,
    appId: config.FIREBASE_APP_ID,
  })

  const messaging = getMessaging(app)

  onBackgroundMessage(messaging, (payload) => {
    const { title, body, icon } = payload.notification ?? {}
    self.registration.showNotification(title ?? 'Float', {
      body: body ?? '',
      icon: icon ?? '/icon-192.png',
      badge: '/icon-192.png',
    })
  })
})()
