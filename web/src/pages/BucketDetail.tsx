import { useState } from 'react'
import { createPortal } from 'react-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams, useNavigate } from 'react-router-dom'
import { ArrowDown, ArrowLeftRight, ArrowUp, ChevronLeft, Filter, Inbox, RotateCw, Trash2, X } from 'lucide-react'
import { api, formatAUD, formatDate, type Transaction, type Transfer, type Trickle } from '../lib/api'
import AssignSheet from '../components/AssignSheet'
import RulesSheet from '../components/RulesSheet'
import TransferSheet from '../components/TransferSheet'
import TrickleSheet from '../components/TrickleSheet'
import { useDraggableSheet } from '../hooks/useDraggableSheet'
import { useMediaQuery } from '../hooks/useMediaQuery'

type ListItem =
  | { kind: 'transaction'; tx: Transaction }
  | { kind: 'transfer'; t: Transfer; amountCents: number }
  | { kind: 'trickle'; tx: Transaction }

function isTrickleEntry(tx: Transaction, bucketTransfers: Transfer[]): boolean {
  if (tx.is_transaction) return false
  const txTime = new Date(tx.created_at).getTime()
  return !bucketTransfers.some((t) => Math.abs(new Date(t.created_at).getTime() - txTime) < 1000)
}

