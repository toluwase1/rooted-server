import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'

interface Props {
  user: any
  profile: any
}

export default function Settings({ user, profile }: Props) {
  const navigate = useNavigate()
  const [confirmDelete, setConfirmDelete] = useState(false)

  const handlePause = async () => {
    await api.pauseProfile()
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('warning')
    alert('Profile paused. You won\'t appear in searches.')
  }

  const handleResume = async () => {
    await api.resumeProfile()
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    alert('Profile resumed!')
  }

  const handleDelete = async () => {
    if (!confirmDelete) {
      setConfirmDelete(true)
      return
    }
    await api.deleteAccount()
    window.Telegram?.WebApp?.close()
  }

  const subLabel: Record<string, string> = {
    free: 'Free',
    plus: 'Rooted Plus',
    premium: 'Rooted Premium',
  }

  return (
    <div className="container page">
      <h1 className="page-header">Settings</h1>

      {/* Complete profile banner */}
      {profile && profile.completeness < 100 && (
        <div className="card" onClick={() => navigate('/complete-profile')}
          style={{ cursor: 'pointer', background: '#81B29A15', borderLeft: '3px solid var(--accent)' }}>
          <div style={{ fontWeight: '600', fontSize: '14px' }}>Boost your profile</div>
          <div style={{ fontSize: '13px', color: 'var(--text-secondary)', marginTop: '4px' }}>
            {profile.completeness}% complete — add prompts and photos to get more matches
          </div>
        </div>
      )}

      {/* Profile preview */}
      {profile && (
        <div className="card" onClick={() => navigate(`/profile/${profile.user_id}`)}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <div style={{
              width: '48px', height: '48px', borderRadius: '50%',
              background: 'linear-gradient(135deg, #E07A5F, #81B29A)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: 'white', fontSize: '18px', fontWeight: 'bold',
            }}>
              {profile.first_name[0]}
            </div>
            <div>
              <div style={{ fontWeight: '600' }}>{profile.first_name}</div>
              <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                View your profile
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Subscription */}
      <div className="card">
        <label style={{ fontSize: '12px', color: 'var(--text-secondary)', textTransform: 'uppercase' }}>
          Subscription
        </label>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '8px' }}>
          <span style={{ fontWeight: '600' }}>{subLabel[user?.subscription] || 'Free'}</span>
          {user?.subscription === 'free' && (
            <button className="btn btn-primary" style={{ width: 'auto', padding: '8px 16px', fontSize: '14px' }}
              onClick={() => navigate('/premium')}>
              Upgrade
            </button>
          )}
        </div>
      </div>

      {/* Profile completeness */}
      {profile && (
        <div className="card">
          <label style={{ fontSize: '12px', color: 'var(--text-secondary)', textTransform: 'uppercase' }}>
            Profile completeness
          </label>
          <div style={{ marginTop: '8px' }}>
            <div style={{
              width: '100%', height: '8px', borderRadius: '4px',
              background: 'var(--border)', overflow: 'hidden',
            }}>
              <div style={{
                width: `${profile.completeness}%`, height: '100%',
                background: profile.completeness >= 80 ? 'var(--accent)' : 'var(--primary)',
                borderRadius: '4px', transition: 'width 0.3s',
              }} />
            </div>
            <span style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              {profile.completeness}% complete
            </span>
          </div>
        </div>
      )}

      {/* Verification */}
      <div className="card">
        <label style={{ fontSize: '12px', color: 'var(--text-secondary)', textTransform: 'uppercase' }}>
          Verification
        </label>
        <div style={{ marginTop: '8px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span>
            {user?.verification === 'photo_verified' ? '✓ Photo verified' : 'Not verified'}
          </span>
          {user?.verification !== 'photo_verified' && (
            <button className="btn btn-outline" style={{ width: 'auto', padding: '8px 16px', fontSize: '14px' }}
              onClick={() => navigate('/verify')}>
              Verify now
            </button>
          )}
        </div>
      </div>

      {/* Actions */}
      <div style={{ marginTop: '24px' }}>
        <button className="btn btn-secondary" style={{ marginBottom: '8px' }}
          onClick={user?.status === 'paused' ? handleResume : handlePause}>
          {user?.status === 'paused' ? 'Resume Profile' : 'Pause Profile'}
        </button>

        <button className="btn" style={{
          background: confirmDelete ? '#FF3B30' : 'transparent',
          color: confirmDelete ? 'white' : '#FF3B30',
          border: '1px solid #FF3B30',
        }}
          onClick={handleDelete}>
          {confirmDelete ? 'Confirm deletion — this is permanent' : 'Delete Account'}
        </button>
      </div>
    </div>
  )
}
