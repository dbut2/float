import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { type BucketHealth, api } from '../lib/api'

// ---------------------------------------------------------------------------
// Colour helpers
// ---------------------------------------------------------------------------

const STATUS_COLOR: Record<string, string> = {
  great: 'var(--green)',
  ok: 'var(--yellow)',
  warn: '#fb923c',
  critical: 'var(--red)',
  stale: 'var(--text-3)',
}

const STATUS_LABEL: Record<string, string> = {
  great: 'Great',
  ok: 'OK',
  warn: 'Warning',
  critical: 'Critical',
  stale: 'Stale',
}

// ---------------------------------------------------------------------------
// 30-day forecast chart
// ---------------------------------------------------------------------------

interface ForecastPoint {
  day: number // 0..29
  balance: number // in dollars
}

function buildForecast(buckets: BucketHealth[]): ForecastPoint[] {
  const activeBuckets = buckets.filter((b) => b.has_trickle)
  if (activeBuckets.length === 0) return []

  const points: ForecastPoint[] = []
  for (let day = 0; day <= 30; day++) {
    let total = 0
    for (const b of activeBuckets) {
      // Start from current balance.
      let bal = b.balance
      // Simulate spending at current daily rate (if daily_allowance > 0 use it, else $0).
      const dailySpend = b.daily_allowance > 0 ? b.daily_allowance : 0
      bal -= dailySpend * day
      // Add any trickle top-ups that land within the next `day` days.
      if (b.days_until_trickle > 0 && b.days_until_trickle <= day) {
        const tricklePeriodDays: Record<string, number> = {
          daily: 1, weekly: 7, fortnightly: 14, monthly: 30,
        }
        const period = tricklePeriodDays[b.period ?? ''] ?? 30
        const trickleCount = Math.floor((day - b.days_until_trickle) / period) + 1
        bal += trickleCount * b.trickle_amount
      }
      total += Math.max(bal, 0)
    }
    points.push({ day, balance: total })
  }
  return points
}

function ForecastChart({ buckets }: { buckets: BucketHealth[] }) {
  const points = buildForecast(buckets)
  if (points.length === 0) return null

  const W = 600
  const H = 140
  const PAD = { top: 12, right: 12, bottom: 24, left: 48 }
  const innerW = W - PAD.left - PAD.right
  const innerH = H - PAD.top - PAD.bottom

  const maxBal = Math.max(...points.map((p) => p.balance), 0.01)

  const toX = (day: number) => PAD.left + (day / 30) * innerW
  const toY = (bal: number) => PAD.top + innerH - (bal / maxBal) * innerH

  const pathD = points
    .map((p, i) => `${i === 0 ? 'M' : 'L'}${toX(p.day).toFixed(1)},${toY(p.balance).toFixed(1)}`)
    .join(' ')

  // Fill area under curve
  const fillD = `${pathD} L${toX(30).toFixed(1)},${(PAD.top + innerH).toFixed(1)} L${toX(0).toFixed(1)},${(PAD.top + innerH).toFixed(1)} Z`

  // Y-axis labels
  const yLabels = [0, 0.5, 1].map((pct) => ({
    y: PAD.top + innerH - pct * innerH,
    label: `$${Math.round(maxBal * pct)}`,
  }))

  // X-axis tick every 7 days
  const xTicks = [0, 7, 14, 21, 30]

  return (
    <svg
      viewBox={`0 0 ${W} ${H}`}
      style={{ width: '100%', height: 'auto', display: 'block' }}
      aria-label="30-day balance forecast"
    >
      <defs>
        <linearGradient id="fg-grad" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0%" stopColor="var(--accent)" stopOpacity="0.22" />
          <stop offset="100%" stopColor="var(--accent)" stopOpacity="0.02" />
        </linearGradient>
      </defs>

      {/* Grid lines */}
      {yLabels.map(({ y, label }) => (
        <g key={label}>
          <line x1={PAD.left} y1={y} x2={PAD.left + innerW} y2={y} stroke="var(--border)" strokeWidth={1} />
          <text x={PAD.left - 6} y={y + 4} textAnchor="end" fill="var(--text-3)" fontSize={10} fontFamily="JetBrains Mono, monospace">
            {label}
          </text>
        </g>
      ))}

      {/* X-axis ticks */}
      {xTicks.map((d) => (
        <text key={d} x={toX(d)} y={H - 6} textAnchor="middle" fill="var(--text-3)" fontSize={10} fontFamily="DM Sans, sans-serif">
          {d === 0 ? 'Today' : `+${d}d`}
        </text>
      ))}

      {/* Fill */}
      <path d={fillD} fill="url(#fg-grad)" />

      {/* Line */}
      <path d={pathD} fill="none" stroke="var(--accent)" strokeWidth={2} strokeLinejoin="round" />
    </svg>
  )
}

