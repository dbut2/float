import { useEffect, useRef, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { ChevronRight, GripVertical, Plus } from 'lucide-react'
import { type Bucket, api, formatAUD } from '../lib/api'
import { useDraggableSheet } from '../hooks/useDraggableSheet'
import { useMediaQuery } from '../hooks/useMediaQuery'

function BucketCard({
  name,
  balanceCents,
  onClick,
  dragHandle,
  isDragging,
}: {
  name: string
  balanceCents: number
  onClick: () => void
  dragHandle?: React.ReactNode
  isDragging?: boolean
}) {
  const isNeg = balanceCents < 0
  return (
    <div
      style={{
        width: '100%',
        background: 'var(--surface)',
        border: '1px solid var(--border)',
        borderRadius: 20,
        padding: '22px 24px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: 12,
        opacity: isDragging ? 0.5 : 1,
        transition: 'opacity 0.1s',
      }}
    >
      {dragHandle}
      <button
        onClick={onClick}
        className="pressable"
        style={{
          flex: 1,
          background: 'none',
          border: 'none',
          padding: 0,
          cursor: 'pointer',
          textAlign: 'left',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <div>
          <p
            style={{
              fontFamily: 'Syne',
              fontWeight: 700,
              fontSize: 13,
              letterSpacing: '0.04em',
              color: 'var(--text-2)',
              marginBottom: 10,
            }}
          >
            {name}
          </p>
          <p
            className={isNeg ? 'amount-negative' : 'amount-neutral'}
            style={{ fontSize: 28, fontWeight: 600, lineHeight: 1 }}
          >
            {isNeg ? '−' : ''}{formatAUD(balanceCents)}
          </p>
        </div>
        <ChevronRight size={20} color="var(--text-3)" strokeWidth={1.75} style={{ flexShrink: 0 }} />
      </button>
    </div>
  )
}

export default function Dashboard() {
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [orderedBuckets, setOrderedBuckets] = useState<Bucket[]>([])
  const dragState = useRef<{ dragIndex: number; startY: number; pointerY: number } | null>(null)
  const [draggingIndex, setDraggingIndex] = useState<number | null>(null)

  const { data: buckets = [], isLoading } = useQuery({
    queryKey: ['buckets'],
    queryFn: api.getBuckets,
  })

  const { data: transactBalance } = useQuery({
    queryKey: ['transact-balance'],
    queryFn: api.getTransactBalance,
  })

  useEffect(() => {
    setOrderedBuckets(buckets)
  }, [buckets])

  const reorder = useMutation({
    mutationFn: (ids: string[]) => api.reorderBuckets(ids),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buckets'] })
    },
  })

  const create = useMutation({
    mutationFn: () => api.createBucket(newName.trim()),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buckets'] })
      setNewName('')
      setShowCreate(false)
    },
  })

  const isDesktop = useMediaQuery('(min-width: 768px)')
  const handleCreateClose = () => { setShowCreate(false); setNewName('') }
  const { handleRef: createHandleRef, sheetStyle: createSheetStyle, backdropStyle: createBackdropStyle, onAnimationEnd: createOnAnimationEnd } = useDraggableSheet({ onClose: handleCreateClose, isOpen: showCreate })

  function onDragHandlePointerDown(e: React.PointerEvent, index: number) {
    e.preventDefault()
    ;(e.currentTarget as HTMLElement).setPointerCapture(e.pointerId)
    dragState.current = { dragIndex: index, startY: e.clientY, pointerY: e.clientY }
    setDraggingIndex(index)
  }

  function onListPointerMove(e: React.PointerEvent) {
    if (!dragState.current) return
    dragState.current.pointerY = e.clientY
    const dy = e.clientY - dragState.current.startY
    const cardHeight = 112 // approximate card height + margin
    const shift = Math.round(dy / cardHeight)
    if (shift === 0) return
    const from = dragState.current.dragIndex
    const to = Math.max(0, Math.min(orderedBuckets.length - 1, from + shift))
    if (to === from) return
    setOrderedBuckets(prev => {
      const next = [...prev]
      const [item] = next.splice(from, 1)
      next.splice(to, 0, item)
      return next
    })
    dragState.current.dragIndex = to
    dragState.current.startY = e.clientY
    setDraggingIndex(to)
  }

  function onListPointerUp() {
    if (!dragState.current) return
    dragState.current = null
    setDraggingIndex(null)
    reorder.mutate(orderedBuckets.map(b => b.bucket_id))
  }

  return (
    <div style={{ padding: '24px 20px', minHeight: '100%' }}>
      <div className="animate-fade-up" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 28, opacity: 0 }}>
        <h1 style={{ fontFamily: 'Syne', fontWeight: 800, fontSize: 22, color: 'var(--text)' }}>
          Buckets
        </h1>
        {transactBalance != null && (
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: 3 }}>
            <span style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 10, letterSpacing: '0.08em', color: 'var(--text-3)', textTransform: 'uppercase' }}>Card</span>
            <span className="amount-neutral" style={{ fontSize: 15 }}>{formatAUD(transactBalance.balance_cents)}</span>
          </div>
        )}
      </div>

      <div
        className="animate-fade-up"
        style={{ opacity: 0 }}
        onPointerMove={onListPointerMove}
        onPointerUp={onListPointerUp}
      >
        {isLoading ? (
          <>
            {Array.from({ length: 3 }).map((_, i) => (
              <div
                key={i}
                className="shimmer"
                style={{ height: 100, borderRadius: 20, marginBottom: 12 }}
              />
            ))}
          </>
        ) : (
          <>
            {orderedBuckets.map((bucket, index) => (
              <BucketCard
                key={bucket.bucket_id}
                name={bucket.name}
                balanceCents={bucket.balance_cents}
                onClick={() => navigate(`/buckets/${bucket.bucket_id}`)}
                isDragging={draggingIndex === index}
                dragHandle={
                  <span
                    onPointerDown={(e) => onDragHandlePointerDown(e, index)}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      paddingRight: 12,
                      cursor: 'grab',
                      touchAction: 'none',
                      flexShrink: 0,
                      color: 'var(--text-3)',
                    }}
                  >
                    <GripVertical size={18} strokeWidth={1.75} />
                  </span>
                }
              />
            ))}
            {/* Add bucket — greyed ghost card at bottom */}
            <button
              onClick={() => setShowCreate(true)}
              className="pressable"
              style={{
                width: '100%',
                background: 'transparent',
                border: '1.5px dashed var(--border)',
                borderRadius: 20,
                padding: '22px 24px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: 12,
                cursor: 'pointer',
                marginBottom: 12,
                opacity: 0.45,
              }}
            >
              <span
                style={{
                  width: 28,
                  height: 28,
                  borderRadius: '50%',
                  border: '1.5px dashed var(--text-3)',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  flexShrink: 0,
                }}
              >
                <Plus size={16} color="var(--text-3)" strokeWidth={1.75} />
              </span>
              <span style={{ fontFamily: 'DM Sans', fontSize: 15, color: 'var(--text-2)' }}>
                New bucket
              </span>
            </button>
          </>
        )}
      </div>

      {/* Create bucket sheet */}
      {showCreate && (
        <>
          <div
            onClick={handleCreateClose}
            style={{
              position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)',
              zIndex: 100, animation: 'fadeIn 0.2s ease forwards',
              ...createBackdropStyle,
            }}
          />
          <div
            onAnimationEnd={createOnAnimationEnd}
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
              ...createSheetStyle,
            }}
          >
            {!isDesktop && (
              <div ref={createHandleRef} style={{ touchAction: 'none', display: 'flex', justifyContent: 'center', marginBottom: 20 }}>
                <div style={{ width: 36, height: 4, background: 'var(--border)', borderRadius: 2 }} />
              </div>
            )}
            <p style={{ fontFamily: 'Syne', fontWeight: 800, fontSize: 20, color: 'var(--text)', marginBottom: 16 }}>
              New Bucket
            </p>
            <input
              autoFocus
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' && newName.trim()) create.mutate()
                if (e.key === 'Escape') { setShowCreate(false); setNewName('') }
              }}
              placeholder="e.g. Groceries, Rent, Travel"
              style={{
                width: '100%',
                background: 'var(--surface-2)',
                border: '1px solid var(--border)',
                borderRadius: 12,
                padding: '14px 16px',
                color: 'var(--text)',
                fontFamily: 'DM Sans',
                fontSize: 16,
                outline: 'none',
                marginBottom: 12,
              }}
            />
            <button
              onClick={() => create.mutate()}
              disabled={!newName.trim() || create.isPending}
              className="pressable"
              style={{
                width: '100%',
                padding: '15px',
                background: newName.trim() ? 'var(--accent)' : 'var(--surface-2)',
                border: 'none',
                borderRadius: 12,
                color: newName.trim() ? '#000' : 'var(--text-2)',
                fontFamily: 'Syne',
                fontWeight: 700,
                fontSize: 15,
                cursor: 'pointer',
                transition: 'all 0.15s',
              }}
            >
              {create.isPending ? 'Creating…' : 'Create Bucket'}
            </button>
          </div>
        </>
      )}
    </div>
  )
}
