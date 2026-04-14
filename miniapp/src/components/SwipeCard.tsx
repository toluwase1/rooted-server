import { useState, useRef } from 'react'

interface Candidate {
  user_id: string
  first_name: string
  age: number
  city: string
  country: string
  heritage: string[]
  diaspora_tag: string
  intention: string
  faith: string
  verification: string
  primary_photo: string
  match_score: number
}

interface Props {
  candidate: Candidate
  onSwipe: (action: 'like' | 'pass', comment?: string) => void
}

export default function SwipeCard({ candidate, onSwipe }: Props) {
  const [showDetails, setShowDetails] = useState(false)
  const [dragX, setDragX] = useState(0)
  const startX = useRef(0)
  const isDragging = useRef(false)

  const handleTouchStart = (e: React.TouchEvent) => {
    startX.current = e.touches[0].clientX
    isDragging.current = true
  }

  const handleTouchMove = (e: React.TouchEvent) => {
    if (!isDragging.current) return
    const diff = e.touches[0].clientX - startX.current
    setDragX(diff)
  }

  const handleTouchEnd = () => {
    isDragging.current = false
    if (dragX > 100) {
      haptic('medium')
      onSwipe('like')
    } else if (dragX < -100) {
      haptic('light')
      onSwipe('pass')
    }
    setDragX(0)
  }

  const diasporaLabel = (tag: string) => {
    const labels: Record<string, string> = {
      born_in_africa: 'Born in Africa',
      diaspora_1st: '1st Gen Diaspora',
      diaspora_2nd: '2nd Gen Diaspora',
      returnee: 'Returnee',
      explorer: 'Explorer',
    }
    return labels[tag] || tag
  }

  const haptic = (style: 'light' | 'medium' | 'heavy') => {
    window.Telegram?.WebApp?.HapticFeedback?.impactOccurred(style)
  }

  const bgColor = dragX > 50 ? 'rgba(129,178,154,0.3)' : dragX < -50 ? 'rgba(255,107,107,0.3)' : 'transparent'

  return (
    <div style={{ background: bgColor, borderRadius: 'var(--radius)', transition: 'background 0.2s' }}>
      <div
        className="swipe-card"
        style={{ transform: `translateX(${dragX}px) rotate(${dragX * 0.05}deg)` }}
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
        onClick={() => setShowDetails(!showDetails)}
      >
        {candidate.primary_photo ? (
          <img src={candidate.primary_photo} alt={candidate.first_name} />
        ) : (
          <div style={{
            width: '100%', height: '100%', display: 'flex',
            alignItems: 'center', justifyContent: 'center',
            fontSize: '64px', background: 'linear-gradient(135deg, #E07A5F, #81B29A)'
          }}>
            {candidate.first_name[0]}
          </div>
        )}

        <div className="swipe-card-info">
          <h2>
            {candidate.first_name}, {candidate.age}
            {candidate.verification === 'photo_verified' && ' ✓'}
          </h2>
          <p>{candidate.city}, {candidate.country}</p>

          <div style={{ display: 'flex', gap: '6px', marginTop: '8px', flexWrap: 'wrap' }}>
            {candidate.heritage.map((h) => (
              <span key={h} className="badge badge-heritage">{h}</span>
            ))}
            <span className="badge badge-diaspora">{diasporaLabel(candidate.diaspora_tag)}</span>
          </div>

          {showDetails && (
            <div style={{ marginTop: '12px', fontSize: '13px' }}>
              {candidate.faith && <p>Faith: {candidate.faith}</p>}
              <p>Looking for: {candidate.intention}</p>
            </div>
          )}
        </div>
      </div>

      <div className="swipe-actions">
        <button className="swipe-btn swipe-btn-pass" onClick={() => { haptic('light'); onSwipe('pass') }}>
          ✕
        </button>
        <button className="swipe-btn swipe-btn-like" onClick={() => { haptic('medium'); onSwipe('like') }}>
          ♡
        </button>
      </div>
    </div>
  )
}
