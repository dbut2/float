import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Check, RefreshCw, X } from 'lucide-react'
import { api } from '../lib/api'

function Section({
  title,
  children,
}: {
  title: string
  children: React.ReactNode
}) {
  return (
    <div style={{ marginBottom: 28 }}>
      <p
        style={{
          fontFamily: 'Syne',
          fontWeight: 700,
          fontSize: 12,
          letterSpacing: '0.08em',
          color: 'var(--text-2)',
          marginBottom: 12,
        }}
      >
        {title}
      </p>
      <div
        style={{
          background: 'var(--surface)',
          borderRadius: 16,
          overflow: 'hidden',
          border: '1px solid var(--border)',
        }}
      >
        {children}
      </div>
    </div>
  )
}

function Row({
  label,
  sublabel,
  right,
  onClick,
}: {
  label: string
  sublabel?: string
  right?: React.ReactNode
  onClick?: () => void
}) {
  return (
    <div
      onClick={onClick}
      className={onClick ? 'pressable' : ''}
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '16px 18px',
        borderBottom: '1px solid var(--border)',
        cursor: onClick ? 'pointer' : 'default',
      }}
    >
      <div style={{ flex: 1, minWidth: 0, marginRight: 12 }}>
        <p
          style={{
            fontFamily: 'DM Sans',
            fontSize: 15,
            color: 'var(--text)',
            fontWeight: 500,
          }}
        >
          {label}
        </p>
        {sublabel && (
          <p style={{ fontFamily: 'DM Sans', fontSize: 12, color: 'var(--text-2)', marginTop: 2 }}>
            {sublabel}
          </p>
        )}
      </div>
      {right}
    </div>
  )
}

export default function Settings() {
  const qc = useQueryClient()
  const [token, setToken] = useState('')
  const [tokenStatus, setTokenStatus] = useState<'idle' | 'ok' | 'fail'>('idle')

  const { data: me } = useQuery({ queryKey: ['user'], queryFn: api.getUser })

  const verifyToken = useMutation({
    mutationFn: () => api.setToken(token),
    onSuccess: () => {
      setTokenStatus('ok')
      setToken('')
      qc.invalidateQueries({ queryKey: ['buckets'] })
    },
    onError: () => setTokenStatus('fail'),
  })

  const sync = useMutation({
    mutationFn: api.sync,
  })

  return (
    <div style={{ padding: '24px 20px', minHeight: '100%' }}>
      <h1
        className="animate-fade-up"
        style={{
          fontFamily: 'Syne',
          fontWeight: 800,
          fontSize: 26,
          color: 'var(--text)',
          marginBottom: 28,
          opacity: 0,
        }}
      >
        Settings
      </h1>

      {/* Account info */}
      <Section title="ACCOUNT">
        <Row
          label={me?.email ?? '—'}
          sublabel="Signed in via Cloudflare Access"
          right={
            <span
              style={{
                fontFamily: 'JetBrains Mono',
                fontSize: 11,
                color: 'var(--text-2)',
                flexShrink: 0,
              }}
            >
              #{me?.user_id?.slice(0, 8)}
            </span>
          }
        />
      </Section>

      {/* Up Bank token */}
      <Section title="UP BANK">
        <div
          style={{
            padding: '16px 18px',
            borderBottom: '1px solid var(--border)',
          }}
        >
          <p style={{ fontFamily: 'DM Sans', fontSize: 14, color: 'var(--text-2)', marginBottom: 10 }}>
            Personal access token from{' '}
            <span style={{ color: 'var(--accent)' }}>up.com.au/api</span>
          </p>
          <div style={{ display: 'flex', gap: 8 }}>
            <input
              value={token}
              onChange={(e) => { setToken(e.target.value); setTokenStatus('idle') }}
              type="password"
              placeholder="up:yeah:…"
              style={{
                flex: 1,
                background: 'var(--surface-2)',
                border: `1px solid ${tokenStatus === 'ok' ? 'var(--green)' : tokenStatus === 'fail' ? 'var(--red)' : 'var(--border)'}`,
                borderRadius: 10,
                padding: '11px 14px',
                color: 'var(--text)',
                fontFamily: 'JetBrains Mono',
                fontSize: 13,
                outline: 'none',
                transition: 'border-color 0.2s',
              }}
            />
            <button
              onClick={() => verifyToken.mutate()}
              disabled={!token || verifyToken.isPending}
              className="pressable"
              style={{
                background: token ? 'var(--accent)' : 'var(--surface-2)',
                border: 'none',
                borderRadius: 10,
                padding: '11px 16px',
                color: token ? '#000' : 'var(--text-2)',
                fontFamily: 'Syne',
                fontWeight: 700,
                fontSize: 13,
                cursor: 'pointer',
                transition: 'all 0.15s',
                whiteSpace: 'nowrap',
              }}
            >
              {verifyToken.isPending ? '…' : 'Save'}
            </button>
          </div>
          {tokenStatus === 'ok' && (
            <p style={{ fontSize: 12, color: 'var(--green)', marginTop: 8, fontFamily: 'DM Sans', display: 'flex', alignItems: 'center', gap: 4 }}>
              <Check size={12} strokeWidth={1.75} /> Token saved
            </p>
          )}
          {tokenStatus === 'fail' && (
            <p style={{ fontSize: 12, color: 'var(--red)', marginTop: 8, fontFamily: 'DM Sans', display: 'flex', alignItems: 'center', gap: 4 }}>
              <X size={12} strokeWidth={1.75} /> Failed to save token
            </p>
          )}
        </div>
      </Section>

      {/* Sync */}
      <Section title="DATA">
        <Row
          label="Sync transactions"
          sublabel="Fetch latest from Up Bank"
          onClick={() => sync.mutate()}
          right={
            sync.isPending ? (
              <span style={{ fontFamily: 'DM Sans', fontSize: 13, color: 'var(--text-2)' }}>
                Syncing…
              </span>
            ) : sync.isSuccess ? (
              <span style={{ fontFamily: 'DM Sans', fontSize: 13, color: 'var(--green)', display: 'flex', alignItems: 'center', gap: 4 }}>
                Done <Check size={13} strokeWidth={1.75} />
              </span>
            ) : (
              <RefreshCw size={18} color="var(--text-2)" strokeWidth={1.75} />
            )
          }
        />
      </Section>

      {/* App info */}
      <div style={{ textAlign: 'center', paddingBottom: 8 }}>
        <p
          style={{
            fontFamily: 'Syne',
            fontWeight: 800,
            fontSize: 16,
            color: 'var(--text-3)',
            letterSpacing: '-0.01em',
          }}
        >
          FLOAT
        </p>
        <p style={{ fontFamily: 'DM Sans', fontSize: 12, color: 'var(--dim)', marginTop: 4 }}>
          Built on Up Bank
        </p>
      </div>
    </div>
  )
}
