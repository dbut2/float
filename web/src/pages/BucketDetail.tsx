import { useState } from 'react'
import { createPortal } from 'react-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams, useNavigate } from 'react-router-dom'
import { api, formatAUD, formatDate, type Transaction, type Transfer } from '../lib/api'
import AssignSheet from '../components/AssignSheet'
import TransferSheet from '../components/TransferSheet'
import { useDraggableSheet } from '../hooks/useDraggableSheet'
import { useMediaQuery } from '../hooks/useMediaQuery'

type ListItem =
  | { kind: 'transaction'; tx: Transaction }
  | { kind: 'transfer'; t: Transfer; amountCents: number }

export default function BucketDetail() {
  const { id } = useParams<{ id: string }>()
  const bucketId = id!
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [assignTx, setAssignTx] = useState<Transaction | null>(null)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [showTransfer, setShowTransfer] = useState(false)

  const { data: buckets = [] } = useQuery({
    queryKey: ['buckets'],
    queryFn: api.getBuckets,
  })
  const bucket = buckets.find((b) => b.bucket_id === bucketId)

  const { data: bucketTx = [], isLoading } = useQuery({
    queryKey: ['bucket-transactions', bucketId],
    queryFn: () => api.getBucketTransactions(bucketId),
    enabled: !!bucketId,
  })

  const { data: allTransfers = [] } = useQuery({
    queryKey: ['transfers'],
    queryFn: api.getTransfers,
  })

  const transactions = bucketTx

  // Transfers relevant to this bucket
  const bucketTransfers = allTransfers.filter((t) =>
    t.from_bucket_id === bucketId || t.to_bucket_id === bucketId,
  )

  // Unified sorted list: transactions + transfers
  const listItems: ListItem[] = [
    ...transactions.map((tx): ListItem => ({ kind: 'transaction', tx })),
    ...bucketTransfers.map((t): ListItem => {
      const isOutgoing = t.from_bucket_id === bucketId
      return { kind: 'transfer', t, amountCents: isOutgoing ? -t.amount_cents : t.amount_cents }
    }),
  ].sort((a, b) => {
    const dateA = a.kind === 'transaction' ? a.tx.created_at : a.t.created_at
    const dateB = b.kind === 'transaction' ? b.tx.created_at : b.t.created_at
    return new Date(dateB).getTime() - new Date(dateA).getTime()
  })

  const deleteBucket = useMutation({
    mutationFn: () => api.deleteBucket(bucketId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buckets'] })
      navigate('/', { replace: true })
    },
  })

  const isDesktop = useMediaQuery('(min-width: 768px)')
  const handleDeleteClose = () => setConfirmDelete(false)
  const { handleRef: deleteHandleRef, sheetStyle: deleteSheetStyle, backdropStyle: deleteBackdropStyle, onAnimationEnd: deleteOnAnimationEnd } = useDraggableSheet({ onClose: handleDeleteClose, isOpen: confirmDelete })

  if (!bucket && buckets.length > 0) {
    navigate('/', { replace: true })
    return null
  }

  return (
    <div style={{ minHeight: '100%' }}>
      {/* Header */}
      <div
        style={{
          padding: '20px 20px 0',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 20 }}>
          <button
            onClick={() => navigate('/')}
            style={{
              background: 'var(--surface)',
              border: 'none',
              borderRadius: 10,
              width: 36,
              height: 36,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: 'pointer',
              color: 'var(--text)',
              fontSize: 18,
              flexShrink: 0,
            }}
          >
            ←
          </button>
          <div style={{ flex: 1, minWidth: 0 }}>
            <h1
              className="line-clamp-1"
              style={{
                fontFamily: 'Syne',
                fontWeight: 800,
                fontSize: 22,
                color: 'var(--text)',
                lineHeight: 1.1,
              }}
            >
              {bucket?.name ?? '…'}
            </h1>
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            <button
              onClick={() => setShowTransfer(true)}
              style={{
                background: 'var(--surface)',
                border: 'none',
                borderRadius: 10,
                width: 36,
                height: 36,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                cursor: 'pointer',
                fontSize: 16,
              }}
            >
              ⇄
            </button>
            {!bucket?.is_general && (
              <button
                onClick={() => setConfirmDelete(true)}
                style={{
                  background: 'rgba(248,113,113,0.1)',
                  border: 'none',
                  borderRadius: 10,
                  width: 36,
                  height: 36,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  cursor: 'pointer',
                  fontSize: 16,
                }}
              >
                🗑
              </button>
            )}
          </div>
        </div>

      </div>

      {/* Transaction list */}
      <div style={{ padding: '8px 20px 24px' }}>
        <p
          style={{
            fontFamily: 'Syne',
            fontWeight: 700,
            fontSize: 12,
            letterSpacing: '0.08em',
            color: 'var(--text-2)',
            marginBottom: 4,
            marginTop: 16,
          }}
        >
          TRANSACTIONS · {listItems.length}
        </p>

        {isLoading ? (
          Array.from({ length: 5 }).map((_, i) => (
            <div
              key={i}
              style={{
                display: 'flex',
                gap: 14,
                padding: '14px 0',
                borderBottom: '1px solid var(--border)',
                alignItems: 'center',
              }}
            >
              <div className="shimmer" style={{ width: 44, height: 44, borderRadius: 14, flexShrink: 0 }} />
              <div style={{ flex: 1 }}>
                <div className="shimmer" style={{ width: '60%', height: 14, marginBottom: 8 }} />
                <div className="shimmer" style={{ width: '30%', height: 12 }} />
              </div>
              <div className="shimmer" style={{ width: 65, height: 14 }} />
            </div>
          ))
        ) : listItems.length === 0 ? (
          <div style={{ padding: '40px 0', textAlign: 'center', color: 'var(--text-2)' }}>
            <p style={{ fontSize: 32, marginBottom: 10 }}>◻</p>
            <p style={{ fontFamily: 'DM Sans', fontSize: 15 }}>No transactions assigned</p>
          </div>
        ) : (
          listItems.map((item) => {
            if (item.kind === 'transaction') {
              const tx = item.tx
              const isDebit = tx.amount_cents < 0
              return (
                <button
                  key={`tx-${tx.transaction_id}`}
                  onClick={() => setAssignTx(tx)}
                  className="pressable"
                  style={{
                    width: '100%',
                    display: 'flex',
                    alignItems: 'center',
                    gap: 14,
                    padding: '14px 0',
                    background: 'transparent',
                    border: 'none',
                    borderBottom: '1px solid var(--border)',
                    cursor: 'pointer',
                    textAlign: 'left',
                  }}
                >
                  <div
                    style={{
                      width: 44,
                      height: 44,
                      borderRadius: 14,
                      background: isDebit ? 'rgba(248,113,113,0.1)' : 'rgba(74,222,128,0.1)',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      flexShrink: 0,
                      fontSize: 20,
                      color: isDebit ? 'var(--red)' : 'var(--green)',
                    }}
                  >
                    {isDebit ? '↓' : '↑'}
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p className="line-clamp-1" style={{ fontSize: 15, color: 'var(--text)', fontWeight: 500 }}>
                      {tx.description}
                    </p>
                    <p style={{ fontSize: 12, color: 'var(--text-2)', marginTop: 3 }}>
                      {formatDate(tx.created_at)}
                    </p>
                  </div>
                  <span
                    className={isDebit ? 'amount-negative' : 'amount-positive'}
                    style={{ fontSize: 15, fontWeight: 600, flexShrink: 0 }}
                  >
                    {isDebit ? '−' : '+'}{formatAUD(tx.amount_cents)}
                  </span>
                </button>
              )
            }

            // Transfer item
            const { t, amountCents } = item
            const isDebit = amountCents < 0
            const otherName = isDebit
              ? (t.to_bucket_name || 'General')
              : (t.from_bucket_name || 'General')
            const label = isDebit ? `Transfer to ${otherName}` : `Transfer from ${otherName}`
            return (
              <div
                key={`tr-${t.transfer_id}`}
                style={{
                  width: '100%',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 14,
                  padding: '14px 0',
                  borderBottom: '1px solid var(--border)',
                }}
              >
                <div
                  style={{
                    width: 44,
                    height: 44,
                    borderRadius: 14,
                    background: 'rgba(202,255,51,0.08)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexShrink: 0,
                    fontSize: 18,
                    color: 'var(--accent)',
                  }}
                >
                  ⇄
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <p className="line-clamp-1" style={{ fontSize: 15, color: 'var(--text)', fontWeight: 500 }}>
                    {label}
                  </p>
                  {t.note && (
                    <p className="line-clamp-1" style={{ fontSize: 12, color: 'var(--text-3)', marginTop: 2 }}>{t.note}</p>
                  )}
                  <p style={{ fontSize: 12, color: 'var(--text-2)', marginTop: 3 }}>
                    {formatDate(t.created_at)}
                  </p>
                </div>
                <span
                  className={isDebit ? 'amount-negative' : 'amount-positive'}
                  style={{ fontSize: 15, fontWeight: 600, flexShrink: 0 }}
                >
                  {isDebit ? '−' : '+'}{formatAUD(Math.abs(amountCents))}
                </span>
              </div>
            )
          })
        )}

        {/* Balance card */}
        <div
          style={{
            background: 'var(--surface)',
            borderRadius: 16,
            padding: '16px 20px',
            marginTop: 16,
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
          }}
        >
          <p
            style={{
              fontFamily: 'Syne',
              fontWeight: 700,
              fontSize: 12,
              letterSpacing: '0.08em',
              color: 'var(--text-2)',
            }}
          >
            BALANCE
          </p>
          <p className="amount-neutral" style={{ fontSize: 24, fontWeight: 600 }}>
            {bucket
              ? `${bucket.balance_cents < 0 ? '−' : ''}${formatAUD(Math.abs(bucket.balance_cents))}`
              : '—'}
          </p>
        </div>
      </div>

      {/* Delete confirmation sheet — portalled to body to clear nav tabs */}
      {confirmDelete && createPortal(
        <>
          <div
            onClick={handleDeleteClose}
            style={{
              position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)',
              zIndex: 100, animation: 'fadeIn 0.2s ease forwards',
              ...deleteBackdropStyle,
            }}
          />
          <div
            onAnimationEnd={deleteOnAnimationEnd}
            style={isDesktop ? {
              position: 'fixed', top: '50%', left: '50%',
              transform: 'translate(-50%, -50%)',
              width: 480, zIndex: 101,
              background: 'var(--surface)', borderRadius: 20,
              padding: '24px', animation: 'fadeIn 0.2s ease forwards',
              maxHeight: '85vh', overflowY: 'auto',
            } : {
              position: 'fixed', bottom: 0, left: 0, right: 0, zIndex: 101,
              background: 'var(--surface)',
              borderRadius: '20px 20px 0 0',
              padding: '20px 20px',
              paddingBottom: 24,
              ...deleteSheetStyle,
            }}
          >
            {!isDesktop && (
              <div ref={deleteHandleRef} style={{ touchAction: 'none', display: 'flex', justifyContent: 'center', marginBottom: 16 }}>
                <div style={{ width: 36, height: 4, background: 'var(--border)', borderRadius: 2 }} />
              </div>
            )}
            <p
              style={{
                fontFamily: 'Syne', fontWeight: 800, fontSize: 20,
                color: 'var(--text)', marginBottom: 8,
              }}
            >
              Delete "{bucket?.name}"?
            </p>
            <p style={{ fontFamily: 'DM Sans', fontSize: 14, color: 'var(--text-2)', marginBottom: 20 }}>
              Transactions will be unassigned. This cannot be undone.
            </p>
            <button
              onClick={() => deleteBucket.mutate()}
              disabled={deleteBucket.isPending}
              className="pressable"
              style={{
                width: '100%',
                padding: '15px',
                background: 'rgba(248,113,113,0.15)',
                border: '1px solid rgba(248,113,113,0.3)',
                borderRadius: 12,
                color: 'var(--red)',
                fontFamily: 'Syne',
                fontWeight: 700,
                fontSize: 15,
                cursor: 'pointer',
                marginBottom: 10,
              }}
            >
              {deleteBucket.isPending ? 'Deleting…' : 'Delete Bucket'}
            </button>
            <button
              onClick={() => setConfirmDelete(false)}
              style={{
                width: '100%',
                padding: '15px',
                background: 'transparent',
                border: 'none',
                color: 'var(--text-2)',
                fontFamily: 'DM Sans',
                fontSize: 15,
                cursor: 'pointer',
              }}
            >
              Cancel
            </button>
          </div>
        </>,
        document.body,
      )}

      {assignTx && (
        <AssignSheet transaction={assignTx} onClose={() => setAssignTx(null)} />
      )}
      {showTransfer && bucketId && (
        <TransferSheet initialFromBucketId={bucketId} onClose={() => setShowTransfer(false)} />
      )}
    </div>
  )
}
