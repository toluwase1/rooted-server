import { useEffect, useState } from 'react'
import { api } from '../api/client'
import SwipeCard from '../components/SwipeCard'

interface Props {
  user: any
}

export default function Explore({ user }: Props) {
  const [candidates, setCandidates] = useState<any[]>([])
  const [currentIndex, setCurrentIndex] = useState(0)
  const [loading, setLoading] = useState(true)
  const [swipesRemaining, setSwipesRemaining] = useState<number | null>(null)
  const [offset, setOffset] = useState(0)
  const [matchPopup, setMatchPopup] = useState<any>(null)

  useEffect(() => {
    loadExplore()
  }, [])

  const loadExplore = async (newOffset = 0) => {
    try {
      const data = await api.getExplore(newOffset)
      if (newOffset === 0) {
        setCandidates(data.candidates || [])
      } else {
        setCandidates((prev) => [...prev, ...(data.candidates || [])])
      }
      setSwipesRemaining(data.swipes_remaining)
      setOffset(newOffset)
    } catch (err: any) {
      if (err.message?.includes('swipe limit')) {
        setSwipesRemaining(0)
      }
    }
    setLoading(false)
  }

  const handleSwipe = async (action: 'like' | 'pass') => {
    const candidate = candidates[currentIndex]
    if (!candidate) return

    try {
      const result = await api.swipe(candidate.user_id, action)
      if (result.is_match) {
        setMatchPopup(candidate)
        window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
      }
    } catch (err) {
      // fail silently
    }

    const nextIndex = currentIndex + 1
    setCurrentIndex(nextIndex)

    if (swipesRemaining !== null) {
      setSwipesRemaining((prev) => (prev !== null ? prev - 1 : null))
    }

    // Load more when running low
    if (nextIndex >= candidates.length - 2) {
      loadExplore(offset + 10)
    }
  }

  if (loading) {
    return <div className="loading">Loading profiles...</div>
  }

  if (swipesRemaining === 0) {
    return (
      <div className="container page">
        <div className="empty-state">
          <h3>Daily swipe limit reached</h3>
          <p>Upgrade to Rooted Plus for unlimited swipes.</p>
          <button className="btn btn-primary" style={{ marginTop: '16px' }}
            onClick={() => window.location.href = '/premium'}>
            Upgrade
          </button>
        </div>
      </div>
    )
  }

  const candidate = candidates[currentIndex]

  return (
    <div className="container page">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h1 className="page-header">Explore</h1>
        {swipesRemaining !== null && user?.subscription === 'free' && (
          <span style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
            {swipesRemaining} swipes left
          </span>
        )}
      </div>

      {!candidate && (
        <div className="empty-state">
          <h3>No more profiles</h3>
          <p>Check back later for new people.</p>
        </div>
      )}

      {candidate && (
        <SwipeCard candidate={candidate} onSwipe={handleSwipe} />
      )}

      {matchPopup && (
        <div style={{
          position: 'fixed', top: 0, left: 0, right: 0, bottom: 0,
          background: 'rgba(0,0,0,0.8)', display: 'flex',
          alignItems: 'center', justifyContent: 'center', zIndex: 1000, padding: '20px',
        }}
          onClick={() => setMatchPopup(null)}
        >
          <div style={{
            background: 'var(--bg)', borderRadius: 'var(--radius)',
            padding: '32px', textAlign: 'center', maxWidth: '320px',
          }}
            onClick={(e) => e.stopPropagation()}>
            <div style={{ fontSize: '48px', marginBottom: '16px' }}>♡</div>
            <h2>It's a match!</h2>
            <p style={{ color: 'var(--text-secondary)', margin: '8px 0 20px' }}>
              You and {matchPopup.first_name} liked each other!
            </p>
            <button className="btn btn-primary" onClick={() => setMatchPopup(null)}>
              Keep exploring
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
