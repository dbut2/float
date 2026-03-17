import { initializeApp, FirebaseApp } from 'firebase/app'
import { getMessaging, isSupported, Messaging } from 'firebase/messaging'
import { getConfig } from './config'

let _app: FirebaseApp | null = null
let _configured = false

export const firebaseReady = getConfig().then(config => {
  _configured = !!(
    config.FIREBASE_API_KEY &&
    config.FIREBASE_PROJECT_ID &&
    config.FIREBASE_MESSAGING_SENDER_ID &&
    config.FIREBASE_APP_ID
  )
  if (_configured) {
    _app = initializeApp({
      apiKey: config.FIREBASE_API_KEY,
      authDomain: config.FIREBASE_AUTH_DOMAIN,
      projectId: config.FIREBASE_PROJECT_ID,
      storageBucket: config.FIREBASE_STORAGE_BUCKET,
      messagingSenderId: config.FIREBASE_MESSAGING_SENDER_ID,
      appId: config.FIREBASE_APP_ID,
    })
  }
  return _configured
})

export function isFirebaseConfigured(): boolean {
  return _configured
}

export const messagingPromise: Promise<Messaging | null> = firebaseReady.then(async () => {
  if (!_configured || !_app) return null
  const supported = await isSupported()
  if (!supported) return null
  try {
    await navigator.serviceWorker.register('/firebase-messaging-sw.js', { type: 'module', updateViaCache: 'none' })
    return getMessaging(_app)
  } catch {
    return null
  }
})
