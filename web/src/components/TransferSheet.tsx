import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { useDraggableSheet } from '../hooks/useDraggableSheet'
import { useMediaQuery } from '../hooks/useMediaQuery'

interface Props {
  initialFromBucketId?: string
  onClose: () => void
}

export default function TransferSheet({ initialFromBucketId, onClose }: Props) {
  const qc = useQueryClient()
  const { data: buckets = [] } = useQuery({ queryKey: ['buckets'], queryFn: api.getBuckets })

  const options = buckets

  const [fromId, setFromId] = useState<string | null>(initialFromBucketId ?? null)
  const [toId, setToId] = useState<string | null>(null)
  const [amountStr, setAmountStr] = useState('')
  const [note, setNote] = useState('')

  const toOptions = useMemo(() => options.filter((b) => b.bucket_id !== fromId), [options, fromId])

  const handleFromChange = (id: string) => {
    setFromId(id)
    if (toId === id) setToId(null)
  }

  const amountCents = Math.round(parseFloat(amountStr) * 100)
  const canSubmit = !isNaN(amountCents) && amountCents > 0 && fromId !== null && toId !== null && fromId !== toId

  const create = useMutation({
    mutationFn: () =>
      api.createTransfer(
        fromId!,
        toId!,
        amountCents,
        note.trim(),
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buckets'] })
      qc.invalidateQueries({ queryKey: ['transfers'] })
      onClose()
    },
  })

  const isDesktop = useMediaQuery('(min-width: 768px)')
  const { handleRef, sheetStyle, backdropStyle, onAnimationEnd } = useDraggableSheet({ onClose })

  const selectStyle = {
    width: '100%',
    background: 'var(--surface-2)',
    border: '1px solid var(--border)',
    borderRadius: 12,
    padding: '14px 16px',
    color: 'var(--text)',
    fontFamily: 'DM Sans',
    fontSize: 16,
    outline: 'none',
    appearance: 'none' as const,
  }

  return (
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
          ...sheetStyle,
        }}
      >
        {!isDesktop && (
          <div ref={handleRef} style={{ touchAction: 'none', display: 'flex', justifyContent: 'center', marginBottom: 20 }}>
            <div style={{ width: 36, height: 4, background: 'var(--border)', borderRadius: 2 }} />
          </div>
        )}
        <p style={{ fontFamily: 'Syne', fontWeight: 800, fontSize: 20, color: 'var(--text)', marginBottom: 16 }}>
          New Transfer
        </p>

        {/* From */}
        <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 11, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 6 }}>FROM</p>
        <select
          value={fromId ?? ''}
          onChange={(e) => handleFromChange(e.target.value)}
          style={{ ...selectStyle, marginBottom: 12 }}
        >
          {fromId === null && <option value="" disabled>Select bucket…</option>}
          {options.map((b) => (
            <option key={b.bucket_id} value={b.bucket_id}>{b.name}</option>
          ))}
        </select>

        {/* To */}
        <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 11, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 6 }}>TO</p>
        <select
          value={toId ?? ''}
          onChange={(e) => setToId(e.target.value || null)}
          style={{ ...selectStyle, marginBottom: 12 }}
        >
          <option value="" disabled>Select bucket…</option>
          {toOptions.map((b) => (
            <option key={b.bucket_id} value={b.bucket_id}>{b.name}</option>
          ))}
        </select>

        {/* Amount */}
        <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 11, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 6 }}>AMOUNT</p>
        <div style={{ position: 'relative', marginBottom: 12 }}>
          <span style={{ position: 'absolute', left: 16, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-2)', fontFamily: 'DM Sans', fontSize: 16 }}>$</span>
          <input
            type="number"
            min="0.01"
            step="0.01"
            value={amountStr}
            onChange={(e) => setAmountStr(e.target.value)}
            placeholder="0.00"
            style={{ ...selectStyle, paddingLeft: 30 }}
          />
        </div>

        {/* Note */}
        <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 11, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 6 }}>NOTE (OPTIONAL)</p>
        <input
          value={note}
          onChange={(e) => setNote(e.target.value)}
          placeholder="e.g. monthly savings"
          style={{ ...selectStyle, marginBottom: 20 }}
        />

        <button
          onClick={() => create.mutate()}
          disabled={!canSubmit || create.isPending}
          className="pressable"
          style={{
            width: '100%',
            padding: '15px',
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
          {create.isPending ? 'Transferring…' : 'Transfer'}
        </button>
      </div>
    </>
  )
}
