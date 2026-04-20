import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import LocationPicker from '../components/LocationPicker'

const HERITAGE_OPTIONS: { label: string; value: string }[] = [
  { label: 'Nigerian', value: 'nigerian' }, { label: 'Ghanaian', value: 'ghanaian' },
  { label: 'Kenyan', value: 'kenyan' }, { label: 'Ethiopian', value: 'ethiopian' },
  { label: 'Cameroonian', value: 'cameroonian' }, { label: 'South African', value: 'south_african' },
  { label: 'Tanzanian', value: 'tanzanian' }, { label: 'Ugandan', value: 'ugandan' },
  { label: 'Senegalese', value: 'senegalese' }, { label: 'Congolese', value: 'congolese' },
  { label: 'Zimbabwean', value: 'zimbabwean' }, { label: 'Somali', value: 'somali' },
  { label: 'Eritrean', value: 'eritrean' }, { label: 'Sierra Leonean', value: 'sierra_leonean' },
  { label: 'Liberian', value: 'liberian' }, { label: 'African American', value: 'african_american' },
  { label: 'Caribbean', value: 'caribbean' }, { label: 'British African', value: 'british_african' },
  { label: 'French African', value: 'french_african' }, { label: 'Other', value: 'other' },
]

const FAITH_OPTIONS: { label: string; value: string }[] = [
  { label: 'Christian', value: 'christian' }, { label: 'Muslim', value: 'muslim' },
  { label: 'Traditional', value: 'traditional' }, { label: 'Spiritual', value: 'spiritual' },
  { label: 'Not religious', value: 'not_religious' }, { label: 'Prefer not to say', value: 'prefer_not_to_say' },
]

interface Props {
  profile: any
  onSaved: () => void
}

export default function EditProfile({ profile, onSaved }: Props) {
  const navigate = useNavigate()
  const [form, setForm] = useState({
    first_name: '',
    city: '',
    country: '',
    heritage: [] as string[],
    diaspora_tag: '',
    intention: '',
    faith: '',
    faith_importance: '',
    bio: '',
  })
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (profile) {
      setForm({
        first_name: profile.first_name || '',
        city: profile.city || '',
        country: profile.country || '',
        heritage: profile.heritage || [],
        diaspora_tag: profile.diaspora_tag || '',
        intention: profile.intention || '',
        faith: profile.faith || '',
        faith_importance: profile.faith_importance || '',
        bio: profile.bio || '',
      })
    }

    window.Telegram?.WebApp?.BackButton?.show()
    window.Telegram?.WebApp?.BackButton?.onClick(() => navigate(-1))
    return () => { window.Telegram?.WebApp?.BackButton?.hide() }
  }, [profile])

  const update = (field: string, value: any) => {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  const toggleHeritage = (value: string) => {
    setForm((prev) => ({
      ...prev,
      heritage: prev.heritage.includes(value)
        ? prev.heritage.filter((x) => x !== value)
        : [...prev.heritage, value],
    }))
  }

  const save = async () => {
    setSaving(true)
    try {
      await api.updateProfile(form)
      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
      onSaved()
      navigate('/settings')
    } catch (err: any) {
      alert(err.message || 'Failed to save')
    }
    setSaving(false)
  }

  return (
    <div className="container page">
      <h1 className="page-header">Edit Profile</h1>

      <div className="input-group">
        <label>First name</label>
        <input value={form.first_name} onChange={(e) => update('first_name', e.target.value)} />
      </div>

      <LocationPicker city={form.city} country={form.country}
        onUpdate={(city, country, lat, lon) => {
          update('city', city)
          update('country', country)
          if (lat && lon) {
            setForm(prev => ({ ...prev, latitude: lat, longitude: lon } as any))
          }
        }} />

      <div className="input-group">
        <label>Heritage</label>
        <div className="tag-selector">
          {HERITAGE_OPTIONS.map((h) => (
            <div key={h.value} className={`tag ${form.heritage.includes(h.value) ? 'selected' : ''}`}
              onClick={() => toggleHeritage(h.value)}>
              {h.label}
            </div>
          ))}
        </div>
      </div>

      <div className="input-group">
        <label>Diaspora status</label>
        {['born_in_africa', 'diaspora_1st', 'diaspora_2nd', 'returnee', 'explorer'].map((tag) => (
          <div key={tag} className={`prompt-option ${form.diaspora_tag === tag ? 'selected' : ''}`}
            onClick={() => update('diaspora_tag', tag)}>
            {{ born_in_africa: 'Born in Africa', diaspora_1st: '1st Generation Diaspora',
              diaspora_2nd: '2nd Generation+ Diaspora', returnee: 'Returnee',
              explorer: 'Explorer (non-African)' }[tag]}
          </div>
        ))}
      </div>

      <div className="input-group">
        <label>Looking for</label>
        <div style={{ display: 'flex', gap: '8px' }}>
          {['dating', 'friendship', 'both'].map((i) => (
            <button key={i} className={`btn ${form.intention === i ? 'btn-primary' : 'btn-secondary'}`}
              style={{ flex: 1 }} onClick={() => update('intention', i)}>
              {{ dating: 'Dating', friendship: 'Friends', both: 'Both' }[i]}
            </button>
          ))}
        </div>
      </div>

      <div className="input-group">
        <label>Faith</label>
        <div className="tag-selector">
          {FAITH_OPTIONS.map((f) => (
            <div key={f.value} className={`tag ${form.faith === f.value ? 'selected' : ''}`}
              onClick={() => update('faith', form.faith === f.value ? '' : f.value)}>
              {f.label}
            </div>
          ))}
        </div>
      </div>

      {form.faith && form.faith !== 'prefer_not_to_say' && (
        <div className="input-group">
          <label>Faith importance</label>
          <div style={{ display: 'flex', gap: '8px' }}>
            {[
              { label: 'Very', value: 'very_important' },
              { label: 'Somewhat', value: 'somewhat' },
              { label: 'Not much', value: 'not_important' },
            ].map((fi) => (
              <button key={fi.value}
                className={`btn ${form.faith_importance === fi.value ? 'btn-primary' : 'btn-secondary'}`}
                style={{ flex: 1 }} onClick={() => update('faith_importance', fi.value)}>
                {fi.label}
              </button>
            ))}
          </div>
        </div>
      )}

      <div className="input-group">
        <label>Bio</label>
        <textarea value={form.bio} onChange={(e) => update('bio', e.target.value)}
          placeholder="A few words about you..." maxLength={150} />
        <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>{form.bio.length}/150</span>
      </div>

      <button className="btn btn-primary" style={{ marginTop: '16px' }}
        disabled={saving || !form.first_name}
        onClick={save}>
        {saving ? 'Saving...' : 'Save changes'}
      </button>
    </div>
  )
}