// ---------------------------------------------------------------------------
// Mini health row (used in bucket ranking)
// ---------------------------------------------------------------------------

function HealthRow({ bh, onApply }: { bh: BucketHealth; onApply: (bh: BucketHealth) => void }) {
  const color = STATUS_COLOR[bh.status] ?? 'var(--text-3)'
  const label = STATUS_LABEL[bh.status] ?? bh.status

  const allowanceLabel = () => {
    if (!bh.has_trickle) return 'No trickle'
    const prefix = bh.status === 'warn' || bh.status === 'critical' || bh.is_at_risk ? 'Recovery' : 'Budget'
    return `${prefix}: $${Math.abs(bh.daily_allowance).toFixed(2)}/day`
  }

  return (
    <div
      style={{
        padding: '14px 0',
        borderBottom: '1px solid var(--border)',
        display: 'flex',
        flexDirection: 'column',
        gap: 8,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 8 }}>
        <div style={{ flex: 1, minWidth: 0 }}>
          <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 14, color: 'var(--text)' }} className="line-clamp-1">
            {bh.bucket_name}
          </p>
          <p style={{ fontSize: 12, color: 'var(--text-2)', marginTop: 3 }}>
            {allowanceLabel()}
            {bh.has_trickle && bh.days_until_trickle > 0 && (
              <span style={{ color: 'var(--text-3)' }}> · {bh.days_until_trickle}d until trickle</span>
            )}
          </p>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexShrink: 0 }}>
          <span style={{
            fontFamily: 'Syne', fontWeight: 700, fontSize: 11, letterSpacing: '0.04em',
            color, background: 'rgba(0,0,0,0.3)', borderRadius: 6, padding: '3px 8px',
          }}>
            {label.toUpperCase()}
          </span>
          {(bh.status === 'warn' || bh.status === 'critical') && (
            <button
              onClick={() => onApply(bh)}
              style={{
                fontFamily: 'Syne', fontWeight: 700, fontSize: 11, letterSpacing: '0.04em',
                color: '#000', background: 'var(--accent)', border: 'none', borderRadius: 6,
                padding: '3px 10px', cursor: 'pointer',
              }}
            >
              Apply Fix
            </button>
          )}
        </div>
      </div>

      {/* Spend progress bar */}
      {bh.has_trickle && (
        <div style={{ height: 4, borderRadius: 2, background: 'rgba(255,255,255,0.07)', overflow: 'hidden' }}>
          <div style={{
            height: '100%',
            width: `${Math.min(bh.spent_pct * 100, 100)}%`,
            background: color,
            borderRadius: 2,
          }} />
        </div>
      )}
    </div>
  )
}

// ---------------------------------------------------------------------------
// Overall score ring
// ---------------------------------------------------------------------------

function ScoreRing({ score }: { score: number }) {
  const R = 44
  const C = 2 * Math.PI * R
  const fill = (score / 100) * C
  const color = score >= 70 ? 'var(--green)' : score >= 40 ? 'var(--yellow)' : 'var(--red)'

  return (
    <svg width={110} height={110} viewBox="0 0 110 110">
      <circle cx={55} cy={55} r={R} fill="none" stroke="var(--surface-2)" strokeWidth={8} />
      <circle
        cx={55} cy={55} r={R} fill="none"
        stroke={color}
        strokeWidth={8}
        strokeLinecap="round"
        strokeDasharray={`${fill} ${C}`}
        transform="rotate(-90 55 55)"
      />
      <text x={55} y={52} textAnchor="middle" fill="var(--text)" fontSize={22} fontWeight={700} fontFamily="JetBrains Mono, monospace">
        {score}
      </text>
      <text x={55} y={68} textAnchor="middle" fill="var(--text-3)" fontSize={11} fontFamily="DM Sans, sans-serif">
        / 100
      </text>
    </svg>
  )
}

// ---------------------------------------------------------------------------
// Apply suggestion modal
// ---------------------------------------------------------------------------

