export interface User {
  user_id: string
  email: string
  created_at: string
  has_token: boolean
}

export interface Cover {
  cover_id: string
  transaction_id: string
  amount_cents: number
  display_amount: string
  note: string
  created_at: string
  display_date: string
}

export interface Transaction {
  transaction_id: string
  bucket_id: string
  description: string
  message: string
  amount_cents: number
  display_amount: string
  created_at: string
  display_date: string
  is_transaction: boolean
  transaction_type?: string | null
  raw_json?: unknown
  foreign_currency_code?: string | null
  foreign_amount_cents?: number | null
  foreign_display_amount?: string | null
  covers?: Cover[]
  covers_amount_cents?: number
  net_amount_cents?: number
  net_display_amount?: string
}

export interface Bucket {
  bucket_id: string
  user_id: string
  name: string
  description: string
  is_general: boolean
  created_at: string
  balance_cents: number
  balance_display: string
  currency_code?: string | null
  fx_rate?: number | null
  foreign_balance_display?: string | null
}

export interface Transfer {
  transfer_id: string
  from_bucket_id: string
  from_bucket_name: string
  to_bucket_id: string
  to_bucket_name: string
  amount_cents: number
  display_amount: string
  description: string
  note: string
  created_at: string
  display_date: string
}

export interface Trickle {
  trickle_id: string
  from_bucket_id: string
  from_bucket_name: string
  to_bucket_id: string
  to_bucket_name: string
  amount_cents: number
  display_amount: string
  description: string
  period: 'daily' | 'weekly' | 'fortnightly' | 'monthly'
  start_date: string
  end_date: string | null
  created_at: string
}

export interface BucketHealth {
  bucket_id: string
  bucket_name: string
  trickle_amount: number
  trickle_amount_cents: number
  spent: number
  daily_allowance: number
  next_trickle_at: string | null
  status: 'great' | 'ok' | 'warn' | 'critical' | 'stale'
  period: string | null
}

export interface HealthSummary {
  buckets: BucketHealth[]
  overall_score: number
  at_risk_count: number
  stale_count: number
  healthy_count: number
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

  createBucket: (name: string, currencyCode?: string, description?: string) =>
    request<Bucket>('/api/buckets', {
      method: 'POST',
      body: JSON.stringify({ name, currency_code: currencyCode ?? null, description: description ?? '' }),
    }),

  deleteBucket: (id: string) =>
    request<void>(`/api/buckets/${id}`, { method: 'DELETE' }),

  closeBucket: (id: string) =>
    request<void>(`/api/buckets/${id}/close`, { method: 'POST' }),

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

  getTransactBalance: () => request<{ balance_cents: number; balance_display: string }>('/api/user/balance'),

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

  reorderBuckets: (bucketIds: string[]) =>
    request<void>('/api/buckets/order', {
      method: 'PUT',
      body: JSON.stringify({ bucket_ids: bucketIds }),
    }),

  updateBucketDescription: (bucketId: string, description: string) =>
    request<void>(`/api/buckets/${bucketId}/description`, {
      method: 'PUT',
      body: JSON.stringify({ description }),
    }),

  createCover: (txId: string, amountCents: number, note: string) =>
    request<Cover>(`/api/transactions/${txId}/covers`, {
      method: 'POST',
      body: JSON.stringify({ amount_cents: amountCents, note }),
    }),

  deleteCover: (coverId: string) =>
    request<void>(`/api/covers/${coverId}`, { method: 'DELETE' }),

  getHealth: () => request<HealthSummary>('/api/health'),

  applyTrickleSuggestion: (bucketId: string, amountCents: number, period: string) =>
    request<Trickle>(`/api/buckets/${bucketId}/trickle/apply-suggestion`, {
      method: 'PUT',
      body: JSON.stringify({ amount_cents: amountCents, period }),
    }),

  reclassify: () =>
    request<{ status: string }>('/api/classify/reclassify', { method: 'POST' }),

  reclassifyStatus: () =>
    request<{ running: boolean; total: number; processed: number; reclassified: number }>('/api/classify/status'),

  registerFCMToken: (token: string) =>
    request<void>('/api/fcm-tokens', {
      method: 'POST',
      body: JSON.stringify({ token }),
    }),

  unregisterFCMToken: (token: string) =>
    request<void>('/api/fcm-tokens', {
      method: 'DELETE',
      body: JSON.stringify({ token }),
    }),
}

