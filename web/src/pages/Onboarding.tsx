import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  ArrowLeft,
  ArrowRight,
  Check,
  ExternalLink,
  Plus,
  Sparkles,
  Wallet,
  Layers,
  Droplets,
  X,
} from 'lucide-react'
import { useMediaQuery } from '../hooks/useMediaQuery'
import { api, type Bucket } from '../lib/api'

type StepId = 0 | 1 | 2 | 3 | 4

interface SuggestedBucket {
  key: string
  name: string
  description: string
}

const SUGGESTED_BUCKETS: SuggestedBucket[] = [
  { key: 'groceries', name: 'Groceries', description: 'Weekly food shop' },
  { key: 'rent', name: 'Rent', description: 'Monthly rent or mortgage' },
  { key: 'bills', name: 'Bills', description: 'Power, internet, phone' },
  { key: 'savings', name: 'Savings', description: 'Long-term cushion' },
  { key: 'coffee', name: 'Coffee', description: 'Daily treat fund' },
  { key: 'travel', name: 'Travel', description: 'Holidays & weekends away' },
]

const PERIODS: Array<{ value: string; label: string }> = [
  { value: 'daily', label: 'Daily' },
  { value: 'weekly', label: 'Weekly' },
  { value: 'fortnightly', label: 'Fortnightly' },
  { value: 'monthly', label: 'Monthly' },
]

