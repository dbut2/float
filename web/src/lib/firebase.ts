import { initializeApp } from 'firebase/app'
import { getMessaging, isSupported } from 'firebase/messaging'

const firebaseConfig = {
  apiKey: import.meta.env.VITE_FIREBASE_API_KEY,
  authDomain: import.meta.env.VITE_FIREBASE_AUTH_DOMAIN,
  projectId: import.meta.env.VITE_FIREBASE_PROJECT_ID,
  storageBucket: import.meta.env.VITE_FIREBASE_STORAGE_BUCKET,
  messagingSenderId: import.meta.env.VITE_FIREBASE_MESSAGING_SENDER_ID,
  appId: import.meta.env.VITE_FIREBASE_APP_ID,
}

export const isFirebaseConfigured = !!(
  import.meta.env.VITE_FIREBASE_API_KEY &&
  import.meta.env.VITE_FIREBASE_PROJECT_ID &&
  import.meta.env.VITE_FIREBASE_MESSAGING_SENDER_ID &&
  import.meta.env.VITE_FIREBASE_APP_ID
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