function ApplySuggestionModal({
  bucket,
  onClose,
}: {
  bucket: BucketHealth
  onClose: () => void
}) {
  const qc = useQueryClient()
  // Suggest increasing trickle by ~20% or by the deficit.
  const suggestedCents = Math.round(bucket.trickle_amount_cents * 1.2)

  const apply = useMutation({
    mutationFn: () =>
      api.applyTrickleSuggestion(bucket.bucket_id, suggestedCents, bucket.period ?? 'monthly'),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['health'] })
      qc.invalidateQueries({ queryKey: ['buckets'] })
      qc.invalidateQueries({ queryKey: ['trickle', bucket.bucket_id] })
      onClose()
    },
  })

  return (
    <>
      <div
        onClick={onClose}
        style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)',
          zIndex: 100, animation: 'fadeIn 0.2s ease forwards',
        }}
      />
      <div
        style={{
          position: 'fixed', top: '50%', left: '50%',
          transform: 'translate(-50%, -50%)',
          width: 'min(420px, 90vw)', zIndex: 101,
          background: 'var(--surface)', borderRadius: 20,
          padding: '24px', animation: 'fadeIn 0.2s ease forwards',
        }}
      >
        <p style={{ fontFamily: 'Syne', fontWeight: 800, fontSize: 20, color: 'var(--text)', marginBottom: 8 }}>
          Adjust trickle
        </p>
        <p style={{ fontFamily: 'DM Sans', fontSize: 14, color: 'var(--text-2)', marginBottom: 4 }}>
          "{bucket.bucket_name}" is over-budget. Current trickle: ${(bucket.trickle_amount).toFixed(2)} / {bucket.period}.
        </p>
        <p style={{ fontFamily: 'DM Sans', fontSize: 14, color: 'var(--text-2)', marginBottom: 20 }}>
          Suggested new amount: <span style={{ color: 'var(--accent)', fontWeight: 700 }}>${(suggestedCents / 100).toFixed(2)}</span> / {bucket.period} (+20%)
        </p>
        <button
          onClick={() => apply.mutate()}
          disabled={apply.isPending}
          className="pressable"
          style={{
            width: '100%', padding: '15px',
            background: 'var(--accent)', border: 'none', borderRadius: 12,
            color: '#000', fontFamily: 'Syne', fontWeight: 700, fontSize: 15,
            cursor: 'pointer', marginBottom: 10,
          }}
        >
          {apply.isPending ? 'Applying…' : 'Apply'}
        </button>
        <button
          onClick={onClose}
          style={{
            width: '100%', padding: '15px',
            background: 'transparent', border: 'none',
            color: 'var(--text-2)', fontFamily: 'DM Sans', fontSize: 15, cursor: 'pointer',
          }}
        >
          Cancel
        </button>
      </div>
    </>
  )
}

// ---------------------------------------------------------------------------
// Main page
// ---------------------------------------------------------------------------