export default function Onboarding() {
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [params, setParams] = useSearchParams()
  const isDesktop = useMediaQuery('(min-width: 768px)')

  const initialStep = (() => {
    const raw = params.get('step')
    if (raw == null) return 0 as StepId
    const n = parseInt(raw, 10)
    if (isNaN(n)) return 0 as StepId
    return Math.min(4, Math.max(0, n)) as StepId
  })()
  const [step, setStep] = useState<StepId>(initialStep)

  const { data: me } = useQuery({ queryKey: ['user'], queryFn: api.getUser })
  const { data: existingBuckets = [] } = useQuery({
    queryKey: ['buckets'],
    queryFn: api.getBuckets,
    enabled: step >= 3,
  })

  const [token, setToken] = useState('')
  const [tokenSaved, setTokenSaved] = useState(false)
  const [tokenError, setTokenError] = useState<string | null>(null)

  useEffect(() => {
    if (me?.has_token) setTokenSaved(true)
  }, [me?.has_token])

  const [selectedBucketKeys, setSelectedBucketKeys] = useState<Set<string>>(
    new Set(['groceries', 'bills', 'savings'])
  )
  const [customBuckets, setCustomBuckets] = useState<string[]>([])
  const [showCustomInput, setShowCustomInput] = useState(false)
  const [customDraft, setCustomDraft] = useState('')
  const [createdBuckets, setCreatedBuckets] = useState<Bucket[]>([])
  const [bucketsError, setBucketsError] = useState<string | null>(null)

  const [trickleBucketId, setTrickleBucketId] = useState<string>('')
  const [trickleAmount, setTrickleAmount] = useState<string>('50')
  const [tricklePeriod, setTricklePeriod] = useState<string>('weekly')
  const [trickleError, setTrickleError] = useState<string | null>(null)

  useEffect(() => {
    setParams({ step: String(step) }, { replace: true })
  }, [step, setParams])

  const totalSteps = 4
  const progressPct = step === 0 ? 0 : (step / totalSteps) * 100

  const setTokenMut = useMutation({
    mutationFn: (t: string) => api.setToken(t),
    onSuccess: () => {
      setTokenSaved(true)
      setTokenError(null)
      qc.invalidateQueries({ queryKey: ['user'] })
    },
    onError: (err: Error) => setTokenError(err.message || 'Failed to save token'),
  })

  const createBucketsMut = useMutation({
    mutationFn: async (names: string[]) => {
      const out: Bucket[] = []
      for (const name of names) {
        out.push(await api.createBucket(name))
      }
      return out
    },
    onSuccess: (buckets) => {
      setCreatedBuckets(buckets)
      setBucketsError(null)
      if (buckets.length > 0 && !trickleBucketId) {
        setTrickleBucketId(buckets[0].bucket_id)
      }
      qc.invalidateQueries({ queryKey: ['buckets'] })
      setStep(3)
    },
    onError: (err: Error) => setBucketsError(err.message || 'Failed to create buckets'),
  })

  const upsertTrickleMut = useMutation({
    mutationFn: () => {
      const amountCents = Math.round(parseFloat(trickleAmount || '0') * 100)
      const melbourneToday = new Date().toLocaleDateString('en-CA', { timeZone: 'Australia/Melbourne' })
      const [y, m, d] = melbourneToday.split('-').map(Number)
      const startDate = new Date(Date.UTC(y, m - 1, d + 1)).toISOString().slice(0, 10)
      return api.upsertTrickle(trickleBucketId, {
        amount_cents: amountCents,
        description: '',
        period: tricklePeriod,
        start_date: startDate,
        end_date: null,
      })
    },
    onSuccess: () => {
      setTrickleError(null)
      qc.invalidateQueries({ queryKey: ['trickles'] })
      qc.invalidateQueries({ queryKey: ['buckets'] })
      setStep(4)
    },
    onError: (err: Error) => setTrickleError(err.message || 'Failed to schedule trickle'),
  })

  const allBucketNames = useMemo(() => {
    const fromSuggested = SUGGESTED_BUCKETS.filter((b) => selectedBucketKeys.has(b.key)).map((b) => b.name)
    return [...fromSuggested, ...customBuckets]
  }, [selectedBucketKeys, customBuckets])

  const bucketCount = allBucketNames.length

  const trickleBuckets = useMemo(() => {
    if (createdBuckets.length > 0) return createdBuckets
    return existingBuckets.filter((b) => !b.is_general)
  }, [createdBuckets, existingBuckets])

  useEffect(() => {
    if (!trickleBucketId && trickleBuckets.length > 0) {
      setTrickleBucketId(trickleBuckets[0].bucket_id)
    }
  }, [trickleBuckets, trickleBucketId])
  const hasTrickleConfig = parseFloat(trickleAmount || '0') > 0 && !!trickleBucketId

  const goNext = () => {
    if (step === 2 && createdBuckets.length === 0 && bucketCount > 0) {
      createBucketsMut.mutate(allBucketNames)
      return
    }
    if (step === 3 && hasTrickleConfig) {
      upsertTrickleMut.mutate()
      return
    }
    if (step < totalSteps) setStep((step + 1) as StepId)
    else navigate('/')
  }

  const goBack = () => {
    if (step > 0) setStep((step - 1) as StepId)
  }

  const ctaPending =
    (step === 2 && createBucketsMut.isPending) || (step === 3 && upsertTrickleMut.isPending)

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'var(--bg)',
        display: 'flex',
        flexDirection: 'column',
        zIndex: 200,
      }}
    >
      {step > 0 && (
        <div
          style={{
            flexShrink: 0,
            padding: isDesktop ? '24px 40px 0' : '16px 20px 0',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              marginBottom: 18,
            }}
          >
            <button
              onClick={goBack}
              className="pressable"
              style={{
                background: 'transparent',
                border: 'none',
                padding: 8,
                marginLeft: -8,
                cursor: 'pointer',
                color: 'var(--text-2)',
                display: 'flex',
                alignItems: 'center',
              }}
              aria-label="Back"
            >
              <ArrowLeft size={20} strokeWidth={1.75} />
            </button>
            <p
              style={{
                fontFamily: 'Syne',
                fontWeight: 700,
                fontSize: 11,
                letterSpacing: '0.12em',
                color: 'var(--text-2)',
                textTransform: 'uppercase',
              }}
            >
              Step {step} of {totalSteps}
            </p>
            {step < totalSteps ? (
              <button
                onClick={goNext}
                className="pressable"
                style={{
                  background: 'transparent',
                  border: 'none',
                  padding: 8,
                  marginRight: -8,
                  cursor: 'pointer',
                  color: 'var(--text-2)',
                  fontFamily: 'Syne',
                  fontWeight: 700,
                  fontSize: 12,
                  letterSpacing: '0.06em',
                }}
              >
                SKIP
              </button>
            ) : (
              <span style={{ width: 36 }} />
            )}
          </div>
          <div
            style={{
              width: '100%',
              height: 3,
              borderRadius: 2,
              background: 'var(--surface-2)',
              overflow: 'hidden',
            }}
          >
            <div
              style={{
                width: `${progressPct}%`,
                height: '100%',
                background: 'var(--accent)',
                transition: 'width 0.35s cubic-bezier(0.2, 0.8, 0.2, 1)',
              }}
            />
          </div>
        </div>
      )}

      <div
        style={{
          flex: 1,
          overflowY: 'auto',
          overflowX: 'hidden',
        }}
      >
        <div
          style={{
            maxWidth: 480,
            margin: '0 auto',
            padding: isDesktop ? '40px 40px 32px' : '32px 24px 24px',
          }}
        >
          {step === 0 && <StepWelcome />}
          {step === 1 && (
            <StepToken
              token={token}
              setToken={setToken}
              tokenSaved={tokenSaved}
              clearSaved={() => setTokenSaved(false)}
              onVerify={() => setTokenMut.mutate(token.trim())}
              verifying={setTokenMut.isPending}
              error={tokenError}
            />
          )}
          {step === 2 && (
            <StepBuckets
              selected={selectedBucketKeys}
              setSelected={setSelectedBucketKeys}
              customBuckets={customBuckets}
              setCustomBuckets={setCustomBuckets}
              showCustomInput={showCustomInput}
              setShowCustomInput={setShowCustomInput}
              customDraft={customDraft}
              setCustomDraft={setCustomDraft}
              error={bucketsError}
            />
          )}
          {step === 3 && (
            <StepTrickle
              buckets={trickleBuckets}
              trickleBucketId={trickleBucketId}
              setTrickleBucketId={setTrickleBucketId}
              trickleAmount={trickleAmount}
              setTrickleAmount={setTrickleAmount}
              tricklePeriod={tricklePeriod}
              setTricklePeriod={setTricklePeriod}
              error={trickleError}
            />
          )}
          {step === 4 && (
            <StepDone
              tokenSaved={tokenSaved}
              bucketCount={createdBuckets.length || trickleBuckets.length}
              hasTrickle={hasTrickleConfig && !upsertTrickleMut.isError}
            />
          )}
        </div>
      </div>

      <div
        style={{
          flexShrink: 0,
          padding: isDesktop ? '20px 40px 32px' : '16px 24px 28px',
          background: 'var(--bg)',
          borderTop: '1px solid transparent',
        }}
      >
        <div style={{ maxWidth: 480, margin: '0 auto' }}>
          <PrimaryCTA
            step={step}
            onNext={goNext}
            tokenSaved={tokenSaved}
            bucketCount={bucketCount}
            pending={ctaPending}
          />
        </div>
      </div>
    </div>
  )
}

