import { useState } from 'react'
import { createPortal } from 'react-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeftRight, Trash2, X } from 'lucide-react'
import { api, type Transaction } from '../lib/api'
import { useDraggableSheet } from '../hooks/useDraggableSheet'
import { useMediaQuery } from '../hooks/useMediaQuery'
import AssignSheet from './AssignSheet'

interface Props {
  transaction: Transaction
  onClose: () => void
}

export default function CoverSheet({ transaction, onClose }: Props) {
  const qc = useQueryClient()
  const [amountStr, setAmountStr] = useState('')
  const [note, setNote] = useState('')
  const [showAssign, setShowAssign] = useState(false)

  const covers = transaction.covers ?? []
  const originalAmount = Math.abs(transaction.amount_cents)
  const coveredAmount = transaction.covers_amount_cents ?? 0
  const netAmount = transaction.net_amount_cents ?? transaction.amount_cents
  const netDisplay = transaction.net_display_amount ?? transaction.display_amount

  const amountCents = Math.round(parseFloat(amountStr) * 100)
  const remaining = originalAmount - coveredAmount
  const canSubmit = !isNaN(amountCents) && amountCents > 0 && amountCents <= remaining

  const addCover = useMutation({
    mutationFn: () => api.createCover(transaction.transaction_id, amountCents, note.trim()),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['bucket-transactions', transaction.bucket_id] })
      qc.invalidateQueries({ queryKey: ['buckets'] })
      setAmountStr('')
      setNote('')
    },
  })

  const deleteCover = useMutation({
    mutationFn: (coverId: string) => api.deleteCover(coverId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['bucket-transactions', transaction.bucket_id] })
      qc.invalidateQueries({ queryKey: ['buckets'] })
    },
  })

  const isDesktop = useMediaQuery('(min-width: 768px)')
  const { handleRef, sheetStyle, backdropStyle, onAnimationEnd } = useDraggableSheet({ onClose })

  const inputStyle = {
    width: '100%',
    background: 'var(--surface-2)',
    border: '1px solid var(--border)',
    borderRadius: 12,
    padding: '14px 16px',
    color: 'var(--text)',
    fontFamily: 'DM Sans',
    fontSize: 16,
    outline: 'none',
  }

  const sheet = (
    <>
      <div
        onClick={onClose}
        style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', zIndex: 100, animation: 'fadeIn 0.2s ease forwards', ...backdropStyle }}
      />
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
          position: 'fixed', bottom: 0, left: 0, right: 0, zIndex: 101,
          background: 'var(--surface)',
          borderRadius: '20px 20px 0 0',
          padding: '16px 20px',
          paddingBottom: 24,
          maxHeight: '85vh', overflowY: 'auto',
          ...sheetStyle,
        }}
      >
        {!isDesktop && (
          <div ref={handleRef} style={{ touchAction: 'none', display: 'flex', justifyContent: 'center', marginBottom: 20 }}>
            <div style={{ width: 36, height: 4, background: 'var(--border)', borderRadius: 2 }} />
          </div>
        )}

        {/* Header */}
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 16 }}>
          <div style={{ flex: 1, minWidth: 0, paddingRight: 12 }}>
            <p style={{ fontFamily: 'Syne', fontWeight: 800, fontSize: 20, color: 'var(--text)', marginBottom: 4 }}>
              Covers
            </p>
            <p className="line-clamp-1" style={{ fontSize: 14, color: 'var(--text-2)' }}>{transaction.description}</p>
          </div>
          <button
            onClick={onClose}
            style={{ background: 'var(--surface-2)', border: 'none', borderRadius: 10, width: 32, height: 32, display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', flexShrink: 0 }}
          >
            <X size={14} color="var(--text-2)" strokeWidth={1.75} />
          </button>
        </div>

        {/* Amount summary */}
        <div style={{ background: 'var(--surface-2)', borderRadius: 14, padding: '14px 16px', marginBottom: 20 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: coveredAmount > 0 ? 8 : 0 }}>
            <span style={{ fontSize: 13, color: 'var(--text-2)', fontFamily: 'DM Sans' }}>Original</span>
            <span className="amount-negative" style={{ fontSize: 13, fontWeight: 600 }}>{transaction.display_amount}</span>
          </div>
          {coveredAmount > 0 && (
            <>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                <span style={{ fontSize: 13, color: 'var(--text-2)', fontFamily: 'DM Sans' }}>Covered</span>
                <span className="amount-positive" style={{ fontSize: 13, fontWeight: 600 }}>+${(coveredAmount / 100).toFixed(2)}</span>
              </div>
              <div style={{ borderTop: '1px solid var(--border)', paddingTop: 8, display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ fontSize: 13, color: 'var(--text)', fontFamily: 'DM Sans', fontWeight: 600 }}>Net</span>
                <span className={netAmount < 0 ? 'amount-negative' : 'amount-positive'} style={{ fontSize: 13, fontWeight: 700 }}>{netDisplay}</span>
              </div>
            </>
          )}
        </div>

        {/* Existing covers */}
        {covers.length > 0 && (
          <div style={{ marginBottom: 20 }}>
            <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 11, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 8 }}>COVERS</p>
            {covers.map((cover) => (
              <div key={cover.cover_id} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '10px 0', borderBottom: '1px solid var(--border)' }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <span className="amount-positive" style={{ fontSize: 14, fontWeight: 600 }}>+${(cover.amount_cents / 100).toFixed(2)}</span>
                  {cover.note && (
                    <p className="line-clamp-1" style={{ fontSize: 12, color: 'var(--text-2)', marginTop: 2 }}>{cover.note}</p>
                  )}
                  <p style={{ fontSize: 11, color: 'var(--text-3)', marginTop: 2 }}>{cover.display_date}</p>
                </div>
                <button
                  onClick={() => deleteCover.mutate(cover.cover_id)}
                  disabled={deleteCover.isPending}
                  style={{ background: 'rgba(248,113,113,0.1)', border: 'none', borderRadius: 8, padding: '6px 8px', color: 'var(--red)', cursor: 'pointer', display: 'flex', alignItems: 'center', flexShrink: 0 }}
                >
                  <Trash2 size={13} strokeWidth={1.75} />
                </button>
              </div>
            ))}
          </div>
        )}

        {/* Add cover form */}
        {remaining > 0 && (
          <div style={{ marginBottom: 20 }}>
            <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 11, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 8 }}>
              {covers.length > 0 ? 'ADD ANOTHER COVER' : 'ADD COVER'}
            </p>
            <div style={{ position: 'relative', marginBottom: 10 }}>
              <span style={{ position: 'absolute', left: 16, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-2)', fontFamily: 'DM Sans', fontSize: 16 }}>$</span>
              <input
                type="number"
                min="0.01"
                step="0.01"
                max={(remaining / 100).toFixed(2)}
                value={amountStr}
                onChange={(e) => setAmountStr(e.target.value)}
                placeholder={`0.00 (max $${(remaining / 100).toFixed(2)})`}
                style={{ ...inputStyle, paddingLeft: 30 }}
              />
            </div>
            <input
              value={note}
              onChange={(e) => setNote(e.target.value)}
              placeholder="Note (optional)"
              style={{ ...inputStyle, marginBottom: 10 }}
            />
            <button
              onClick={() => addCover.mutate()}
              disabled={!canSubmit || addCover.isPending}
              className="pressable"
              style={{
                width: '100%',
                padding: '13px',
                background: canSubmit ? 'var(--accent)' : 'var(--surface-2)',
                border: 'none',
                borderRadius: 12,
                color: canSubmit ? '#000' : 'var(--text-2)',
                fontFamily: 'Syne',
                fontWeight: 700,
                fontSize: 15,
                cursor: canSubmit ? 'pointer' : 'default',
                transition: 'all 0.15s',
              }}
            >
              {addCover.isPending ? 'Adding…' : 'Add Cover'}
            </button>
          </div>
        )}

        {/* Assign to bucket */}
        <button
          onClick={() => setShowAssign(true)}
          style={{
            width: '100%',
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            background: 'transparent',
            border: 'none',
            padding: '10px 0',
            cursor: 'pointer',
          }}
        >
          <ArrowLeftRight size={16} color="var(--text-2)" strokeWidth={1.75} />
          <span style={{ fontFamily: 'DM Sans', fontSize: 14, color: 'var(--text-2)' }}>Assign to different bucket</span>
        </button>
      </div>

      {showAssign && (
        <AssignSheet transaction={transaction} onClose={() => setShowAssign(false)} />
      )}
    </>
  )

  return createPortal(sheet, document.body)
}
