export interface User {
  user_id: string
  email: string
  created_at: string
}

export interface Transaction {
  transaction_id: string
  bucket_id: string
  description: string
  message: string
  amount_cents: number
  display_amount: string
  currency_code: string
  created_at: string
  is_transaction: boolean
  transaction_type: string | null
  deep_link_url: string
  raw_json: unknown
}

export interface Bucket {
  bucket_id: string
  user_id: string
  name: string
  is_general: boolean
  created_at: string
  balance_cents: number
}

export interface Transfer {
  transfer_id: string
  from_bucket_id: string
  from_bucket_name: string
  to_bucket_id: string
  to_bucket_name: string
  amount_cents: number
  note: string
  created_at: string
}

export interface Trickle {
  trickle_id: string
  from_bucket_id: string
  from_bucket_name: string
  to_bucket_id: string
  to_bucket_name: string
  amount_cents: number
  description: string
  period: 'daily' | 'weekly' | 'fortnightly' | 'monthly'
  start_date: string
  end_date: string | null
  created_at: string
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    headers: { 'Content-Type': 'application/json', ...init?.headers },
    ...init,
  })
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText)
    throw new Error(text || `HTTP ${res.status}`)
  }
  if (res.status === 204) return undefined as T
  const ct = res.headers.get('Content-Type') ?? ''
  if (!ct.includes('json')) return undefined as T
  return res.json()
}

export const api = {
  getUser: () => request<User>('/api/user'),

  getTransactions: () => request<Transaction[]>('/api/transactions'),

  getBuckets: () => request<Bucket[]>('/api/buckets'),

  getBucketTransactions: (id: string) =>
    request<Transaction[]>(`/api/buckets/${id}/transactions`),

  createBucket: (name: string) =>
    request<Bucket>('/api/buckets', {
      method: 'POST',
      body: JSON.stringify({ name }),
    }),

  deleteBucket: (id: string) =>
    request<void>(`/api/buckets/${id}`, { method: 'DELETE' }),

  assignTransaction: (txId: string, bucketId: string) =>
    request<void>(`/api/transactions/${txId}/bucket`, {
      method: 'PUT',
      body: JSON.stringify({ bucket_id: bucketId }),
    }),

  setToken: (token: string) =>
    request<void>('/api/user/token', {
      method: 'PUT',
      body: JSON.stringify({ token }),
    }),

  sync: () => request<{ synced: number }>('/api/user/sync', { method: 'POST' }),

  getTransfers: () => request<Transfer[]>('/api/transfers'),

  createTransfer: (fromBucketId: string, toBucketId: string, amountCents: number, note: string) =>
    request<Transfer>('/api/transfers', {
      method: 'POST',
      body: JSON.stringify({ from_bucket_id: fromBucketId, to_bucket_id: toBucketId, amount_cents: amountCents, note }),
    }),

  deleteTransfer: (id: string) =>
    request<void>(`/api/transfers/${id}`, { method: 'DELETE' }),

  getTrickles: () => request<Trickle[]>('/api/trickles'),

  getTrickle: (bucketId: string) => request<Trickle>(`/api/buckets/${bucketId}/trickle`),

  upsertTrickle: (bucketId: string, data: { amount_cents: number; description: string; period: string; start_date: string; end_date: string | null }) =>
    request<Trickle>(`/api/buckets/${bucketId}/trickle`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  deleteTrickle: (bucketId: string) =>
    request<void>(`/api/buckets/${bucketId}/trickle`, { method: 'DELETE' }),
}

export function formatAUD(cents: number): string {
  const abs = Math.abs(cents) / 100
  return new Intl.NumberFormat('en-AU', {
    style: 'currency',
    currency: 'AUD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(abs)
}

export function formatDate(dateStr: string): string {
  const d = new Date(dateStr)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const days = Math.floor(diff / 86400000)

  if (days === 0) return 'Today'
  if (days === 1) return 'Yesterday'
  if (days < 7) return d.toLocaleDateString('en-AU', { weekday: 'long' })
  return d.toLocaleDateString('en-AU', { month: 'short', day: 'numeric' })
}

export function formatDateShort(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-AU', {
    month: 'short',
    day: 'numeric',
  })
}