function PrimaryCTA({
  step,
  onNext,
  tokenSaved,
  bucketCount,
  pending,
}: {
  step: StepId
  onNext: () => void
  tokenSaved: boolean
  bucketCount: number
  pending: boolean
}) {
  let label = 'Continue'
  let enabled = true

  if (step === 0) label = 'Get started'
  if (step === 1) {
    label = tokenSaved ? 'Continue' : 'Save token first'
    enabled = tokenSaved
  }
  if (step === 2) {
    label = bucketCount > 0 ? `Continue with ${bucketCount} bucket${bucketCount === 1 ? '' : 's'}` : 'Pick at least one bucket'
    enabled = bucketCount > 0
  }
  if (step === 3) label = 'Continue'
  if (step === 4) label = 'Open Float'

  if (pending) {
    label = step === 2 ? 'Creating buckets…' : step === 3 ? 'Scheduling…' : 'Working…'
    enabled = false
  }

  return (
    <button
      onClick={onNext}
      disabled={!enabled}
      className="pressable"
      style={{
        width: '100%',
        padding: '17px',
        background: enabled ? 'var(--accent)' : 'var(--surface-2)',
        border: 'none',
        borderRadius: 14,
        color: enabled ? '#000' : 'var(--text-2)',
        fontFamily: 'Syne',
        fontWeight: 800,
        fontSize: 15,
        letterSpacing: '0.02em',
        cursor: enabled ? 'pointer' : 'default',
        transition: 'all 0.15s',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        gap: 8,
      }}
    >
      {label}
      {enabled && step < 4 && <ArrowRight size={17} strokeWidth={2.25} />}
    </button>
  )
}

