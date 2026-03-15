import { initializeApp } from 'firebase/app'
import { getMessaging, isSupported } from 'firebase/messaging'
import { config } from './config'

const firebaseConfig = {
  apiKey: config.FIREBASE_API_KEY,
  authDomain: config.FIREBASE_AUTH_DOMAIN,
  projectId: config.FIREBASE_PROJECT_ID,
  storageBucket: config.FIREBASE_STORAGE_BUCKET,
  messagingSenderId: config.FIREBASE_MESSAGING_SENDER_ID,
  appId: config.FIREBASE_APP_ID,
}

export const isFirebaseConfigured = !!(
  config.FIREBASE_API_KEY &&
  config.FIREBASE_PROJECT_ID &&
  config.FIREBASE_MESSAGING_SENDER_ID &&
  config.FIREBASE_APP_ID
)

const app = isFirebaseConfigured ? initializeApp(firebaseConfig) : null

export const messagingPromise = isFirebaseConfigured && app
  ? isSupported().then(supported => {
      if (!supported) return null
      try {
        return getMessaging(app)
      } catch {
        return null
      }
    })
  : Promise.resolve(null)
