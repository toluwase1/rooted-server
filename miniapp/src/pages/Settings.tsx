import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'

interface Props {
  user: any
  profile: any
  onUserUpdated?: () => void
}

export default function Settings({ user: initialUser, profile, onUserUpdated }: Props) {
  const navigate = useNavigate()
  const [user, setUser] = useState(initialUser)
  const [confirmDelete, setConfirmDelete] = useState(false)

  const handlePause = async () => {
    await api.pauseProfile()
    setUser((prev: any) => ({ ...prev, status: 'paused' }))
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('warning')
    onUserUpdated?.()
  }

  const handleResume = async () => {
    await api.resumeProfile()
    setUser((prev: any) => ({ ...prev, status: 'active' }))
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    onUserUpdated?.()
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

  const photoCount = profile?.photos?.length || 0

  return (
    <div className="container page">
      <h1 className="page-header">Settings</h1>

      {/* Profile preview */}
      {profile && (
        <div className="card" onClick={() => navigate(`/profile/${profile.user_id}`)}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <div style={{
              width: '48px', height: '48px', borderRadius: '50%',
              background: 'linear-gradient(135deg, #E07A5F, #81B29A)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              color: 'white', fontSize: '18px', fontWeight: 'bold',
              overflow: 'hidden',
            }}>
              {profile.photos?.[0] ? (
                <img src={profile.photos[0].url_medium} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
              ) : (
                profile.first_name[0]
              )}
            </div>
            <div style={{ flex: 1 }}>
              <div style={{ fontWeight: '600' }}>{profile.first_name}</div>
              <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
                View your profile
              </div>
            </div>
            <span style={{ color: 'var(--text-secondary)' }}>→</span>
          </div>
        </div>
      )}

      {/* Quick actions */}
      <div className="section-label" style={{ marginTop: '16px' }}>Profile</div>

      <div className="card" onClick={() => navigate('/edit-profile')} style={{ cursor: 'pointer' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <div style={{ fontWeight: '600', fontSize: '14px' }}>Edit profile info</div>
            <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
              Name, city, heritage, faith, bio
            </div>
          </div>
          <span style={{ color: 'var(--text-secondary)' }}>→</span>
        </div>
      </div>

      <div className="card" onClick={() => navigate('/complete-profile')} style={{ cursor: 'pointer' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <div style={{ fontWeight: '600', fontSize: '14px' }}>Edit prompts & photos</div>
            <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
              {photoCount}/6 photos · {(profile?.cultural_prompts?.length || 0)} cultural prompts
            </div>
          </div>
          <span style={{ color: 'var(--text-secondary)' }}>→</span>
        </div>
      </div>

      {/* Completeness */}
      {profile && profile.completeness < 90 && (
        <div className="card" style={{ background: '#81B29A10' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '6px' }}>
            <span style={{ fontSize: '13px', fontWeight: '600' }}>Profile strength</span>
            <span style={{ fontSize: '13px', color: 'var(--accent)', fontWeight: '600' }}>{profile.completeness}%</span>
          </div>
          <div style={{
            width: '100%', height: '6px', borderRadius: '3px',
            background: 'var(--border)', overflow: 'hidden',
          }}>
            <div style={{
              width: `${profile.completeness}%`, height: '100%',
              background: 'var(--accent)', borderRadius: '3px',
            }} />
          </div>
        </div>
      )}

      {/* Subscription */}
      <div className="section-label" style={{ marginTop: '16px' }}>Account</div>

      <div className="card" style={{ opacity: 0.6 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <div style={{ fontWeight: '600', fontSize: '14px' }}>Subscription</div>
            <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              {subLabel[user?.subscription] || 'Free'}
            </div>
          </div>
          <span style={{
            fontSize: '11px', fontWeight: '600', color: 'var(--primary)',
            background: 'var(--primary)12', padding: '4px 10px', borderRadius: '12px',
          }}>Coming Soon</span>
        </div>
      </div>

      {/* Verification */}
      <div className="card" style={{ opacity: 0.6 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <div style={{ fontWeight: '600', fontSize: '14px' }}>Verification</div>
            <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              {user?.verification === 'photo_verified' ? 'Photo verified' : 'Not verified'}
            </div>
          </div>
          <span style={{
            fontSize: '11px', fontWeight: '600', color: 'var(--primary)',
            background: 'var(--primary)12', padding: '4px 10px', borderRadius: '12px',
          }}>Coming Soon</span>
        </div>
      </div>

      {/* Profile controls */}
      <div className="section-label" style={{ marginTop: '16px' }}>Controls</div>

      <div className="card" onClick={user?.status === 'paused' ? handleResume : handlePause}
        style={{ cursor: 'pointer' }}>
        <div style={{ fontWeight: '600', fontSize: '14px' }}>
          {user?.status === 'paused' ? 'Resume Profile' : 'Pause Profile'}
        </div>
        <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
          {user?.status === 'paused' ? 'Make your profile visible again' : 'Hide from searches temporarily'}
        </div>
      </div>

      <div className="card" onClick={handleDelete}
        style={{ cursor: 'pointer', borderColor: confirmDelete ? '#FF3B30' : undefined }}>
        <div style={{ fontWeight: '600', fontSize: '14px', color: '#FF3B30' }}>
          {confirmDelete ? 'Tap again to confirm deletion' : 'Delete Account'}
        </div>
        <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
          {confirmDelete ? 'This is permanent and cannot be undone' : 'Permanently delete your account and data'}
        </div>
      </div>

      {/* Close app */}
      <button className="btn btn-secondary" style={{ marginTop: '16px', width: '100%' }}
        onClick={() => window.Telegram?.WebApp?.close()}>
        Close App
      </button>
    </div>
  )
}