function StepHeading({ kicker, title, body }: { kicker: string; title: string; body: string }) {
  return (
    <>
      <p
        style={{
          fontFamily: 'Syne',
          fontWeight: 700,
          fontSize: 11,
          letterSpacing: '0.14em',
          color: 'var(--accent)',
          textTransform: 'uppercase',
          marginBottom: 12,
        }}
      >
        {kicker}
      </p>
      <h1
        style={{
          fontFamily: 'Syne',
          fontWeight: 800,
          fontSize: 30,
          lineHeight: 1.1,
          color: 'var(--text)',
          marginBottom: 14,
          letterSpacing: '-0.01em',
        }}
      >
        {title}
      </h1>
      <p
        style={{
          fontFamily: 'DM Sans',
          fontSize: 15,
          lineHeight: 1.55,
          color: 'var(--text-2)',
          marginBottom: 28,
        }}
      >
        {body}
      </p>
    </>
  )
}

function StepWelcome() {
  const features = [
    {
      Icon: Layers,
      title: 'Bucket your money',
      body: 'Carve up your Up balance into envelopes — rent, groceries, the next holiday.',
    },
    {
      Icon: Sparkles,
      title: 'Auto-classify spending',
      body: 'Each transaction lands in the right bucket so you always know what is left.',
    },
    {
      Icon: Droplets,
      title: 'Trickle into goals',
      body: 'Drip a fixed amount into a bucket every week, fortnight, or month.',
    },
  ]
  return (
    <div>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          marginBottom: 36,
          marginTop: 16,
        }}
      >
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: 14,
          }}
        >
          <div
            style={{
              width: 76,
              height: 76,
              borderRadius: 22,
              background: 'var(--accent-dim)',
              border: '1px solid rgba(202, 255, 51, 0.25)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Wallet size={34} color="var(--accent)" strokeWidth={1.75} />
          </div>
          <p
            style={{
              fontFamily: 'Syne',
              fontWeight: 800,
              fontSize: 36,
              color: 'var(--accent)',
              letterSpacing: '-0.03em',
            }}
          >
            FLOAT
          </p>
        </div>
      </div>
      <h1
        style={{
          fontFamily: 'Syne',
          fontWeight: 800,
          fontSize: 28,
          lineHeight: 1.15,
          color: 'var(--text)',
          marginBottom: 12,
          textAlign: 'center',
          letterSpacing: '-0.01em',
        }}
      >
        Envelope budgeting,<br />on top of Up Bank.
      </h1>
      <p
        style={{
          fontFamily: 'DM Sans',
          fontSize: 15,
          lineHeight: 1.55,
          color: 'var(--text-2)',
          marginBottom: 32,
          textAlign: 'center',
        }}
      >
        Float reads your Up transactions and sorts them into buckets you control.
      </p>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
        {features.map(({ Icon, title, body }, i) => (
          <div
            key={i}
            style={{
              display: 'flex',
              gap: 14,
              padding: '16px 18px',
              background: 'var(--surface)',
              border: '1px solid var(--border)',
              borderRadius: 16,
            }}
          >
            <div
              style={{
                width: 36,
                height: 36,
                borderRadius: 10,
                background: 'var(--surface-2)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0,
              }}
            >
              <Icon size={18} color="var(--accent)" strokeWidth={1.75} />
            </div>
            <div style={{ flex: 1, minWidth: 0 }}>
              <p
                style={{
                  fontFamily: 'Syne',
                  fontWeight: 700,
                  fontSize: 14,
                  color: 'var(--text)',
                  marginBottom: 3,
                }}
              >
                {title}
              </p>
              <p style={{ fontFamily: 'DM Sans', fontSize: 13, lineHeight: 1.45, color: 'var(--text-2)' }}>{body}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function StepToken({
  token,
  setToken,
  tokenSaved,
  clearSaved,
  onVerify,
  verifying,
  error,
}: {
  token: string
  setToken: (v: string) => void
  tokenSaved: boolean
  clearSaved: () => void
  onVerify: () => void
  verifying: boolean
  error: string | null
}) {
  const handleVerify = () => {
    if (!token.trim()) return
    onVerify()
  }
  return (
    <div>
      <StepHeading
        kicker="Connect"
        title="Link your Up account"
        body="Float needs a personal access token from Up so it can read your transactions. Float can only read — it can never move money."
      />

      <div
        style={{
          background: 'var(--surface)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          padding: '18px',
          marginBottom: 16,
        }}
      >
        <p
          style={{
            fontFamily: 'Syne',
            fontWeight: 700,
            fontSize: 11,
            letterSpacing: '0.08em',
            color: 'var(--text-2)',
            marginBottom: 10,
            textTransform: 'uppercase',
          }}
        >
          Personal access token
        </p>
        <div style={{ display: 'flex', gap: 8 }}>
          <input
            value={token}
            onChange={(e) => {
              setToken(e.target.value)
              clearSaved()
            }}
            type="password"
            placeholder="up:yeah:…"
            style={{
              flex: 1,
              background: 'var(--surface-2)',
              border: `1px solid ${error ? 'var(--red)' : tokenSaved ? 'var(--green)' : 'var(--border)'}`,
              borderRadius: 10,
              padding: '13px 14px',
              color: 'var(--text)',
              fontFamily: 'JetBrains Mono',
              fontSize: 13,
              outline: 'none',
              minWidth: 0,
            }}
          />
          <button
            onClick={handleVerify}
            disabled={!token.trim() || verifying || tokenSaved}
            className="pressable"
            style={{
              background: tokenSaved ? 'var(--surface-2)' : token.trim() ? 'var(--accent)' : 'var(--surface-2)',
              border: 'none',
              borderRadius: 10,
              padding: '13px 16px',
              color: tokenSaved ? 'var(--green)' : token.trim() ? '#000' : 'var(--text-2)',
              fontFamily: 'Syne',
              fontWeight: 700,
              fontSize: 13,
              cursor: 'pointer',
              whiteSpace: 'nowrap',
              display: 'flex',
              alignItems: 'center',
              gap: 6,
            }}
          >
            {tokenSaved ? (
              <>
                <Check size={14} strokeWidth={2.25} /> Saved
              </>
            ) : verifying ? (
              'Checking…'
            ) : (
              'Verify'
            )}
          </button>
        </div>
        {tokenSaved && !error && (
          <p
            style={{
              fontSize: 12,
              color: 'var(--green)',
              marginTop: 10,
              fontFamily: 'DM Sans',
              display: 'flex',
              alignItems: 'center',
              gap: 5,
            }}
          >
            <Check size={12} strokeWidth={2.25} /> Connected to Up
          </p>
        )}
        {error && (
          <p
            style={{
              fontSize: 12,
              color: 'var(--red)',
              marginTop: 10,
              fontFamily: 'DM Sans',
              display: 'flex',
              alignItems: 'center',
              gap: 5,
            }}
          >
            <X size={12} strokeWidth={2.25} /> {error}
          </p>
        )}
      </div>

      <a
        href="https://api.up.com.au/getting_started"
        target="_blank"
        rel="noreferrer"
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '14px 18px',
          background: 'var(--surface)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          textDecoration: 'none',
          marginBottom: 16,
        }}
      >
        <div>
          <p style={{ fontFamily: 'DM Sans', fontSize: 14, color: 'var(--text)', fontWeight: 500, marginBottom: 2 }}>
            Get a token from Up
          </p>
          <p style={{ fontFamily: 'DM Sans', fontSize: 12, color: 'var(--text-2)' }}>api.up.com.au/getting_started</p>
        </div>
        <ExternalLink size={16} color="var(--text-2)" strokeWidth={1.75} />
      </a>

      <div
        style={{
          padding: '14px 18px',
          background: 'rgba(202,255,51,0.06)',
          border: '1px solid rgba(202,255,51,0.18)',
          borderRadius: 16,
          fontFamily: 'DM Sans',
          fontSize: 13,
          lineHeight: 1.5,
          color: 'var(--text-2)',
        }}
      >
        Tokens are stored encrypted and only used to fetch your own transactions. You can revoke access at any time from
        Up.
      </div>
    </div>
  )
}

function StepBuckets({
  selected,
  setSelected,
  customBuckets,
  setCustomBuckets,
  showCustomInput,
  setShowCustomInput,
  customDraft,
  setCustomDraft,
  error,
}: {
  selected: Set<string>
  setSelected: (v: Set<string>) => void
  customBuckets: string[]
  setCustomBuckets: (v: string[]) => void
  showCustomInput: boolean
  setShowCustomInput: (v: boolean) => void
  customDraft: string
  setCustomDraft: (v: string) => void
  error: string | null
}) {
  const toggle = (key: string) => {
    const next = new Set(selected)
    if (next.has(key)) next.delete(key)
    else next.add(key)
    setSelected(next)
  }
  const addCustom = () => {
    const name = customDraft.trim()
    if (!name) return
    setCustomBuckets([...customBuckets, name])
    setCustomDraft('')
    setShowCustomInput(false)
  }
  return (
    <div>
      <StepHeading
        kicker="Organise"
        title="Pick your buckets"
        body="Buckets are envelopes for your money. Pick a few to start — you can always add more later."
      />

      <div
        style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(2, 1fr)',
          gap: 10,
          marginBottom: 14,
        }}
      >
        {SUGGESTED_BUCKETS.map((b) => {
          const active = selected.has(b.key)
          return (
            <button
              key={b.key}
              onClick={() => toggle(b.key)}
              className="pressable"
              style={{
                textAlign: 'left',
                background: active ? 'var(--accent-dim)' : 'var(--surface)',
                border: `1.5px solid ${active ? 'var(--accent)' : 'var(--border)'}`,
                borderRadius: 14,
                padding: '14px 14px',
                cursor: 'pointer',
                position: 'relative',
                transition: 'all 0.15s',
              }}
            >
              <p
                style={{
                  fontFamily: 'Syne',
                  fontWeight: 700,
                  fontSize: 14,
                  color: active ? 'var(--accent)' : 'var(--text)',
                  marginBottom: 3,
                }}
              >
                {b.name}
              </p>
              <p style={{ fontFamily: 'DM Sans', fontSize: 12, color: 'var(--text-2)', lineHeight: 1.35 }}>
                {b.description}
              </p>
              <div
                style={{
                  position: 'absolute',
                  top: 12,
                  right: 12,
                  width: 18,
                  height: 18,
                  borderRadius: '50%',
                  border: `1.5px solid ${active ? 'var(--accent)' : 'var(--text-3)'}`,
                  background: active ? 'var(--accent)' : 'transparent',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                {active && <Check size={11} color="#000" strokeWidth={3} />}
              </div>
            </button>
          )
        })}
      </div>

      {customBuckets.length > 0 && (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 14 }}>
          {customBuckets.map((name, i) => (
            <span
              key={i}
              style={{
                fontFamily: 'Syne',
                fontWeight: 700,
                fontSize: 12,
                color: 'var(--accent)',
                background: 'var(--accent-dim)',
                border: '1px solid rgba(202,255,51,0.25)',
                borderRadius: 999,
                padding: '7px 12px',
                display: 'inline-flex',
                alignItems: 'center',
                gap: 6,
              }}
            >
              {name}
              <button
                onClick={() => setCustomBuckets(customBuckets.filter((_, idx) => idx !== i))}
                style={{
                  background: 'none',
                  border: 'none',
                  color: 'var(--accent)',
                  cursor: 'pointer',
                  padding: 0,
                  fontSize: 14,
                  lineHeight: 1,
                }}
                aria-label={`Remove ${name}`}
              >
                ×
              </button>
            </span>
          ))}
        </div>
      )}

      {showCustomInput ? (
        <div style={{ display: 'flex', gap: 8 }}>
          <input
            autoFocus
            value={customDraft}
            onChange={(e) => setCustomDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') addCustom()
              if (e.key === 'Escape') {
                setShowCustomInput(false)
                setCustomDraft('')
              }
            }}
            placeholder="Bucket name"
            style={{
              flex: 1,
              background: 'var(--surface-2)',
              border: '1px solid var(--accent)',
              borderRadius: 12,
              padding: '13px 14px',
              color: 'var(--text)',
              fontFamily: 'DM Sans',
              fontSize: 15,
              outline: 'none',
              minWidth: 0,
            }}
          />
          <button
            onClick={addCustom}
            disabled={!customDraft.trim()}
            className="pressable"
            style={{
              background: customDraft.trim() ? 'var(--accent)' : 'var(--surface-2)',
              border: 'none',
              borderRadius: 12,
              padding: '13px 16px',
              color: customDraft.trim() ? '#000' : 'var(--text-2)',
              fontFamily: 'Syne',
              fontWeight: 700,
              fontSize: 13,
              cursor: 'pointer',
            }}
          >
            Add
          </button>
        </div>
      ) : (
        <button
          onClick={() => setShowCustomInput(true)}
          className="pressable"
          style={{
            width: '100%',
            background: 'transparent',
            border: '1.5px dashed var(--border)',
            borderRadius: 14,
            padding: '14px 16px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: 8,
            cursor: 'pointer',
            color: 'var(--text-2)',
            fontFamily: 'DM Sans',
            fontSize: 14,
          }}
        >
          <Plus size={15} strokeWidth={1.75} />
          Add custom bucket
        </button>
      )}
      {error && (
        <p
          style={{
            fontSize: 12,
            color: 'var(--red)',
            marginTop: 12,
            fontFamily: 'DM Sans',
            display: 'flex',
            alignItems: 'center',
            gap: 5,
          }}
        >
          <X size={12} strokeWidth={2.25} /> {error}
        </p>
      )}
    </div>
  )
}

