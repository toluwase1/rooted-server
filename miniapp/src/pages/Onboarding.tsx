import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import PhotoUpload from '../components/PhotoUpload'
import LocationPicker from '../components/LocationPicker'

const HERITAGE_OPTIONS: { label: string; value: string }[] = [
  { label: 'Nigerian', value: 'nigerian' },
  { label: 'Ghanaian', value: 'ghanaian' },
  { label: 'Kenyan', value: 'kenyan' },
  { label: 'Ethiopian', value: 'ethiopian' },
  { label: 'Cameroonian', value: 'cameroonian' },
  { label: 'South African', value: 'south_african' },
  { label: 'Tanzanian', value: 'tanzanian' },
  { label: 'Ugandan', value: 'ugandan' },
  { label: 'Senegalese', value: 'senegalese' },
  { label: 'Congolese', value: 'congolese' },
  { label: 'Zimbabwean', value: 'zimbabwean' },
  { label: 'Somali', value: 'somali' },
  { label: 'Eritrean', value: 'eritrean' },
  { label: 'Sierra Leonean', value: 'sierra_leonean' },
  { label: 'Liberian', value: 'liberian' },
  { label: 'African American', value: 'african_american' },
  { label: 'Caribbean', value: 'caribbean' },
  { label: 'British African', value: 'british_african' },
  { label: 'French African', value: 'french_african' },
  { label: 'Other', value: 'other' },
]

const FAITH_OPTIONS: { label: string; value: string }[] = [
  { label: 'Christian', value: 'christian' },
  { label: 'Muslim', value: 'muslim' },
  { label: 'Traditional', value: 'traditional' },
  { label: 'Spiritual', value: 'spiritual' },
  { label: 'Not religious', value: 'not_religious' },
  { label: 'Prefer not to say', value: 'prefer_not_to_say' },
]

interface Props {
  onComplete: (profile: any) => void
}

