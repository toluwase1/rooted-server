import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import SwipeCard from '../components/SwipeCard'

interface Props {
  user: any
  profile: any
}

export default function Circle({ user, profile }: Props) {
  const navigate = useNavigate()
  const [candidates, setCandidates] = useState<any[]>([])
  const [currentIndex, setCurrentIndex] = useState(0)
  const [loading, setLoading] = useState(true)
  const [matchPopup, setMatchPopup] = useState<any>(null)

  useEffect(() => {
    loadCircle()
  }, [])

  const loadCircle = async () => {
    try {
      const data = await api.getCircle()
      setCandidates(data.candidates || [])
    } catch (err) {
      // Circle not ready yet
    }
    setLoading(false)
  }

  const handleSwipe = async (action: 'like' | 'pass', comment?: string) => {
    const candidate = candidates[currentIndex]
    if (!candidate) return

    try {
      const result = await api.swipe(candidate.user_id, action, comment)
      if (result.is_match) {
        setMatchPopup(candidate)
        window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
      }
    } catch (err) {
      // Swipe failed
    }

    setCurrentIndex((prev) => prev + 1)
  }

  if (loading) {
    return <div className="loading">Loading your circle...</div>
  }

  const candidate = candidates[currentIndex]

  return (
    <div className="container page">
      <h1 className="page-header">Your Circle</h1>

      {profile && profile.completeness < 100 && (
        <div className="card" onClick={() => navigate('/complete-profile')}
          style={{ cursor: 'pointer', display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
          <div>
            <div style={{ fontWeight: '600', fontSize: '14px' }}>Complete your profile</div>
            <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
              {profile.completeness}% done — add prompts to get 3x more likes
            </div>
          </div>
          <span style={{ color: 'var(--primary)', fontSize: '20px' }}>→</span>
        </div>
      )}

      {!candidate && candidates.length === 0 && (
        <div className="empty-state">
          <span className="empty-state-icon">◎</span>
          <h3>No matches yet</h3>
          <p>Your daily circle will be delivered at 7 PM. Check back later!</p>
        </div>
      )}

      {!candidate && candidates.length > 0 && (
        <div className="empty-state">
          <span className="empty-state-icon">✓</span>
          <h3>You've seen everyone</h3>
          <p>Come back tomorrow for new profiles, or try Explore for more.</p>
        </div>
      )}

      {candidate && (
        <SwipeCard candidate={candidate} onSwipe={handleSwipe} />
      )}

      {candidates.length > 0 && (
        <p style={{ textAlign: 'center', color: 'var(--text-secondary)', fontSize: '13px' }}>
          {currentIndex + 1} of {candidates.length} profiles
        </p>
      )}

      {matchPopup && (
        <div className="match-overlay" onClick={() => setMatchPopup(null)}>
          <div className="match-overlay-card" onClick={(e) => e.stopPropagation()}>
            <span className="empty-state-icon">♡</span>
            <h2 style={{ marginBottom: '8px' }}>It's a match!</h2>
            <p style={{ color: 'var(--text-secondary)', marginBottom: '20px', fontSize: '14px' }}>
              You and {matchPopup.first_name} liked each other. Start chatting!
            </p>
            <button className="btn btn-primary" onClick={() => setMatchPopup(null)}>
              Keep swiping
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