function StepTrickle({
  buckets,
  trickleBucketId,
  setTrickleBucketId,
  trickleAmount,
  setTrickleAmount,
  tricklePeriod,
  setTricklePeriod,
  error,
}: {
  buckets: Bucket[]
  trickleBucketId: string
  setTrickleBucketId: (v: string) => void
  trickleAmount: string
  setTrickleAmount: (v: string) => void
  tricklePeriod: string
  setTricklePeriod: (v: string) => void
  error: string | null
}) {
  const periodLabel = PERIODS.find((p) => p.value === tricklePeriod)?.label.toLowerCase() ?? 'weekly'
  const previewAmount = parseFloat(trickleAmount) || 0
  const selectedBucket = buckets.find((b) => b.bucket_id === trickleBucketId)
  return (
    <div>
      <StepHeading
        kicker="Automate"
        title="Drip into a goal"
        body="A trickle moves a fixed amount into a bucket on a schedule. Great for savings, holidays, or any goal you’re building toward."
      />

      <div
        style={{
          background: 'var(--surface)',
          border: '1px solid var(--border)',
          borderRadius: 18,
          padding: '20px',
          marginBottom: 16,
        }}
      >
        <p
          style={{
            fontFamily: 'Syne',
            fontWeight: 700,
            fontSize: 11,
            letterSpacing: '0.08em',
            color: 'var(--text-2)',
            marginBottom: 8,
            textTransform: 'uppercase',
          }}
        >
          Into bucket
        </p>
        <select
          value={trickleBucketId}
          onChange={(e) => setTrickleBucketId(e.target.value)}
          disabled={buckets.length === 0}
          style={{
            width: '100%',
            background: 'var(--surface-2)',
            border: '1px solid var(--border)',
            borderRadius: 12,
            padding: '14px 16px',
            color: 'var(--text)',
            fontFamily: 'DM Sans',
            fontSize: 15,
            outline: 'none',
            appearance: 'none',
            marginBottom: 16,
          }}
        >
          {buckets.length === 0 && <option value="">No buckets created</option>}
          {buckets.map((b) => (
            <option key={b.bucket_id} value={b.bucket_id}>
              {b.name}
            </option>
          ))}
        </select>

        <p
          style={{
            fontFamily: 'Syne',
            fontWeight: 700,
            fontSize: 11,
            letterSpacing: '0.08em',
            color: 'var(--text-2)',
            marginBottom: 8,
            textTransform: 'uppercase',
          }}
        >
          Amount
        </p>
        <div style={{ position: 'relative', marginBottom: 16 }}>
          <span
            style={{
              position: 'absolute',
              left: 16,
              top: '50%',
              transform: 'translateY(-50%)',
              color: 'var(--text-2)',
              fontFamily: 'DM Sans',
              fontSize: 16,
            }}
          >
            $
          </span>
          <input
            type="number"
            min="0"
            step="0.01"
            value={trickleAmount}
            onChange={(e) => setTrickleAmount(e.target.value)}
            placeholder="0.00"
            style={{
              width: '100%',
              background: 'var(--surface-2)',
              border: '1px solid var(--border)',
              borderRadius: 12,
              padding: '14px 16px 14px 28px',
              color: 'var(--text)',
              fontFamily: 'JetBrains Mono',
              fontSize: 16,
              outline: 'none',
            }}
          />
        </div>

        <p
          style={{
            fontFamily: 'Syne',
            fontWeight: 700,
            fontSize: 11,
            letterSpacing: '0.08em',
            color: 'var(--text-2)',
            marginBottom: 8,
            textTransform: 'uppercase',
          }}
        >
          How often
        </p>
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(4, 1fr)',
            gap: 6,
          }}
        >
          {PERIODS.map((p) => {
            const active = tricklePeriod === p.value
            return (
              <button
                key={p.value}
                onClick={() => setTricklePeriod(p.value)}
                className="pressable"
                style={{
                  background: active ? 'var(--accent)' : 'var(--surface-2)',
                  border: `1px solid ${active ? 'var(--accent)' : 'var(--border)'}`,
                  borderRadius: 10,
                  padding: '11px 4px',
                  color: active ? '#000' : 'var(--text-2)',
                  fontFamily: 'Syne',
                  fontWeight: 700,
                  fontSize: 11,
                  letterSpacing: '0.04em',
                  cursor: 'pointer',
                  textTransform: 'uppercase',
                }}
              >
                {p.label}
              </button>
            )
          })}
        </div>
      </div>

      {previewAmount > 0 && selectedBucket && (
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 12,
            padding: '14px 18px',
            background: 'rgba(202,255,51,0.06)',
            border: '1px solid rgba(202,255,51,0.18)',
            borderRadius: 16,
          }}
        >
          <Droplets size={18} color="var(--accent)" strokeWidth={1.75} style={{ flexShrink: 0 }} />
          <p style={{ fontFamily: 'DM Sans', fontSize: 13, lineHeight: 1.5, color: 'var(--text-2)' }}>
            ${previewAmount.toFixed(2)} will trickle into{' '}
            <span style={{ color: 'var(--text)', fontWeight: 600 }}>{selectedBucket.name}</span> {periodLabel}.
          </p>
        </div>
      )}
      {error && (
        <p
          style={{
            fontSize: 12,
            color: 'var(--red)',
            marginTop: 12,
            fontFamily: 'DM Sans',
            display: 'flex',
            alignItems: 'center',
            gap: 5,
          }}
        >
          <X size={12} strokeWidth={2.25} /> {error}
        </p>
      )}
    </div>
  )
}

