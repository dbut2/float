import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { X } from 'lucide-react'
import { api } from '../lib/api'
import { useDraggableSheet } from '../hooks/useDraggableSheet'
import { useMediaQuery } from '../hooks/useMediaQuery'

interface Props {
  bucketId: string
  onClose: () => void
}

export default function DescriptionSheet({ bucketId, onClose }: Props) {
  const qc = useQueryClient()
  const { data: buckets = [] } = useQuery({
    queryKey: ['buckets'],
    queryFn: api.getBuckets,
  })
  const bucket = buckets.find((b) => b.bucket_id === bucketId)

  const [description, setDescription] = useState(bucket?.description ?? '')

  const updateMutation = useMutation({
    mutationFn: () => api.updateBucketDescription(bucketId, description),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buckets'] })
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
    resize: 'none' as const,
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
          width: 520, zIndex: 101,
          background: 'var(--surface)', borderRadius: 20,
          padding: '24px', animation: 'fadeIn 0.2s ease forwards',
          maxHeight: '85vh', overflowY: 'auto',
        } : {
          position: 'fixed', bottom: 0, left: 0, right: 0, zIndex: 101,
          background: 'var(--surface)',
          borderRadius: '20px 20px 0 0',
          padding: '16px 20px',
          paddingBottom: 24,
          maxHeight: '90vh',
          overflowY: 'auto',
          ...sheetStyle,
        }}
      >
        {!isDesktop && (
          <div ref={handleRef} style={{ touchAction: 'none', display: 'flex', justifyContent: 'center', marginBottom: 20 }}>
            <div style={{ width: 36, height: 4, background: 'var(--border)', borderRadius: 2 }} />
          </div>
        )}

        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 16 }}>
          <p style={{ fontFamily: 'Syne', fontWeight: 800, fontSize: 20, color: 'var(--text)' }}>
            Description
          </p>
          <button
            onClick={onClose}
            style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-2)', padding: 4 }}
          >
            <X size={18} strokeWidth={1.75} />
          </button>
        </div>

        <p style={{ fontSize: 13, color: 'var(--text-3)', marginBottom: 14, lineHeight: 1.5 }}>
          Describe what belongs in this bucket. The AI uses this to classify incoming transactions.
        </p>

        <textarea
          autoFocus
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); updateMutation.mutate() }
            if (e.key === 'Escape') onClose()
          }}
          placeholder="e.g. My daily coffee at Charlie Bit Me"
          rows={3}
          style={{ ...inputStyle, marginBottom: 16 }}
        />

        <button
          onClick={() => updateMutation.mutate()}
          disabled={updateMutation.isPending}
          className="pressable"
          style={{
            width: '100%',
            padding: '13px',
            background: 'var(--accent)',
            border: 'none',
            borderRadius: 12,
            color: '#000',
            fontFamily: 'Syne',
            fontWeight: 700,
            fontSize: 15,
            cursor: 'pointer',
            transition: 'all 0.15s',
          }}
        >
          {updateMutation.isPending ? 'Saving…' : 'Save'}
        </button>
      </div>
    </>
  )
}