export default function Health() {
  const { data: summary, isLoading } = useQuery({
    queryKey: ['health'],
    queryFn: api.getHealth,
    staleTime: 30_000,
  })

  const [applyBucket, setApplyBucket] = useState<BucketHealth | null>(null)

  if (isLoading || !summary) {
    return (
      <div style={{ padding: '24px 20px' }}>
        <div className="shimmer" style={{ height: 160, borderRadius: 20, marginBottom: 16 }} />
        <div className="shimmer" style={{ height: 200, borderRadius: 20, marginBottom: 16 }} />
        <div className="shimmer" style={{ height: 300, borderRadius: 20 }} />
      </div>
    )
  }

  // Sort buckets: critical first, then warn, ok, great, stale.
  const statusOrder: Record<string, number> = { critical: 0, warn: 1, ok: 2, great: 3, stale: 4 }
  const sorted = [...summary.buckets].sort(
    (a, b) => (statusOrder[a.status] ?? 4) - (statusOrder[b.status] ?? 4),
  )

  const activeBuckets = summary.buckets.filter((b) => b.has_trickle)
  const recommendBuckets = summary.buckets.filter(
    (b) => b.status === 'warn' || b.status === 'critical',
  )

  const scoreLabel =
    summary.overall_score >= 80
      ? 'Finances looking healthy'
      : summary.overall_score >= 50
        ? 'Some buckets need attention'
        : 'Multiple buckets at risk'

  return (
    <div style={{ padding: '24px 20px', minHeight: '100%' }}>
      {/* Header */}
      <div className="animate-fade-up" style={{ marginBottom: 24, opacity: 0 }}>
        <h1 style={{ fontFamily: 'Syne', fontWeight: 800, fontSize: 22, color: 'var(--text)' }}>
          Health
        </h1>
      </div>

      {/* Overall score card */}
      <div
        className="animate-fade-up"
        style={{
          background: 'var(--surface)', borderRadius: 20, padding: '24px',
          marginBottom: 16, opacity: 0,
          display: 'flex', alignItems: 'center', gap: 24,
        }}
      >
        <ScoreRing score={summary.overall_score} />
        <div>
          <p style={{ fontFamily: 'DM Sans', fontSize: 15, color: 'var(--text)', marginBottom: 8 }}>
            {scoreLabel}
          </p>
          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 10 }}>
            {summary.healthy_count > 0 && (
              <span style={{ fontFamily: 'DM Sans', fontSize: 13, color: 'var(--green)' }}>
                ✓ {summary.healthy_count} healthy
              </span>
            )}
            {summary.at_risk_count > 0 && (
              <span style={{ fontFamily: 'DM Sans', fontSize: 13, color: 'var(--red)' }}>
                ✗ {summary.at_risk_count} at risk
              </span>
            )}
            {summary.stale_count > 0 && (
              <span style={{ fontFamily: 'DM Sans', fontSize: 13, color: 'var(--text-3)' }}>
                — {summary.stale_count} no trickle
              </span>
            )}
          </div>
        </div>
      </div>

      {/* 30-day forecast chart */}
      {activeBuckets.length > 0 && (
        <div
          className="animate-fade-up"
          style={{
            background: 'var(--surface)', borderRadius: 20, padding: '20px 20px 16px',
            marginBottom: 16, opacity: 0,
          }}
        >
          <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 12, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 12 }}>
            30-DAY FORECAST
          </p>
          <ForecastChart buckets={activeBuckets} />
        </div>
      )}

      {/* Recommendations */}
      {recommendBuckets.length > 0 && (
        <div
          className="animate-fade-up"
          style={{
            background: 'var(--surface)', borderRadius: 20, padding: '20px',
            marginBottom: 16, opacity: 0,
          }}
        >
          <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 12, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 4 }}>
            RECOMMENDATIONS
          </p>
          <p style={{ fontFamily: 'DM Sans', fontSize: 13, color: 'var(--text-3)', marginBottom: 16 }}>
            These buckets are spending faster than their trickle replenishes them.
          </p>
          {recommendBuckets.map((bh) => (
            <div key={bh.bucket_id} style={{ marginBottom: 12 }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 10 }}>
                <div>
                  <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 14, color: 'var(--text)' }}>
                    {bh.bucket_name}
                  </p>
                  <p style={{ fontFamily: 'DM Sans', fontSize: 13, color: 'var(--text-2)', marginTop: 2 }}>
                    Increase trickle from ${bh.trickle_amount.toFixed(2)} → ${(bh.trickle_amount * 1.2).toFixed(2)} / {bh.period}
                  </p>
                </div>
                <button
                  onClick={() => setApplyBucket(bh)}
                  className="pressable"
                  style={{
                    fontFamily: 'Syne', fontWeight: 700, fontSize: 13,
                    color: '#000', background: 'var(--accent)', border: 'none',
                    borderRadius: 10, padding: '8px 16px', cursor: 'pointer', flexShrink: 0,
                  }}
                >
                  Apply
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Buckets ranked by health */}
      <div
        className="animate-fade-up"
        style={{ background: 'var(--surface)', borderRadius: 20, padding: '20px', opacity: 0 }}
      >
        <p style={{ fontFamily: 'Syne', fontWeight: 700, fontSize: 12, letterSpacing: '0.08em', color: 'var(--text-2)', marginBottom: 4 }}>
          BUCKETS · {sorted.length}
        </p>
        {sorted.length === 0 ? (
          <p style={{ fontFamily: 'DM Sans', fontSize: 14, color: 'var(--text-3)', marginTop: 12 }}>
            No buckets found.
          </p>
        ) : (
          sorted.map((bh) => (
            <HealthRow key={bh.bucket_id} bh={bh} onApply={setApplyBucket} />
          ))
        )}
      </div>

      {applyBucket && (
        <ApplySuggestionModal bucket={applyBucket} onClose={() => setApplyBucket(null)} />
      )}
    </div>
  )
}
