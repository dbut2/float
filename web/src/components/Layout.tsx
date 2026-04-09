import { Outlet, NavLink, useLocation } from 'react-router-dom'
import { House, Settings2 } from 'lucide-react'
import { useMediaQuery } from '../hooks/useMediaQuery'

// ECG / heart-pulse icon drawn as an inline SVG component.
function HeartPulseIcon({ size = 22, color = 'currentColor', strokeWidth = 1.75 }: { size?: number; color?: string; strokeWidth?: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth={strokeWidth} strokeLinecap="round" strokeLinejoin="round">
      <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
    </svg>
  )
}

const navItems = [
  { to: '/', label: 'Home', Icon: House, exact: true },
  { to: '/health', label: 'Health', Icon: HeartPulseIcon, exact: true },
  { to: '/settings', label: 'Settings', Icon: Settings2, exact: false },
]

export default function Layout() {
  const { pathname } = useLocation()
  const isDesktop = useMediaQuery('(min-width: 768px)')
  const today = new Date().toLocaleDateString('en-AU', {
    weekday: 'long', month: 'long', day: 'numeric',
  })

  if (isDesktop) {
    return (
      <div style={{ display: 'flex', height: '100%', background: 'var(--bg)' }}>
        {/* Left sidebar */}
        <div
          style={{
            width: 220,
            flexShrink: 0,
            display: 'flex',
            flexDirection: 'column',
            background: 'var(--surface)',
            borderRight: '1px solid var(--border)',
            padding: '32px 20px',
          }}
        >
          <div
            style={{
              fontFamily: 'Syne',
              fontWeight: 800,
              fontSize: 22,
              color: 'var(--accent)',
              letterSpacing: '-0.02em',
              marginBottom: 8,
            }}
          >
            FLOAT
          </div>
          <p
            style={{
              fontSize: 12,
              color: 'var(--text-2)',
              fontFamily: 'DM Sans',
              marginBottom: 40,
            }}
          >
            {today}
          </p>
          <nav style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            {navItems.map(({ to, label, Icon, exact }) => {
              const active = exact ? pathname === to : pathname.startsWith(to)
              return (
                <NavLink
                  key={to}
                  to={to}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 12,
                    padding: '10px 14px',
                    borderRadius: 10,
                    textDecoration: 'none',
                    background: active ? 'var(--accent-dim)' : 'transparent',
                    transition: 'background 0.15s',
                  }}
                >
                  <Icon size={18} color={active ? 'var(--accent)' : 'var(--text-2)'} strokeWidth={1.75} />
                  <span
                    style={{
                      fontFamily: 'Syne',
                      fontWeight: active ? 700 : 500,
                      fontSize: 14,
                      color: active ? 'var(--accent)' : 'var(--text-2)',
                      transition: 'color 0.15s',
                    }}
                  >
                    {label}
                  </span>
                </NavLink>
              )
            })}
          </nav>
        </div>

        {/* Main content */}
        <div style={{ flex: 1, overflowY: 'auto', overflowX: 'hidden' }}>
          <div style={{ maxWidth: 680, margin: '0 auto', padding: '32px 40px' }}>
            <Outlet />
          </div>
        </div>
      </div>
    )
  }

  // Mobile layout (unchanged)
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        height: '100%',
        background: 'var(--bg)',
        position: 'relative',
        overflow: 'hidden',
      }}
    >
      {/* Global header — lives above the scroll container so it never moves */}
      <div
        style={{
          flexShrink: 0,
          background: 'var(--bg)',
          zIndex: 20,
        }}
      >
        <div
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '14px 20px',
          }}
        >
          <p style={{ fontSize: 13, color: 'var(--text-2)', fontFamily: 'DM Sans' }}>
            {today}
          </p>
          <div
            style={{
              fontFamily: 'Syne',
              fontWeight: 800,
              fontSize: 20,
              color: 'var(--accent)',
              letterSpacing: '-0.02em',
            }}
          >
            FLOAT
          </div>
        </div>
      </div>

      {/* Page content */}
      <div
        style={{
          flex: 1,
          overflowY: 'auto',
          overflowX: 'hidden',
        }}
      >
        <Outlet />
      </div>

      {/* Bottom nav */}
      <nav
        style={{
          flexShrink: 0,
          background: 'rgba(14, 14, 14, 0.92)',
          backdropFilter: 'blur(20px)',
          borderTop: '1px solid var(--border)',
        }}
      >
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(3, 1fr)',
            height: 64,
          }}
        >
          {navItems.map(({ to, label, Icon, exact }) => {
            const active = exact ? pathname === to : pathname.startsWith(to)
            return (
              <NavLink
                key={to}
                to={to}
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: 4,
                  textDecoration: 'none',
                  transition: 'opacity 0.15s',
                }}
              >
                <Icon size={22} color={active ? 'var(--accent)' : 'var(--text-2)'} strokeWidth={1.75} />
                <span
                  style={{
                    fontSize: 10,
                    fontFamily: 'Syne, sans-serif',
                    fontWeight: active ? 700 : 500,
                    letterSpacing: '0.04em',
                    color: active ? 'var(--accent)' : 'var(--text-2)',
                    transition: 'color 0.15s',
                  }}
                >
                  {label}
                </span>
              </NavLink>
            )
          })}
        </div>
      </nav>
    </div>
  )
}
