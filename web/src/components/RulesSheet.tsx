import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Pencil, Trash2, X } from 'lucide-react'
import { api, type Rule } from '../lib/api'
import { useDraggableSheet } from '../hooks/useDraggableSheet'
import { useMediaQuery } from '../hooks/useMediaQuery'

interface Props {
  bucketId: string
  onClose: () => void
}

type FormState = {
  name: string
  priority: string
  descriptionContains: string
  minAmountAud: string
  maxAmountAud: string
  transactionType: string
  categoryId: string
}

const emptyForm = (): FormState => ({
  name: '',
  priority: '0',
  descriptionContains: '',
  minAmountAud: '',
  maxAmountAud: '',
  transactionType: '',
  categoryId: '',
})

function ruleToForm(r: Rule): FormState {
  return {
    name: r.name,
    priority: String(r.priority),
    descriptionContains: r.description_contains ?? '',
    minAmountAud: r.min_amount_cents != null ? String(r.min_amount_cents / 100) : '',
    maxAmountAud: r.max_amount_cents != null ? String(r.max_amount_cents / 100) : '',
    transactionType: r.transaction_type ?? '',
    categoryId: r.category_id ?? '',
  }
}

export default function RulesSheet({ bucketId, onClose }: Props) {
  const qc = useQueryClient()
  const [editingRule, setEditingRule] = useState<Rule | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState<FormState>(emptyForm())

  const { data: rules = [] } = useQuery({
    queryKey: ['bucket-rules', bucketId],
    queryFn: () => api.getBucketRules(bucketId),
  })

  const invalidate = () => {
    qc.invalidateQueries({ queryKey: ['bucket-rules', bucketId] })
  }

  const createMutation = useMutation({
    mutationFn: () => api.createRule(bucketId, formToPayload(form)),
    onSuccess: () => {
      invalidate()
      setShowForm(false)
      setForm(emptyForm())
    },
  })

  const updateMutation = useMutation({
    mutationFn: () => api.updateRule(editingRule!.rule_id, formToPayload(form)),
    onSuccess: () => {
      invalidate()
      setEditingRule(null)
      setShowForm(false)
      setForm(emptyForm())
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (ruleId: string) => api.deleteRule(ruleId),
    onSuccess: invalidate,
  })

  const applyMutation = useMutation({
    mutationFn: () => api.applyRules(),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['bucket-transactions', bucketId] })
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
    appearance: 'none' as const,
    boxSizing: 'border-box' as const,
  }

  const labelStyle = {
    fontFamily: 'Syne',
    fontWeight: 700,
    fontSize: 11,
    letterSpacing: '0.08em',
    color: 'var(--text-2)',
    marginBottom: 6,
    display: 'block' as const,
  }

  function openCreate() {
    setEditingRule(null)
    setForm(emptyForm())
    setShowForm(true)
  }

  function openEdit(r: Rule) {
    setEditingRule(r)
    setForm(ruleToForm(r))
    setShowForm(true)
  }

  function cancelForm() {
    setEditingRule(null)
    setShowForm(false)
    setForm(emptyForm())
  }

  function handleSubmit() {
    if (editingRule) {
      updateMutation.mutate()
    } else {
      createMutation.mutate()
    }
  }

  const isPending = createMutation.isPending || updateMutation.isPending
  const canSubmit = form.name.trim() !== ''

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
            Rules
          </p>
          <button
            onClick={onClose}
            style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--text-2)', padding: 4 }}
          >
            <X size={18} strokeWidth={1.75} />
          </button>
        </div>

        {/* Existing rules list */}
        {rules.length > 0 && (
          <div style={{ marginBottom: 16 }}>
            {rules.map((r) => (
              <div
                key={r.rule_id}
                style={{
                  display: 'flex',
                  alignItems: 'flex-start',
                  gap: 10,
                  padding: '10px 0',
                  borderBottom: '1px solid var(--border)',
                }}
              >
                <div style={{ flex: 1, minWidth: 0 }}>
                  <p style={{ fontSize: 14, color: 'var(--text)', fontWeight: 500 }}>{r.name}</p>
                  <p style={{ fontSize: 12, color: 'var(--text-2)', marginTop: 2 }}>
                    Priority {r.priority}
                    {r.description_contains && ` · contains "${r.description_contains}"`}
                    {r.transaction_type && ` · type "${r.transaction_type}"`}
                    {r.category_id && ` · category "${r.category_id}"`}
                    {r.min_amount_cents != null && ` · min $${(r.min_amount_cents / 100).toFixed(2)}`}
                    {r.max_amount_cents != null && ` · max $${(r.max_amount_cents / 100).toFixed(2)}`}
                  </p>
                </div>
                <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
                  <button
                    onClick={() => openEdit(r)}
                    style={{
                      background: 'var(--surface-2)',
                      border: 'none',
                      borderRadius: 8,
                      padding: '6px 8px',
                      color: 'var(--text)',
                      cursor: 'pointer',
                    }}
                  >
                    <Pencil size={13} strokeWidth={1.75} />
                  </button>
                  <button
                    onClick={() => deleteMutation.mutate(r.rule_id)}
                    disabled={deleteMutation.isPending}
                    style={{
                      background: 'rgba(248,113,113,0.1)',
                      border: 'none',
                      borderRadius: 8,
                      padding: '6px 8px',
                      color: 'var(--red)',
                      cursor: 'pointer',
                    }}
                  >
                    <Trash2 size={13} strokeWidth={1.75} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Form */}
        {showForm ? (
          <div>
            <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 14, color: 'var(--text)', marginBottom: 14 }}>
              {editingRule ? 'Edit Rule' : 'New Rule'}
            </p>

            <label style={labelStyle}>NAME</label>
            <input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="e.g. Coffee shops"
              style={{ ...inputStyle, marginBottom: 12 }}
            />

            <label style={labelStyle}>DESCRIPTION CONTAINS (OPTIONAL)</label>
            <input
              value={form.descriptionContains}
              onChange={(e) => setForm({ ...form, descriptionContains: e.target.value })}
              placeholder="e.g. coffee"
              style={{ ...inputStyle, marginBottom: 12 }}
            />

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, marginBottom: 12 }}>
              <div>
                <label style={labelStyle}>MIN AMOUNT AUD (OPTIONAL)</label>
                <div style={{ position: 'relative' }}>
                  <span style={{ position: 'absolute', left: 16, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-2)', fontFamily: 'DM Sans', fontSize: 16 }}>$</span>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    value={form.minAmountAud}
                    onChange={(e) => setForm({ ...form, minAmountAud: e.target.value })}
                    placeholder="0.00"
                    style={{ ...inputStyle, paddingLeft: 30 }}
                  />
                </div>
              </div>
              <div>
                <label style={labelStyle}>MAX AMOUNT AUD (OPTIONAL)</label>
                <div style={{ position: 'relative' }}>
                  <span style={{ position: 'absolute', left: 16, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-2)', fontFamily: 'DM Sans', fontSize: 16 }}>$</span>
                  <input
                    type="number"
                    min="0"
                    step="0.01"
                    value={form.maxAmountAud}
                    onChange={(e) => setForm({ ...form, maxAmountAud: e.target.value })}
                    placeholder="0.00"
                    style={{ ...inputStyle, paddingLeft: 30 }}
                  />
                </div>
              </div>
            </div>

            <label style={labelStyle}>TRANSACTION TYPE (OPTIONAL)</label>
            <input
              value={form.transactionType}
              onChange={(e) => setForm({ ...form, transactionType: e.target.value })}
              placeholder="e.g. Transfer, Round Up"
              style={{ ...inputStyle, marginBottom: 12 }}
            />

            <label style={labelStyle}>CATEGORY (OPTIONAL)</label>
            <input
              value={form.categoryId}
              onChange={(e) => setForm({ ...form, categoryId: e.target.value })}
              placeholder="e.g. restaurants-and-cafes"
              style={{ ...inputStyle, marginBottom: 12 }}
            />

            <label style={labelStyle}>PRIORITY (LOWER = FIRST)</label>
            <input
              type="number"
              value={form.priority}
              onChange={(e) => setForm({ ...form, priority: e.target.value })}
              placeholder="0"
              style={{ ...inputStyle, marginBottom: 16 }}
            />

            <div style={{ display: 'flex', gap: 8 }}>
              <button
                onClick={cancelForm}
                style={{
                  flex: 1,
                  padding: '13px',
                  background: 'transparent',
                  border: '1px solid var(--border)',
                  borderRadius: 12,
                  color: 'var(--text-2)',
                  fontFamily: 'DM Sans',
                  fontSize: 15,
                  cursor: 'pointer',
                }}
              >
                Cancel
              </button>
              <button
                onClick={handleSubmit}
                disabled={!canSubmit || isPending}
                className="pressable"
                style={{
                  flex: 2,
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
                {isPending ? 'Saving…' : editingRule ? 'Update Rule' : 'Add Rule'}
              </button>
            </div>
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            <button
              onClick={openCreate}
              className="pressable"
              style={{
                width: '100%',
                padding: '13px',
                background: 'var(--surface-2)',
                border: '1px dashed var(--border)',
                borderRadius: 12,
                color: 'var(--text-2)',
                fontFamily: 'DM Sans',
                fontSize: 15,
                cursor: 'pointer',
              }}
            >
              + Add Rule
            </button>
            <button
              onClick={() => applyMutation.mutate()}
              disabled={applyMutation.isPending}
              className="pressable"
              style={{
                width: '100%',
                padding: '13px',
                background: applyMutation.isPending ? 'var(--surface-2)' : 'var(--accent)',
                border: 'none',
                borderRadius: 12,
                color: applyMutation.isPending ? 'var(--text-2)' : '#000',
                fontFamily: 'Syne',
                fontWeight: 700,
                fontSize: 15,
                cursor: applyMutation.isPending ? 'default' : 'pointer',
                transition: 'all 0.15s',
              }}
            >
              {applyMutation.isPending ? 'Applying…' : applyMutation.isSuccess ? `Applied!` : 'Apply Rules Now'}
            </button>
          </div>
        )}
      </div>
    </>
  )
}

function formToPayload(form: FormState) {
  return {
    name: form.name.trim(),
    priority: parseInt(form.priority) || 0,
    description_contains: form.descriptionContains.trim() || null,
    min_amount_aud: form.minAmountAud !== '' ? parseFloat(form.minAmountAud) : null,
    max_amount_aud: form.maxAmountAud !== '' ? parseFloat(form.maxAmountAud) : null,
    transaction_type: form.transactionType.trim() || null,
    category_id: form.categoryId.trim() || null,
  }
}
