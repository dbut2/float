interface AppConfig {
  FIREBASE_API_KEY: string
  FIREBASE_AUTH_DOMAIN: string
  FIREBASE_PROJECT_ID: string
  FIREBASE_STORAGE_BUCKET: string
  FIREBASE_MESSAGING_SENDER_ID: string
  FIREBASE_APP_ID: string
  FIREBASE_VAPID_KEY: string
}

let cached: AppConfig | null = null

export async function getConfig(): Promise<AppConfig> {
  if (!cached) {
    cached = await fetch('/config.json').then(r => r.json())
  }
  return cached!
}
