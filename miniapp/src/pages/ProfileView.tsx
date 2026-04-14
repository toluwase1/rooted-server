import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../api/client'

interface Props {
  currentUserId?: string
}

export default function ProfileView({ currentUserId }: Props) {
  const { id } = useParams()
  const navigate = useNavigate()
  const [profile, setProfile] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    window.Telegram?.WebApp?.BackButton?.show()
    window.Telegram?.WebApp?.BackButton?.onClick(() => navigate(-1))

    if (id) {
      api.getProfile(id).then(setProfile).catch(() => {}).finally(() => setLoading(false))
    }

    return () => {
      window.Telegram?.WebApp?.BackButton?.hide()
    }
  }, [id])

  if (loading) return <div className="loading">Loading profile...</div>
  if (!profile) return <div className="loading">Profile not found</div>

  const isOwnProfile = profile.user_id === currentUserId

  const diasporaLabel: Record<string, string> = {
    born_in_africa: 'Born in Africa',
    diaspora_1st: '1st Gen Diaspora',
    diaspora_2nd: '2nd Gen+ Diaspora',
    returnee: 'Returnee',
    explorer: 'Explorer',
  }

  return (
    <div className="container page">
      {/* Photo */}
      <div style={{
        width: '100%', aspectRatio: '3/4', borderRadius: 'var(--radius)',
        overflow: 'hidden', marginBottom: '16px', background: 'var(--card-bg)',
      }}>
        {profile.photos?.[0] ? (
          <img src={profile.photos[0].url_large} alt="" style={{
            width: '100%', height: '100%', objectFit: 'cover',
          }} />
        ) : (
          <div style={{
            width: '100%', height: '100%', display: 'flex',
            alignItems: 'center', justifyContent: 'center',
            fontSize: '80px', background: 'linear-gradient(135deg, #E07A5F, #81B29A)',
            color: 'white',
          }}>
            {profile.first_name[0]}
          </div>
        )}
      </div>

      {/* Name & basics */}
      <h1 style={{ fontSize: '28px', marginBottom: '4px' }}>
        {profile.first_name}
        {profile.verification === 'photo_verified' && (
          <span style={{ color: 'var(--accent)', marginLeft: '8px' }}>✓</span>
        )}
      </h1>
      <p style={{ color: 'var(--text-secondary)', marginBottom: '16px' }}>
        {profile.city}, {profile.country}
      </p>

      {/* Badges */}
      <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap', marginBottom: '20px' }}>
        {profile.heritage?.map((h: string) => (
          <span key={h} className="badge badge-heritage">{h}</span>
        ))}
        <span className="badge badge-diaspora">
          {diasporaLabel[profile.diaspora_tag] || profile.diaspora_tag}
        </span>
        {profile.intention && (
          <span className="badge" style={{ background: '#3D405B15', color: 'var(--secondary)' }}>
            {profile.intention}
          </span>
        )}
      </div>

      {/* Bio */}
      {profile.bio && (
        <div className="card">
          <p>{profile.bio}</p>
        </div>
      )}

      {/* Faith */}
      {profile.faith && (
        <div className="card">
          <label style={{ fontSize: '12px', color: 'var(--text-secondary)', textTransform: 'uppercase' }}>
            Faith
          </label>
          <p style={{ marginTop: '4px' }}>
            {profile.faith}
            {profile.faith_importance === 'very_important' && ' (very important to me)'}
            {profile.faith_importance === 'somewhat' && ' (somewhat important)'}
          </p>
        </div>
      )}

      {/* Cultural prompts */}
      {profile.cultural_prompts?.map((p: any, i: number) => (
        <div key={i} className="card">
          <label style={{ fontSize: '12px', color: 'var(--text-secondary)', textTransform: 'uppercase' }}>
            {p.prompt}
          </label>
          <p style={{ marginTop: '4px', fontSize: '16px' }}>{p.answer}</p>
        </div>
      ))}

      {/* Personality prompts */}
      {profile.personality_prompts?.map((p: any, i: number) => (
        <div key={i} className="card">
          <label style={{ fontSize: '12px', color: 'var(--text-secondary)', textTransform: 'uppercase' }}>
            {p.prompt}
          </label>
          <p style={{ marginTop: '4px', fontSize: '16px' }}>{p.answer}</p>
        </div>
      ))}

      {/* Additional photos */}
      {profile.photos?.length > 1 && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px', marginTop: '8px' }}>
          {profile.photos.slice(1).map((photo: any) => (
            <div key={photo.id} style={{
              aspectRatio: '1', borderRadius: 'var(--radius-sm)', overflow: 'hidden',
            }}>
              <img src={photo.url_medium} alt="" style={{
                width: '100%', height: '100%', objectFit: 'cover',
              }} />
            </div>
          ))}
        </div>
      )}

      {/* Actions */}
      {!isOwnProfile && (
        <div style={{ marginTop: '20px' }}>
          <button className="btn btn-outline" style={{ marginBottom: '8px' }}
            onClick={() => {
              api.report(profile.user_id, 'inappropriate', '')
              navigate(-1)
            }}>
            Report profile
          </button>
        </div>
      )}

      {isOwnProfile && (
        <button className="btn btn-secondary" style={{ marginTop: '20px' }}
          onClick={() => navigate('/settings')}>
          Edit profile
        </button>
      )}
    </div>
  )
}
