import { useState, useEffect, useRef } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Pencil, Sparkles, Check, X } from 'lucide-react'
import { api, type Bucket } from '../lib/api'

export default function Classify() {
  const qc = useQueryClient()
  const [editingId, setEditingId] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  const [polling, setPolling] = useState(false)
  const [status, setStatus] = useState<{ total: number; processed: number; reclassified: number } | null>(null)
  const pollRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const { data: buckets = [], isLoading } = useQuery({
    queryKey: ['buckets'],
    queryFn: api.getBuckets,
  })

  const nonGeneralBuckets = buckets.filter((b) => !b.is_general)

  const updateDesc = useMutation({
    mutationFn: ({ id, desc }: { id: string; desc: string }) =>
      api.updateBucketDescription(id, desc),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buckets'] })
      setEditingId(null)
    },
  })

  const reclassify = useMutation({
    mutationFn: () => api.reclassify(),
    onSuccess: () => {
      setPolling(true)
      setStatus({ total: 0, processed: 0, reclassified: 0 })
    },
  })

  useEffect(() => {
    if (!polling) return
    pollRef.current = setInterval(async () => {
      try {
        const s = await api.reclassifyStatus()
        setStatus(s)
        if (!s.running) {
          setPolling(false)
          qc.invalidateQueries({ queryKey: ['buckets'] })
          qc.invalidateQueries({ queryKey: ['bucket-transactions'] })
        }
      } catch {
        // ignore transient errors
      }
    }, 1500)
    return () => { if (pollRef.current) clearInterval(pollRef.current) }
  }, [polling, qc])

  // Check if there's already a running job on mount
  useEffect(() => {
    api.reclassifyStatus().then((s) => {
      if (s.running) {
        setPolling(true)
        setStatus(s)
      }
    }).catch(() => {})
  }, [])

  function startEdit(b: Bucket) {
    setEditingId(b.bucket_id)
    setEditValue(b.description)
  }

  function saveEdit() {
    if (editingId) {
      updateDesc.mutate({ id: editingId, desc: editValue })
    }
  }

  const isRunning = polling || reclassify.isPending
  const isDone = !isRunning && status !== null && status.processed > 0

  let buttonLabel = 'Reclassify General'
  if (reclassify.isPending) {
    buttonLabel = 'Starting…'
  } else if (polling && status) {
    buttonLabel = `${status.processed}/${status.total}`
  } else if (isDone && status) {
    buttonLabel = `${status.reclassified} reclassified`
  }

  const inputStyle = {
    width: '100%',
    background: 'var(--surface-2)',
    border: '1px solid var(--border)',
    borderRadius: 12,
    padding: '13px 16px',
    color: 'var(--text)',
    fontFamily: 'DM Sans',
    fontSize: 15,
    outline: 'none',
    resize: 'none' as const,
    boxSizing: 'border-box' as const,
  }

  return (
    <div style={{ padding: '24px 20px', minHeight: '100%' }}>
      <div className="animate-fade-up" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20, opacity: 0 }}>
        <h1 style={{ fontFamily: 'Syne', fontWeight: 800, fontSize: 22, color: 'var(--text)' }}>
          Classify
        </h1>
        <button
          onClick={() => reclassify.mutate()}
          disabled={isRunning}
          className="pressable"
          style={{
            background: isRunning ? 'var(--surface)' : 'var(--accent)',
            border: '1px solid var(--border)',
            borderRadius: 10,
            padding: '0 14px',
            height: 34,
            fontFamily: 'Syne',
            fontWeight: 700,
            fontSize: 12,
            color: isRunning ? 'var(--text-2)' : '#000',
            cursor: isRunning ? 'default' : 'pointer',
            transition: 'all 0.15s',
          }}
        >
          {buttonLabel}
        </button>
      </div>

      <p className="animate-fade-up" style={{ fontSize: 13, color: 'var(--text-3)', marginBottom: 20, lineHeight: 1.5, opacity: 0 }}>
        Describe what belongs in each bucket. The AI uses these descriptions and example transactions to automatically classify new spending.
      </p>

      <div className="animate-fade-up stagger-1" style={{ opacity: 0 }}>
        {isLoading ? (
          Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="shimmer" style={{ height: 80, borderRadius: 14, marginBottom: 10 }} />
          ))
        ) : nonGeneralBuckets.length === 0 ? (
          <div style={{ padding: '60px 0', textAlign: 'center', color: 'var(--text-2)' }}>
            <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 10 }}>
              <Sparkles size={32} color="var(--text-2)" strokeWidth={1.75} />
            </div>
            <p style={{ fontFamily: 'DM Sans', fontSize: 15, marginBottom: 6 }}>No buckets yet</p>
            <p style={{ fontFamily: 'DM Sans', fontSize: 13, color: 'var(--text-3)' }}>
              Create buckets on the home page to start classifying.
            </p>
          </div>
        ) : (
          nonGeneralBuckets.map((b) => (
            <div
              key={b.bucket_id}
              style={{
                background: 'var(--surface)',
                border: '1px solid var(--border)',
                borderRadius: 14,
                padding: '14px 16px',
                marginBottom: 10,
              }}
            >
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: editingId === b.bucket_id ? 10 : 0 }}>
                <p style={{
                  fontFamily: 'Syne',
                  fontWeight: 700,
                  fontSize: 13,
                  letterSpacing: '0.04em',
                  color: 'var(--text)',
                }}>
                  {b.name}
                </p>
                {editingId === b.bucket_id ? (
                  <div style={{ display: 'flex', gap: 6 }}>
                    <button
                      onClick={saveEdit}
                      disabled={updateDesc.isPending}
                      style={{
                        background: 'rgba(74,222,128,0.1)',
                        border: 'none',
                        borderRadius: 8,
                        width: 30,
                        height: 30,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        cursor: 'pointer',
                      }}
                    >
                      <Check size={13} strokeWidth={2} color="var(--green)" />
                    </button>
                    <button
                      onClick={() => setEditingId(null)}
                      style={{
                        background: 'var(--surface-2)',
                        border: 'none',
                        borderRadius: 8,
                        width: 30,
                        height: 30,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        cursor: 'pointer',
                      }}
                    >
                      <X size={13} strokeWidth={1.75} color="var(--text-2)" />
                    </button>
                  </div>
                ) : (
                  <button
                    onClick={() => startEdit(b)}
                    style={{
                      background: 'var(--surface-2)',
                      border: 'none',
                      borderRadius: 8,
                      width: 30,
                      height: 30,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      cursor: 'pointer',
                    }}
                  >
                    <Pencil size={13} strokeWidth={1.75} color="var(--text)" />
                  </button>
                )}
              </div>
              {editingId === b.bucket_id ? (
                <textarea
                  autoFocus
                  value={editValue}
                  onChange={(e) => setEditValue(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); saveEdit() }
                    if (e.key === 'Escape') setEditingId(null)
                  }}
                  placeholder="e.g. My daily coffee at Charlie Bit Me"
                  rows={2}
                  style={inputStyle}
                />
              ) : (
                <p style={{ fontSize: 13, color: b.description ? 'var(--text-2)' : 'var(--text-3)', marginTop: 4, lineHeight: 1.4 }}>
                  {b.description || 'No description — click edit to add one'}
                </p>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  )
}
