import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Filter, Pencil, Plus, Trash2 } from 'lucide-react'
import { api, type Rule } from '../lib/api'

type FormState = {
  bucketId: string
  name: string
  priority: string
  descriptionContains: string
  minAmountAud: string
  maxAmountAud: string
  transactionType: string
  categoryId: string
}

const emptyForm = (bucketId = ''): FormState => ({
  bucketId,
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
    bucketId: r.bucket_id,
    name: r.name,
    priority: String(r.priority),
    descriptionContains: r.description_contains ?? '',
    minAmountAud: r.min_amount_cents != null ? String(r.min_amount_cents / 100) : '',
    maxAmountAud: r.max_amount_cents != null ? String(r.max_amount_cents / 100) : '',
    transactionType: r.transaction_type ?? '',
    categoryId: r.category_id ?? '',
  }
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

export default function Rules() {
  const qc = useQueryClient()
  const [editingRule, setEditingRule] = useState<Rule | null>(null)
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState<FormState>(emptyForm())

  const { data: rules = [], isLoading } = useQuery({
    queryKey: ['rules'],
    queryFn: api.getRules,
  })

  const { data: buckets = [] } = useQuery({
    queryKey: ['buckets'],
    queryFn: api.getBuckets,
  })

  const nonGeneralBuckets = buckets.filter((b) => !b.is_general)

  const invalidate = () => qc.invalidateQueries({ queryKey: ['rules'] })

  const createMutation = useMutation({
    mutationFn: () => api.createRule(form.bucketId, formToPayload(form)),
    onSuccess: () => { invalidate(); cancelForm() },
  })

  const updateMutation = useMutation({
    mutationFn: () => api.updateRule(editingRule!.rule_id, formToPayload(form)),
    onSuccess: () => { invalidate(); cancelForm() },
  })

  const deleteMutation = useMutation({
    mutationFn: (ruleId: string) => api.deleteRule(ruleId),
    onSuccess: invalidate,
  })

  const applyMutation = useMutation({
    mutationFn: () => api.applyRules(),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['buckets'] })
      qc.invalidateQueries({ queryKey: ['bucket-transactions'] })
    },
  })

  function openCreate() {
    setEditingRule(null)
    setForm(emptyForm(nonGeneralBuckets[0]?.bucket_id ?? ''))
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
    if (editingRule) updateMutation.mutate()
    else createMutation.mutate()
  }

  const isPending = createMutation.isPending || updateMutation.isPending
  const canSubmit = form.name.trim() !== '' && form.bucketId !== ''

  // Group rules by bucket
  const byBucket = rules.reduce<Record<string, Rule[]>>((acc, r) => {
    if (!acc[r.bucket_id]) acc[r.bucket_id] = []
    acc[r.bucket_id].push(r)
    return acc
  }, {})

  const inputStyle = {
    width: '100%',
    background: 'var(--surface)',
    border: '1px solid var(--border)',
    borderRadius: 12,
    padding: '13px 16px',
    color: 'var(--text)',
    fontFamily: 'DM Sans',
    fontSize: 15,
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

  return (
    <div style={{ padding: '24px 20px', minHeight: '100%' }}>
      <div className="animate-fade-up" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 20, opacity: 0 }}>
        <h1 style={{ fontFamily: 'Syne', fontWeight: 800, fontSize: 22, color: 'var(--text)' }}>
          Rules
        </h1>
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            onClick={() => applyMutation.mutate()}
            disabled={applyMutation.isPending || rules.length === 0}
            className="pressable"
            style={{
              background: applyMutation.isPending || rules.length === 0 ? 'var(--surface)' : 'var(--accent)',
              border: '1px solid var(--border)',
              borderRadius: 10,
              padding: '0 14px',
              height: 34,
              fontFamily: 'Syne',
              fontWeight: 700,
              fontSize: 12,
              color: applyMutation.isPending || rules.length === 0 ? 'var(--text-2)' : '#000',
              cursor: applyMutation.isPending || rules.length === 0 ? 'default' : 'pointer',
              transition: 'all 0.15s',
            }}
          >
            {applyMutation.isPending ? 'Applying…' : applyMutation.isSuccess ? 'Applied!' : 'Apply All'}
          </button>
          <button
            onClick={openCreate}
            disabled={nonGeneralBuckets.length === 0}
            className="pressable"
            style={{
              background: 'var(--surface)',
              border: '1px solid var(--border)',
              borderRadius: 10,
              width: 34,
              height: 34,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: nonGeneralBuckets.length === 0 ? 'default' : 'pointer',
            }}
          >
            <Plus size={20} color="var(--accent)" strokeWidth={1.75} />
          </button>
        </div>
      </div>

      {/* Rule form */}
      {showForm && (
        <div
          className="animate-fade-up"
          style={{
            background: 'var(--surface)',
            border: '1px solid var(--border)',
            borderRadius: 16,
            padding: '20px',
            marginBottom: 20,
            opacity: 0,
          }}
        >
          <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 15, color: 'var(--text)', marginBottom: 16 }}>
            {editingRule ? 'Edit Rule' : 'New Rule'}
          </p>

          {!editingRule && (
            <>
              <label style={labelStyle}>BUCKET</label>
              <select
                value={form.bucketId}
                onChange={(e) => setForm({ ...form, bucketId: e.target.value })}
                style={{ ...inputStyle, marginBottom: 12 }}
              >
                {nonGeneralBuckets.map((b) => (
                  <option key={b.bucket_id} value={b.bucket_id}>{b.name}</option>
                ))}
              </select>
            </>
          )}

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
                <span style={{ position: 'absolute', left: 16, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-2)', fontFamily: 'DM Sans', fontSize: 15 }}>$</span>
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
                <span style={{ position: 'absolute', left: 16, top: '50%', transform: 'translateY(-50%)', color: 'var(--text-2)', fontFamily: 'DM Sans', fontSize: 15 }}>$</span>
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

          <label style={labelStyle}>PRIORITY (LOWER = EVALUATED FIRST)</label>
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
                padding: '12px',
                background: 'transparent',
                border: '1px solid var(--border)',
                borderRadius: 10,
                color: 'var(--text-2)',
                fontFamily: 'DM Sans',
                fontSize: 14,
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
                padding: '12px',
                background: canSubmit ? 'var(--accent)' : 'var(--surface-2)',
                border: 'none',
                borderRadius: 10,
                color: canSubmit ? '#000' : 'var(--text-2)',
                fontFamily: 'Syne',
                fontWeight: 700,
                fontSize: 14,
                cursor: canSubmit ? 'pointer' : 'default',
                transition: 'all 0.15s',
              }}
            >
              {isPending ? 'Saving…' : editingRule ? 'Update Rule' : 'Add Rule'}
            </button>
          </div>
        </div>
      )}

      {/* Rules list */}
      <div className="animate-fade-up stagger-1" style={{ opacity: 0 }}>
        {isLoading ? (
          Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="shimmer" style={{ height: 68, borderRadius: 14, marginBottom: 10 }} />
          ))
        ) : rules.length === 0 ? (
          <div style={{ padding: '60px 0', textAlign: 'center', color: 'var(--text-2)' }}>
            <div style={{ display: 'flex', justifyContent: 'center', marginBottom: 10 }}>
              <Filter size={32} color="var(--text-2)" strokeWidth={1.75} />
            </div>
            <p style={{ fontFamily: 'DM Sans', fontSize: 15, marginBottom: 6 }}>No rules yet</p>
            <p style={{ fontFamily: 'DM Sans', fontSize: 13, color: 'var(--text-3)' }}>
              Rules auto-assign incoming transactions to buckets.
            </p>
          </div>
        ) : (
          Object.entries(byBucket).map(([bucketId, bucketRules]) => (
            <div key={bucketId} style={{ marginBottom: 20 }}>
              <p
                style={{
                  fontFamily: 'Syne',
                  fontWeight: 700,
                  fontSize: 11,
                  letterSpacing: '0.08em',
                  color: 'var(--text-2)',
                  marginBottom: 8,
                }}
              >
                {bucketRules[0].bucket_name.toUpperCase()}
              </p>
              {bucketRules.map((r) => (
                <div
                  key={r.rule_id}
                  style={{
                    background: 'var(--surface)',
                    border: '1px solid var(--border)',
                    borderRadius: 14,
                    padding: '14px 16px',
                    marginBottom: 8,
                    display: 'flex',
                    alignItems: 'flex-start',
                    gap: 12,
                  }}
                >
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <p style={{ fontSize: 14, color: 'var(--text)', fontWeight: 500, marginBottom: 4 }}>
                      {r.name}
                    </p>
                    <p style={{ fontSize: 12, color: 'var(--text-2)', lineHeight: 1.5 }}>
                      {[
                        r.description_contains && `contains "${r.description_contains}"`,
                        r.transaction_type && `type "${r.transaction_type}"`,
                        r.category_id && `category "${r.category_id}"`,
                        r.min_amount_cents != null && `min $${(r.min_amount_cents / 100).toFixed(2)}`,
                        r.max_amount_cents != null && `max $${(r.max_amount_cents / 100).toFixed(2)}`,
                      ].filter(Boolean).join(' · ') || 'Matches all transactions'}
                    </p>
                    <p style={{ fontSize: 11, color: 'var(--text-3)', marginTop: 3 }}>Priority {r.priority}</p>
                  </div>
                  <div style={{ display: 'flex', gap: 6, flexShrink: 0 }}>
                    <button
                      onClick={() => openEdit(r)}
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
                    <button
                      onClick={() => deleteMutation.mutate(r.rule_id)}
                      disabled={deleteMutation.isPending}
                      style={{
                        background: 'rgba(248,113,113,0.1)',
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
                      <Trash2 size={13} strokeWidth={1.75} color="var(--red)" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          ))
        )}
      </div>
    </div>
  )
}