export default function Onboarding({ onComplete }: Props) {
  const navigate = useNavigate()
  const [step, setStep] = useState(0)
  const [consentChecked, setConsentChecked] = useState(false)
  const [form, setForm] = useState({
    first_name: '',
    date_of_birth: '',
    gender: '',
    gender_pref: '',
    city: '',
    country: '',
    latitude: 0,
    longitude: 0,
    heritage: [] as string[],
    diaspora_tag: '',
    intention: '',
    faith: '',
    faith_importance: '',
    bio: '',
    cultural_prompts: [] as { prompt: string; answer: string }[],
    personality_prompts: [] as { prompt: string; answer: string }[],
  })
  const [photos, setPhotos] = useState<{ id: string; url: string; file?: File }[]>([])
  const [uploadingPhoto, setUploadingPhoto] = useState(false)
  const [submitting, setSubmitting] = useState(false)

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


  const handleAddPhoto = (file: File) => {
    setUploadingPhoto(true)
    const localURL = URL.createObjectURL(file)
    const tempId = `temp-${Date.now()}`
    setPhotos((prev) => [...prev, { id: tempId, url: localURL, file }])
    setUploadingPhoto(false)
  }

  const handleRemovePhoto = (id: string) => {
    setPhotos((prev) => prev.filter((p) => p.id !== id))
  }

  const submit = async () => {
    setSubmitting(true)
    try {
      const profile = await api.createProfile({
        ...form,
        terms_accepted: true,
        privacy_accepted: true,
      })

      for (const photo of photos) {
        if (photo.file) {
          try { await api.uploadPhoto(photo.file) } catch { /* continue */ }
        }
      }

      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
      onComplete(profile)
    } catch (err: any) {
      alert(err.message || 'Failed to create profile. Please try again.')
    }
    setSubmitting(false)
  }

  const steps = [
    // Step 1: Basics + Photo
    <div key="basics" className="container page">
      <h1 className="page-header">Let's get started</h1>

      <div className="input-group">
        <label>First name</label>
        <input value={form.first_name} onChange={(e) => update('first_name', e.target.value)}
          placeholder="What should we call you?" />
      </div>

      <div className="input-group">
        <label>Date of birth (must be 18+)</label>
        <input type="date" value={form.date_of_birth}
          max={new Date(Date.now() - 18 * 365.25 * 24 * 60 * 60 * 1000).toISOString().split('T')[0]}
          min="1924-01-01"
          onChange={(e) => update('date_of_birth', e.target.value)} />
      </div>

      <div className="input-group">
        <label>I am</label>
        <div style={{ display: 'flex', gap: '8px' }}>
          {['male', 'female'].map((g) => (
            <button key={g} className={`btn ${form.gender === g ? 'btn-primary' : 'btn-secondary'}`}
              style={{ flex: 1 }} onClick={() => update('gender', g)}>
              {g === 'male' ? 'Male' : 'Female'}
            </button>
          ))}
        </div>
      </div>

      <div className="input-group">
        <label>Interested in</label>
        <div style={{ display: 'flex', gap: '8px' }}>
          {['male', 'female', 'everyone'].map((g) => (
            <button key={g} className={`btn ${form.gender_pref === g ? 'btn-primary' : 'btn-secondary'}`}
              style={{ flex: 1 }} onClick={() => update('gender_pref', g)}>
              {g === 'everyone' ? 'Everyone' : g === 'male' ? 'Men' : 'Women'}
            </button>
          ))}
        </div>
      </div>

      <div className="input-group" style={{ marginTop: '8px' }}>
        <label>Add a photo</label>
        <PhotoUpload
          photos={photos}
          onAdd={handleAddPhoto}
          onRemove={handleRemovePhoto}
          uploading={uploadingPhoto}
          maxPhotos={1}
        />
      </div>

      <button className="btn btn-primary" style={{ marginTop: '16px' }}
        disabled={!form.first_name || !form.date_of_birth || !form.gender || !form.gender_pref}
        onClick={() => setStep(1)}>
        Continue
      </button>
    </div>,

    // Step 2: Heritage + Diaspora + Intention
    <div key="identity" className="container page">
      <h1 className="page-header">Your roots</h1>

      <div className="input-group">
        <label>Heritage (select all that apply)</label>
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
        <label>I'm here for</label>
        <div style={{ display: 'flex', gap: '8px' }}>
          {['dating', 'friendship', 'both'].map((i) => (
            <button key={i} className={`btn ${form.intention === i ? 'btn-primary' : 'btn-secondary'}`}
              style={{ flex: 1 }} onClick={() => update('intention', i)}>
              {{ dating: 'Dating', friendship: 'Friends', both: 'Both' }[i]}
            </button>
          ))}
        </div>
      </div>

      <button className="btn btn-primary" style={{ marginTop: '16px' }}
        disabled={form.heritage.length === 0 || !form.diaspora_tag || !form.intention}
        onClick={() => setStep(2)}>
        Continue
      </button>
    </div>,

    // Step 3: Location + Faith + Submit
    <div key="details" className="container page">
      <h1 className="page-header">Almost done</h1>

      <LocationPicker city={form.city} country={form.country}
        onUpdate={(city, country, lat, lon) => {
          update('city', city)
          update('country', country)
          if (lat && lon) { update('latitude', lat); update('longitude', lon) }
        }} />

      <div className="input-group">
        <label>Faith (optional)</label>
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
          <label>How important is faith to you?</label>
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

      <div className="input-group" style={{ marginTop: '4px' }}>
        <label>Short bio (optional)</label>
        <textarea value={form.bio} onChange={(e) => update('bio', e.target.value)}
          placeholder="A few words about you..." maxLength={150}
          style={{ minHeight: '60px' }} />
        <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
          {form.bio.length}/150
        </span>
      </div>

      <label
        style={{
          display: 'flex',
          alignItems: 'flex-start',
          gap: '10px',
          marginTop: '20px',
          cursor: 'pointer',
          fontSize: '13px',
          lineHeight: '1.5',
          color: 'var(--text-primary)',
        }}
      >
        <input
          type="checkbox"
          checked={consentChecked}
          onChange={(e) => setConsentChecked(e.target.checked)}
          style={{
            marginTop: '3px',
            width: '20px',
            height: '20px',
            minWidth: '20px',
            accentColor: 'var(--primary, #8B5CF6)',
          }}
        />
        <span>
          I agree to the{' '}
          <span
            style={{ color: 'var(--primary, #8B5CF6)', textDecoration: 'underline', fontWeight: 500 }}
            onClick={(e) => { e.preventDefault(); e.stopPropagation(); window.Telegram?.WebApp?.openLink(window.location.origin + '/terms') }}
          >
            Terms of Service
          </span>
          ,{' '}
          <span
            style={{ color: 'var(--primary, #8B5CF6)', textDecoration: 'underline', fontWeight: 500 }}
            onClick={(e) => { e.preventDefault(); e.stopPropagation(); window.Telegram?.WebApp?.openLink(window.location.origin + '/privacy') }}
          >
            Privacy Policy
          </span>
          , and{' '}
          <span
            style={{ color: 'var(--primary, #8B5CF6)', textDecoration: 'underline', fontWeight: 500 }}
            onClick={(e) => { e.preventDefault(); e.stopPropagation(); window.Telegram?.WebApp?.openLink(window.location.origin + '/guidelines') }}
          >
            Community Guidelines
          </span>
        </span>
      </label>

      <button className="btn btn-primary" style={{ marginTop: '16px' }}
        disabled={submitting || !consentChecked}
        onClick={submit}>
        {submitting ? 'Creating profile...' : 'Start matching'}
      </button>

      <p style={{ textAlign: 'center', fontSize: '12px', color: 'var(--text-secondary)', marginTop: '12px' }}>
        You can add cultural prompts and more photos later to improve your matches.
      </p>
    </div>,
  ]

  return (
    <div>
      <div className="progress-bar">
        {steps.map((_, i) => (
          <div key={i} className={`progress-segment ${i <= step ? 'active' : ''}`} />
        ))}
      </div>

      {steps[step]}
    </div>
  )
}
