import { useCallback, useEffect, useState } from 'react'
import { getToken, deleteToken } from 'firebase/messaging'
import { messagingPromise, firebaseReady } from './firebase'
import { getConfig } from './config'
import { api } from './api'

const FCM_TOKEN_KEY = 'fcm_token'

export type NotificationPermission = 'default' | 'granted' | 'denied' | 'unsupported' | 'unconfigured'

export function useNotifications() {
  const [permission, setPermission] = useState<NotificationPermission>('default')
  const [ready, setReady] = useState(false)
  const [registered, setRegistered] = useState(() => !!localStorage.getItem(FCM_TOKEN_KEY))
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    firebaseReady.then(configured => {
      if (!configured) {
        setPermission('unconfigured')
      } else if (!('Notification' in window)) {
        setPermission('unsupported')
      } else {
        setPermission(Notification.permission as NotificationPermission)
      }
      setReady(true)
    })
  }, [])

  useEffect(() => {
    if (!ready || permission === 'unsupported' || permission === 'unconfigured') return
    const interval = setInterval(() => {
      const current = Notification.permission as NotificationPermission
      setPermission(current)
    }, 2000)
    return () => clearInterval(interval)
  }, [ready, permission])

  const enable = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await Notification.requestPermission()
      setPermission(result as NotificationPermission)
      if (result !== 'granted') return

      const messaging = await messagingPromise
      if (!messaging) {
        setError('Push notifications are not supported in this browser')
        return
      }

      const oldToken = localStorage.getItem(FCM_TOKEN_KEY)
      if (oldToken) {
        await api.unregisterFCMToken(oldToken).catch(() => {})
        localStorage.removeItem(FCM_TOKEN_KEY)
      }
      const existingRegs = await navigator.serviceWorker.getRegistrations()
      for (const reg of existingRegs) {
        const swURL = reg.active?.scriptURL ?? reg.installing?.scriptURL ?? reg.waiting?.scriptURL ?? ''
        if (swURL.includes('firebase-messaging-sw')) await reg.unregister()
      }
      const swRegistration = await navigator.serviceWorker.register('/firebase-messaging-sw.js', { type: 'module', updateViaCache: 'none' })
      const config = await getConfig()
      const token = await getToken(messaging, { vapidKey: config.FIREBASE_VAPID_KEY, serviceWorkerRegistration: swRegistration })
      await api.registerFCMToken(token)
      localStorage.setItem(FCM_TOKEN_KEY, token)
      setRegistered(true)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to enable notifications')
    } finally {
      setLoading(false)
    }
  }, [])

  const disable = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const token = localStorage.getItem(FCM_TOKEN_KEY)
      if (token) {
        const messaging = await messagingPromise
        if (messaging) await deleteToken(messaging)
        await api.unregisterFCMToken(token)
        localStorage.removeItem(FCM_TOKEN_KEY)
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to disable notifications')
    } finally {
      setRegistered(false)
      setLoading(false)
    }
  }, [])

  const enabled = permission === 'granted' && registered

  return { permission, enabled, loading, error, enable, disable }
}