function StepDone({
  tokenSaved,
  bucketCount,
  hasTrickle,
}: {
  tokenSaved: boolean
  bucketCount: number
  hasTrickle: boolean
}) {
  const items = [
    { ok: tokenSaved, label: tokenSaved ? 'Up Bank linked' : 'Up Bank not linked' },
    { ok: bucketCount > 0, label: `${bucketCount} bucket${bucketCount === 1 ? '' : 's'} created` },
    { ok: hasTrickle, label: hasTrickle ? 'Trickle scheduled' : 'No trickle yet' },
  ]
  return (
    <div style={{ paddingTop: 24 }}>
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          marginBottom: 32,
        }}
      >
        <div
          style={{
            width: 92,
            height: 92,
            borderRadius: '50%',
            background: 'var(--accent-dim)',
            border: '1.5px solid rgba(202, 255, 51, 0.3)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            marginBottom: 24,
          }}
        >
          <Check size={42} color="var(--accent)" strokeWidth={2} />
        </div>
        <h1
          style={{
            fontFamily: 'Syne',
            fontWeight: 800,
            fontSize: 30,
            color: 'var(--text)',
            marginBottom: 10,
            letterSpacing: '-0.01em',
            textAlign: 'center',
          }}
        >
          You’re all set
        </h1>
        <p
          style={{
            fontFamily: 'DM Sans',
            fontSize: 15,
            lineHeight: 1.55,
            color: 'var(--text-2)',
            textAlign: 'center',
            maxWidth: 320,
          }}
        >
          Float will start sorting your transactions into buckets. New ones land automatically.
        </p>
      </div>

      <div
        style={{
          background: 'var(--surface)',
          border: '1px solid var(--border)',
          borderRadius: 16,
          overflow: 'hidden',
        }}
      >
        {items.map((item, i) => (
          <div
            key={i}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 12,
              padding: '15px 18px',
              borderBottom: i < items.length - 1 ? '1px solid var(--border)' : 'none',
            }}
          >
            <div
              style={{
                width: 22,
                height: 22,
                borderRadius: '50%',
                background: item.ok ? 'var(--accent)' : 'var(--surface-2)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0,
              }}
            >
              {item.ok && <Check size={13} color="#000" strokeWidth={3} />}
            </div>
            <p
              style={{
                fontFamily: 'DM Sans',
                fontSize: 14,
                color: item.ok ? 'var(--text)' : 'var(--text-2)',
              }}
            >
              {item.label}
            </p>
          </div>
        ))}
      </div>
    </div>
  )
}
