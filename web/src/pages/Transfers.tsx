import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api, formatAUD, formatDate } from '../lib/api'
import TransferSheet from '../components/TransferSheet'

export default function Transfers() {
  const qc = useQueryClient()
  const [showCreate, setShowCreate] = useState(false)
  const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null)

  const { data: transfers = [], isLoading } = useQuery({
    queryKey: ['transfers'],
    queryFn: api.getTransfers,
  })

  const deleteTransfer = useMutation({
    mutationFn: (id: string) => api.deleteTransfer(id),
    onMutate: (id) => setPendingDeleteId(id),
    onSettled: () => setPendingDeleteId(null),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['transfers'] })
      qc.invalidateQueries({ queryKey: ['buckets'] })
    },
  })

  return (
    <div style={{ padding: '24px 20px', minHeight: '100%' }}>
      <div className="animate-fade-up" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16, opacity: 0 }}>
        <h1 style={{ fontFamily: 'Syne', fontWeight: 800, fontSize: 22, color: 'var(--text)' }}>
          Transfers
        </h1>
        <button
          onClick={() => setShowCreate(true)}
          className="pressable"
          style={{
            background: 'var(--surface)',
            border: '1px solid var(--border)',
            borderRadius: 10,
            width: 34,
            height: 34,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            cursor: 'pointer',
            color: 'var(--accent)',
            fontSize: 20,
            fontWeight: 300,
            lineHeight: 1,
          }}
        >
          +
        </button>
      </div>

      <div className="animate-fade-up stagger-1" style={{ opacity: 0 }}>
        {isLoading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="shimmer" style={{ height: 68, borderRadius: 14, marginBottom: 10 }} />
          ))
        ) : transfers.length === 0 ? (
          <div style={{ padding: '60px 0', textAlign: 'center', color: 'var(--text-2)' }}>
            <p style={{ fontSize: 32, marginBottom: 10 }}>⇄</p>
            <p style={{ fontFamily: 'DM Sans', fontSize: 15 }}>No transfers yet</p>
          </div>
        ) : (
          transfers.map((t) => (
            <div
              key={t.transfer_id}
              style={{
                background: 'var(--surface)',
                border: '1px solid var(--border)',
                borderRadius: 14,
                padding: '14px 16px',
                marginBottom: 10,
                display: 'flex',
                alignItems: 'center',
                gap: 12,
              }}
            >
              <div style={{ flex: 1, minWidth: 0 }}>
                <p style={{ fontSize: 14, color: 'var(--text)', fontWeight: 500, marginBottom: 3 }}>
                  <span style={{ color: 'var(--text-2)' }}>{t.from_bucket_name}</span>
                  {' → '}
                  <span>{t.to_bucket_name}</span>
                </p>
                {t.note && (
                  <p className="line-clamp-1" style={{ fontSize: 12, color: 'var(--text-3)', marginBottom: 3 }}>{t.note}</p>
                )}
                <p style={{ fontSize: 12, color: 'var(--text-2)' }}>{formatDate(t.created_at)}</p>
              </div>
              <p className="amount-neutral" style={{ fontSize: 16, fontWeight: 600, flexShrink: 0 }}>
                {formatAUD(t.amount_cents)}
              </p>
              <button
                onClick={() => deleteTransfer.mutate(t.transfer_id)}
                disabled={pendingDeleteId === t.transfer_id}
                style={{
                  background: 'rgba(248,113,113,0.1)',
                  border: 'none',
                  borderRadius: 8,
                  width: 32,
                  height: 32,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  cursor: 'pointer',
                  fontSize: 14,
                  flexShrink: 0,
                }}
              >
                🗑
              </button>
            </div>
          ))
        )}
      </div>

      {showCreate && (
        <TransferSheet onClose={() => setShowCreate(false)} />
      )}
    </div>
  )
}