export default function BucketDetail() {
  const { id } = useParams<{ id: string }>()
  const bucketId = id!
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [assignTx, setAssignTx] = useState<Transaction | null>(null)
  const [confirmDelete, setConfirmDelete] = useState(false)
  const [showTransfer, setShowTransfer] = useState(false)
  const [showTrickle, setShowTrickle] = useState(false)
  const [showRules, setShowRules] = useState(false)

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

  const { data: trickle = null } = useQuery({
    queryKey: ['trickle', bucketId],
    queryFn: () => api.getTrickle(bucketId).catch((e: Error) => {
      if (e.message.includes('404') || e.message.toLowerCase().includes('not found')) return null
      throw e
    }),
    enabled: !!bucketId,
  })

  // Transfers relevant to this bucket
  const bucketTransfers = allTransfers.filter((t) =>
    t.from_bucket_id === bucketId || t.to_bucket_id === bucketId,
  )

  const listItems: ListItem[] = [
    ...bucketTx.map((tx): ListItem => {
      if (!tx.is_transaction && isTrickleEntry(tx, bucketTransfers)) {
        return { kind: 'trickle', tx }
      }
      return { kind: 'transaction', tx }
    }),
    ...bucketTransfers.map((t): ListItem => {
      const isOutgoing = t.from_bucket_id === bucketId
      return { kind: 'transfer', t, amountCents: isOutgoing ? -t.amount_cents : t.amount_cents }
    }),
  ].sort((a, b) => {
    const dateA = a.kind === 'transaction' || a.kind === 'trickle' ? a.tx.created_at : a.t.created_at
    const dateB = b.kind === 'transaction' || b.kind === 'trickle' ? b.tx.created_at : b.t.created_at
    return new Date(dateB).getTime() - new Date(dateA).getTime()
  })

  const deleteBucket = useMutation({
    mutationFn: () => api.deleteBucket(bucketId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buckets'] })
      navigate('/', { replace: true })
    },
  })

  const deleteTrickleMutation = useMutation({
    mutationFn: () => api.deleteTrickle(bucketId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buckets'] })
      qc.invalidateQueries({ queryKey: ['trickle', bucketId] })
    },
  })

  const isDesktop = useMediaQuery('(min-width: 768px)')
  const handleDeleteClose = () => setConfirmDelete(false)
  const { handleRef: deleteHandleRef, sheetStyle: deleteSheetStyle, backdropStyle: deleteBackdropStyle, onAnimationEnd: deleteOnAnimationEnd } = useDraggableSheet({ onClose: handleDeleteClose, isOpen: confirmDelete })

  if (!bucket && buckets.length > 0) {
    navigate('/', { replace: true })
    return null
  }

  const periodLabel = (t: Trickle) => {
    const map: Record<string, string> = { daily: 'Daily', weekly: 'Weekly', fortnightly: 'Fortnightly', monthly: 'Monthly' }
    return map[t.period] ?? t.period
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
              flexShrink: 0,
            }}
          >
            <ChevronLeft size={18} strokeWidth={1.75} />
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
            {!bucket?.is_general && (
              <button
                onClick={() => setShowRules(true)}
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
                }}
                title="Manage rules"
              >
                <Filter size={16} strokeWidth={1.75} />
              </button>
            )}
            {!bucket?.is_general && (
              <button
                onClick={() => setShowTrickle(true)}
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
                }}
                title="Manage trickle"
              >
                <RotateCw size={16} strokeWidth={1.75} />
              </button>
            )}
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
              }}
            >
              <ArrowLeftRight size={16} strokeWidth={1.75} />
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
                }}
              >
                <Trash2 size={16} strokeWidth={1.75} color="var(--red)" />
              </button>
            )}
          </div>
        </div>

        {!bucket?.is_general && (
          <div
            style={{
              background: 'var(--surface)',
              borderRadius: 16,
              padding: '14px 16px',
              marginBottom: 16,
            }}
          >
            {trickle ? (
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <RotateCw size={20} color="var(--accent)" strokeWidth={1.75} style={{ flexShrink: 0 }} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <p style={{ fontSize: 14, color: 'var(--text)', fontWeight: 500 }}>
                    {formatAUD(trickle.amount_cents)} · {periodLabel(trickle)}
                  </p>
                  {trickle.description && (
                    <p className="line-clamp-1" style={{ fontSize: 12, color: 'var(--text-3)', marginTop: 2 }}>{trickle.description}</p>
                  )}
                  <p style={{ fontSize: 12, color: 'var(--text-2)', marginTop: 2 }}>
                    From {trickle.from_bucket_name} · starts {trickle.start_date.slice(0, 10)}{trickle.end_date ? ` · ends ${trickle.end_date.slice(0, 10)}` : ''}
                  </p>
                </div>
                <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
                  <button
                    onClick={() => setShowTrickle(true)}
                    style={{
                      background: 'var(--surface-2)',
                      border: 'none',
                      borderRadius: 8,
                      padding: '6px 12px',
                      color: 'var(--text)',
                      fontFamily: 'DM Sans',
                      fontSize: 13,
                      cursor: 'pointer',
                    }}
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => deleteTrickleMutation.mutate()}
                    disabled={deleteTrickleMutation.isPending}
                    style={{
                      background: 'rgba(248,113,113,0.1)',
                      border: 'none',
                      borderRadius: 8,
                      padding: '6px 10px',
                      color: 'var(--red)',
                      fontFamily: 'DM Sans',
                      fontSize: 13,
                      cursor: 'pointer',
                    }}
                  >
                    {deleteTrickleMutation.isPending ? '…' : <X size={13} strokeWidth={1.75} />}
                  </button>
                </div>
              </div>
            ) : (
              <button
                onClick={() => setShowTrickle(true)}
                style={{
                  width: '100%',
                  background: 'transparent',
                  border: 'none',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 10,
                  cursor: 'pointer',
                  padding: 0,
                }}
              >
                <RotateCw size={20} color="var(--text-2)" strokeWidth={1.75} />
                <span style={{ fontFamily: 'DM Sans', fontSize: 14, color: 'var(--text-2)' }}>Add Trickle</span>
              </button>
            )}
          </div>
        )}
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
            <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 10 }}>
              <Inbox size={32} color="var(--text-2)" strokeWidth={1.75} />
            </div>
            <p style={{ fontFamily: 'DM Sans', fontSize: 15 }}>No transactions assigned</p>
          </div>
        ) : (
          listItems.map((item, idx) => {
            if (item.kind === 'transaction') {
              const tx = item.tx
              const isDebit = tx.amount_cents < 0
              return (
                <button
                  key={`tx-${idx}`}
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
                    }}
                  >
                    {isDebit ? <ArrowDown size={20} color="var(--red)" strokeWidth={1.75} /> : <ArrowUp size={20} color="var(--green)" strokeWidth={1.75} />}
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

            if (item.kind === 'trickle') {
              const tx = item.tx
              const isDebit = tx.amount_cents < 0
              const label = tx.description
              const txDate = new Date(tx.created_at)
              const isActive = Math.abs(Date.now() - txDate.getTime()) < 300000
              return (
                <div
                  key={`trickle-${idx}`}
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
                    }}
                  >
                    <RotateCw size={18} color="var(--accent)" strokeWidth={1.75} />
                  </div>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p className="line-clamp-1" style={{ fontSize: 15, color: 'var(--text)', fontWeight: 500 }}>
                      {label}
                    </p>
                    {tx.description && (
                      <p className="line-clamp-1" style={{ fontSize: 12, color: 'var(--text-3)', marginTop: 2 }}>{tx.description}</p>
                    )}
                    {!isActive && (
                      <p style={{ fontSize: 12, color: 'var(--text-2)', marginTop: 3 }}>
                        {formatDate(tx.created_at)}
                      </p>
                    )}
                  </div>
                  <span
                    className={isDebit ? 'amount-negative' : 'amount-positive'}
                    style={{ fontSize: 15, fontWeight: 600, flexShrink: 0 }}
                  >
                    {isDebit ? '−' : '+'}{formatAUD(Math.abs(tx.amount_cents))}
                  </span>
                </div>
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
                  }}
                >
                  <ArrowLeftRight size={18} color="var(--accent)" strokeWidth={1.75} />
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
      {showTrickle && bucketId && (
        <TrickleSheet bucketId={bucketId} trickle={trickle} onClose={() => setShowTrickle(false)} />
      )}
      {showRules && bucketId && (
        <RulesSheet bucketId={bucketId} onClose={() => setShowRules(false)} />
      )}
    </div>
  )
}
