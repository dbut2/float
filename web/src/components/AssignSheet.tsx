import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus } from 'lucide-react'
import { api, type Transaction } from '../lib/api'
import { useDraggableSheet } from '../hooks/useDraggableSheet'
import { useMediaQuery } from '../hooks/useMediaQuery'

interface Props {
  transaction: Transaction
  onClose: () => void
}

export default function AssignSheet({ transaction, onClose }: Props) {
  const qc = useQueryClient()
  const [creating, setCreating] = useState(false)
  const [newName, setNewName] = useState('')

  const { data: buckets = [] } = useQuery({
    queryKey: ['buckets'],
    queryFn: api.getBuckets,
  })

  const generalBucket = buckets.find((b) => b.is_general)

  const assign = useMutation({
    mutationFn: (bucketId: string) =>
      api.assignTransaction(transaction.transaction_id, bucketId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['transactions'] })
      qc.invalidateQueries({ queryKey: ['bucket-transactions'] })
      qc.invalidateQueries({ queryKey: ['buckets'] })
      onClose()
    },
  })

  const createAndAssign = useMutation({
    mutationFn: async () => {
      const bucket = await api.createBucket(newName.trim())
      await api.assignTransaction(transaction.transaction_id, bucket.bucket_id)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['transactions'] })
      qc.invalidateQueries({ queryKey: ['bucket-transactions'] })
      qc.invalidateQueries({ queryKey: ['buckets'] })
      onClose()
    },
  })

  const isLoading = assign.isPending || createAndAssign.isPending

  const isDesktop = useMediaQuery('(min-width: 768px)')
  const { handleRef, sheetStyle, backdropStyle, onAnimationEnd } = useDraggableSheet({ onClose })

  return (
    <>
      {/* Backdrop */}
      <div
        onClick={onClose}
        style={{
          position: 'fixed',
          inset: 0,
          background: 'rgba(0,0,0,0.6)',
          zIndex: 100,
          animation: 'fadeIn 0.2s ease forwards',
          ...backdropStyle,
        }}
      />

      {/* Sheet */}
      <div
        onAnimationEnd={onAnimationEnd}
        style={isDesktop ? {
          position: 'fixed', top: '50%', left: '50%',
          transform: 'translate(-50%, -50%)',
          width: 480, zIndex: 101,
          background: 'var(--surface)', borderRadius: 20,
          padding: '24px', animation: 'fadeIn 0.2s ease forwards',
          maxHeight: '85vh', overflowY: 'auto',
        } : {
          position: 'fixed',
          bottom: 0,
          left: 0,
          right: 0,
          zIndex: 101,
          background: 'var(--surface)',
          borderRadius: '20px 20px 0 0',
          paddingBottom: 'calc(64px + 24px)',
          ...sheetStyle,
        }}
      >
        {/* Handle */}
        {!isDesktop && (
          <div ref={handleRef} style={{ touchAction: 'none', display: 'flex', justifyContent: 'center', paddingTop: 12, paddingBottom: 8 }}>
            <div style={{ width: 36, height: 4, background: 'var(--border)', borderRadius: 2 }} />
          </div>
        )}

        {/* Transaction summary */}
        <div style={{ padding: '12px 20px 16px', borderBottom: '1px solid var(--border)' }}>
          <p style={{ fontSize: 13, color: 'var(--text-2)', fontFamily: 'DM Sans', marginBottom: 4 }}>
            Assign to bucket
          </p>
          <p style={{ fontSize: 16, fontWeight: 500, color: 'var(--text)', lineHeight: 1.3 }}>
            {transaction.description}
          </p>
          <p
            className={transaction.display_amount.startsWith('-') ? 'amount-negative' : 'amount-positive'}
            style={{ fontSize: 20, fontWeight: 600, marginTop: 4 }}
          >
            {transaction.display_amount}
          </p>
        </div>

        {/* Bucket list */}
        <div style={{ maxHeight: 280, overflowY: 'auto', padding: '8px 0' }}>
          {buckets.map((bucket) => (
            <button
              key={bucket.bucket_id}
              onClick={() => assign.mutate(bucket.bucket_id)}
              disabled={isLoading}
              style={{
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                padding: '14px 20px',
                background: transaction.bucket_id === bucket.bucket_id
                  ? 'var(--accent-dim)'
                  : 'transparent',
                border: 'none',
                cursor: 'pointer',
                transition: 'background 0.1s',
              }}
            >
              <span style={{
                fontFamily: 'DM Sans',
                fontSize: 16,
                color: transaction.bucket_id === bucket.bucket_id ? 'var(--accent)' : 'var(--text)',
                fontWeight: transaction.bucket_id === bucket.bucket_id ? 600 : 400,
              }}>
                {bucket.name}
              </span>
              <span style={{
                fontFamily: 'JetBrains Mono',
                fontSize: 13,
                color: 'var(--text-2)',
              }}>
                {bucket.balance_display}
              </span>
            </button>
          ))}

          {/* Create new bucket row */}
          {creating ? (
            <div style={{ padding: '12px 20px', display: 'flex', gap: 10 }}>
              <input
                autoFocus
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && newName.trim()) createAndAssign.mutate()
                  if (e.key === 'Escape') { setCreating(false); setNewName('') }
                }}
                placeholder="Bucket name"
                style={{
                  flex: 1,
                  background: 'var(--surface-2)',
                  border: '1px solid var(--border)',
                  borderRadius: 8,
                  padding: '10px 14px',
                  color: 'var(--text)',
                  fontFamily: 'DM Sans',
                  fontSize: 15,
                  outline: 'none',
                }}
              />
              <button
                onClick={() => newName.trim() && createAndAssign.mutate()}
                disabled={!newName.trim() || isLoading}
                style={{
                  background: 'var(--accent)',
                  border: 'none',
                  borderRadius: 8,
                  padding: '10px 18px',
                  color: '#000',
                  fontFamily: 'Syne',
                  fontWeight: 700,
                  fontSize: 14,
                  cursor: 'pointer',
                  opacity: !newName.trim() ? 0.4 : 1,
                }}
              >
                Create
              </button>
            </div>
          ) : (
            <button
              onClick={() => setCreating(true)}
              style={{
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                gap: 12,
                padding: '14px 20px',
                background: 'transparent',
                border: 'none',
                cursor: 'pointer',
                color: 'var(--text-2)',
              }}
            >
              <span style={{
                width: 28, height: 28,
                borderRadius: '50%',
                border: '1.5px dashed var(--border)',
                display: 'flex', alignItems: 'center', justifyContent: 'center',
              }}>
                <Plus size={16} color="var(--text-3)" strokeWidth={1.75} />
              </span>
              <span style={{ fontFamily: 'DM Sans', fontSize: 15 }}>New bucket</span>
            </button>
          )}
        </div>

        {/* Unassign option — assign back to General bucket */}
        {generalBucket && transaction.bucket_id !== generalBucket.bucket_id && (
          <div style={{ padding: '8px 20px 0', borderTop: '1px solid var(--border)' }}>
            <button
              onClick={() => assign.mutate(generalBucket.bucket_id)}
              disabled={isLoading}
              style={{
                width: '100%',
                padding: '14px',
                background: 'transparent',
                border: 'none',
                color: 'var(--red)',
                fontFamily: 'DM Sans',
                fontSize: 15,
                cursor: 'pointer',
              }}
            >
              Remove from bucket
            </button>
          </div>
        )}
      </div>
    </>
  )
}
