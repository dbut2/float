import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api, type Trickle } from '../lib/api'
import { useDraggableSheet } from '../hooks/useDraggableSheet'
import { useMediaQuery } from '../hooks/useMediaQuery'

interface Props {
  bucketId: string
  trickle?: Trickle | null
  onClose: () => void
}

const tomorrow = new Date(Date.now() + 86400000).toISOString().slice(0, 10)

export default function TrickleSheet({ bucketId, trickle, onClose }: Props) {
  const qc = useQueryClient()

  const [description, setDescription] = useState(trickle?.description ?? '')
  const [amountStr, setAmountStr] = useState(trickle ? String(trickle.amount_cents / 100) : '')
  const [period, setPeriod] = useState<string>(trickle?.period ?? 'monthly')
  const [startDate, setStartDate] = useState(tomorrow)
  const [endDate, setEndDate] = useState(trickle?.end_date ? trickle.end_date.slice(0, 10) : '')

  const amountCents = Math.round(parseFloat(amountStr) * 100)
  const canSubmit = !isNaN(amountCents) && amountCents > 0 && period !== '' && startDate !== ''

  const upsert = useMutation({
    mutationFn: () =>
      api.upsertTrickle(bucketId, {
        amount_cents: amountCents,
        description: description.trim(),
        period,
        start_date: startDate,
        end_date: endDate || null,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buckets'] })
      qc.invalidateQueries({ queryKey: ['trickle', bucketId] })
      onClose()
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
    appearance: 'none' as const,
    boxSizing: 'border-box' as const,
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
          {trickle ? 'Edit Trickle' : 'Add Trickle'}
        </p>

        <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 11, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 6 }}>DESCRIPTION (OPTIONAL)</p>
        <input
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="e.g. Monthly savings"
          style={{ ...inputStyle, marginBottom: 12 }}
        />

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
            style={{ ...inputStyle, paddingLeft: 30 }}
          />
        </div>

        <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 11, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 6 }}>PERIOD</p>
        <select
          value={period}
          onChange={(e) => setPeriod(e.target.value)}
          style={{ ...inputStyle, marginBottom: 12 }}
        >
          <option value="daily">Daily</option>
          <option value="weekly">Weekly</option>
          <option value="fortnightly">Fortnightly</option>
          <option value="monthly">Monthly</option>
        </select>

        <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 11, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 6 }}>START DATE</p>
        <input
          type="date"
          value={startDate}
          min={tomorrow}
          onChange={(e) => setStartDate(e.target.value)}
          style={{ ...inputStyle, marginBottom: 12 }}
        />

        <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 11, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 6 }}>END DATE <span style={{ fontWeight: 400, textTransform: 'none', letterSpacing: 0 }}>(optional — leave blank for never)</span></p>
        <input
          type="date"
          value={endDate}
          onChange={(e) => setEndDate(e.target.value)}
          style={{ ...inputStyle, marginBottom: 20 }}
        />

        <button
          onClick={() => upsert.mutate()}
          disabled={!canSubmit || upsert.isPending}
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
          {upsert.isPending ? 'Saving…' : trickle ? 'Update Trickle' : 'Add Trickle'}
        </button>
      </div>
    </>
  )
}
