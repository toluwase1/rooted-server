import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'

interface Props {
  user: any
}

export default function Matches({ user }: Props) {
  const [matches, setMatches] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()

  useEffect(() => {
    api.getMatches().then((data) => {
      setMatches(data.matches || [])
      setLoading(false)
    }).catch(() => setLoading(false))
  }, [])

  if (loading) {
    return <div className="loading">Loading matches...</div>
  }

  return (
    <div className="container page">
      <h1 className="page-header">Matches</h1>

      {matches.length === 0 && (
        <div className="empty-state">
          <span className="empty-state-icon">♡</span>
          <h3>No matches yet</h3>
          <p>Keep swiping in your Circle or Explore to find your match.</p>
        </div>
      )}

      {matches.map(({ match, profile }) => {
        if (!profile) return null

        return (
          <div key={match.id} className="match-item card">
            <div className="match-avatar" style={{
              width: '56px', height: '56px', borderRadius: '50%', overflow: 'hidden',
              background: 'linear-gradient(135deg, #E07A5F, #81B29A)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: 'white', fontSize: '20px', fontWeight: 'bold', flexShrink: 0,
            }}>
              {profile.photos?.[0] ? (
                <img src={profile.photos[0].url_medium} alt="" style={{
                  width: '100%', height: '100%', objectFit: 'cover',
                }} />
              ) : (
                profile.first_name[0]
              )}
            </div>

            <div className="match-info">
              <div className="match-name">
                {profile.first_name}
                {profile.verification === 'photo_verified' && ' ✓'}
              </div>
              <div className="match-preview">
                {profile.heritage?.join(', ')} · {profile.city}
              </div>
              {match.last_message_at && (
                <div style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '2px' }}>
                  Last message: {timeAgo(match.last_message_at)}
                </div>
              )}
            </div>

            <div style={{ display: 'flex', gap: '8px', flexShrink: 0 }}>
              <button className="btn btn-secondary" style={{ width: 'auto', padding: '6px 12px', fontSize: '12px' }}
                onClick={(e) => { e.stopPropagation(); navigate(`/profile/${profile.user_id}`) }}>
                Profile
              </button>
              <button className="btn btn-primary" style={{ width: 'auto', padding: '6px 12px', fontSize: '12px' }}
                onClick={(e) => {
                  e.stopPropagation()
                  window.Telegram?.WebApp?.close()
                  window.open('https://t.me/RootedDatingBot', '_blank')
                }}>
                Chat
              </button>
            </div>
          </div>
        )
      })}

      {matches.length > 0 && (
        <p style={{
          textAlign: 'center', color: 'var(--text-secondary)',
          fontSize: '13px', marginTop: '20px',
        }}>
          Tap "Chat" to message your match in the Rooted bot.
          Your messages are private — no phone numbers shared.
        </p>
      )}
    </div>
  )
}

function timeAgo(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)

  if (diffMins < 1) return 'just now'
  if (diffMins < 60) return `${diffMins}m ago`
  const diffHrs = Math.floor(diffMins / 60)
  if (diffHrs < 24) return `${diffHrs}h ago`
  const diffDays = Math.floor(diffHrs / 24)
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
}
