import { useRef, useState, useEffect, useCallback } from 'react'

interface Options {
  onClose: () => void
  isOpen?: boolean
}

export function useDraggableSheet({ onClose, isOpen }: Options) {
  const handleRef = useRef<HTMLDivElement>(null)
  const [dragY, setDragY] = useState(0)
  const [isDragging, setIsDragging] = useState(false)
  const [dismissing, setDismissing] = useState(false)
  const [entryDone, setEntryDone] = useState(false)

  // Refs to read latest values inside event listeners without stale closures
  const startYRef = useRef(0)
  const dragYRef = useRef(0)
  const onCloseRef = useRef(onClose)
  onCloseRef.current = onClose

  // Reset state when sheet opens (for inline sheets that stay mounted)
  useEffect(() => {
    if (isOpen !== undefined && isOpen) {
      setDragY(0)
      dragYRef.current = 0
      setIsDragging(false)
      setDismissing(false)
      setEntryDone(false)
    }
  }, [isOpen])

  // Direct DOM listeners with { passive: false } so preventDefault works.
  // React's synthetic onTouchMove is passive in many environments and silently
  // ignores preventDefault, so we bypass it here.
  useEffect(() => {
    const el = handleRef.current
    if (!el) return

    const onStart = (e: TouchEvent) => {
      startYRef.current = e.touches[0].clientY
      setIsDragging(true)
    }

    const onMove = (e: TouchEvent) => {
      const delta = e.touches[0].clientY - startYRef.current
      const clamped = Math.max(0, delta)
      dragYRef.current = clamped
      setDragY(clamped)
      if (clamped > 0) e.preventDefault()
    }

    const onEnd = () => {
      setIsDragging(false)
      if (dragYRef.current > 100) {
        setDismissing(true)
        const target = window.innerHeight
        dragYRef.current = target
        setDragY(target)
        setTimeout(() => onCloseRef.current(), 280)
      } else {
        dragYRef.current = 0
        setDragY(0)
      }
    }

    el.addEventListener('touchstart', onStart, { passive: true })
    el.addEventListener('touchmove', onMove, { passive: false })
    el.addEventListener('touchend', onEnd, { passive: true })

    return () => {
      el.removeEventListener('touchstart', onStart)
      el.removeEventListener('touchmove', onMove)
      el.removeEventListener('touchend', onEnd)
    }
  }, []) // empty — all mutable values accessed via refs

  const onAnimationEnd = useCallback(() => {
    setEntryDone(true)
  }, [])

  let sheetStyle: React.CSSProperties
  if (!entryDone) {
    sheetStyle = {
      animation: 'sheetUp 0.35s cubic-bezier(0.32, 0.72, 0, 1) forwards',
    }
  } else if (isDragging) {
    sheetStyle = {
      transform: `translateY(${dragY}px)`,
      transition: 'none',
    }
  } else if (dismissing) {
    sheetStyle = {
      transform: `translateY(${dragY}px)`,
      transition: 'transform 0.28s cubic-bezier(0.32, 0.72, 0, 1)',
    }
  } else {
    sheetStyle = {
      transform: `translateY(${dragY}px)`,
      transition: 'transform 0.3s cubic-bezier(0.32, 0.72, 0, 1)',
    }
  }

  const opacity = Math.max(0, 1 - dragY / 300)
  const backdropStyle: React.CSSProperties = {
    opacity,
    transition: isDragging ? 'none' : 'opacity 0.3s ease',
  }

  return { handleRef, sheetStyle, backdropStyle, onAnimationEnd }
}
