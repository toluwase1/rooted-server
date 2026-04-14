import { useState } from 'react'
import { api } from '../api/client'
import PhotoUpload from '../components/PhotoUpload'

const CULTURAL_PROMPTS = [
  'My culture means...',
  'A tradition I love is...',
  'Home is...',
  'My family would describe me as...',
  "I'm proudly...",
  'The food I\'d cook for you is...',
  'My Sunday looks like...',
  'I switch between ___ and ___ cultures when...',
  'The song that defines me is...',
  'What I miss most about home is...',
]

const PERSONALITY_PROMPTS = [
  'My ideal weekend is...',
  "I'm looking for someone who...",
  'The way to my heart is...',
  'I geek out about...',
  'My friends would say I\'m...',
  'A perfect first date is...',
  "The hill I'll die on is...",
  'I recently discovered...',
]

const HERITAGE_OPTIONS = [
  'Nigerian', 'Ghanaian', 'Kenyan', 'Ethiopian', 'Cameroonian',
  'South African', 'Tanzanian', 'Ugandan', 'Senegalese', 'Congolese',
  'Zimbabwean', 'Somali', 'Eritrean', 'Sierra Leonean', 'Liberian',
  'African American', 'Caribbean', 'British African', 'French African', 'Other',
]

const FAITH_OPTIONS = [
  'Christian', 'Muslim', 'Traditional', 'Spiritual', 'Not religious', 'Prefer not to say',
]

interface Props {
  onComplete: (profile: any) => void
}

export default function Onboarding({ onComplete }: Props) {
  const [step, setStep] = useState(0)
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
  const [selectedCulturalPrompts, setSelectedCulturalPrompts] = useState<string[]>([])
  const [selectedPersonalityPrompt, setSelectedPersonalityPrompt] = useState('')
  const [photos, setPhotos] = useState<{ id: string; url: string; file?: File }[]>([])
  const [uploadingPhoto, setUploadingPhoto] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const update = (field: string, value: any) => {
    setForm((prev) => ({ ...prev, [field]: value }))
  }

  const toggleHeritage = (h: string) => {
    const lower = h.toLowerCase()
    setForm((prev) => ({
      ...prev,
      heritage: prev.heritage.includes(lower)
        ? prev.heritage.filter((x) => x !== lower)
        : [...prev.heritage, lower],
    }))
  }

  const requestLocation = () => {
    if (navigator.geolocation) {
      navigator.geolocation.getCurrentPosition(
        (pos) => {
          update('latitude', pos.coords.latitude)
          update('longitude', pos.coords.longitude)
        },
        () => {} // silently fail — location is optional
      )
    }
  }

  const handleAddPhoto = async (file: File) => {
    setUploadingPhoto(true)
    try {
      // Create a local preview immediately
      const localURL = URL.createObjectURL(file)
      const tempId = `temp-${Date.now()}`
      setPhotos((prev) => [...prev, { id: tempId, url: localURL, file }])
    } catch (err) {
      alert('Failed to add photo')
    }
    setUploadingPhoto(false)
  }

  const handleRemovePhoto = (id: string) => {
    setPhotos((prev) => prev.filter((p) => p.id !== id))
  }

  const submit = async () => {
    setSubmitting(true)
    try {
      // 1. Create profile first
      const profile = await api.createProfile(form)

      // 2. Upload photos
      for (const photo of photos) {
        if (photo.file) {
          try {
            await api.uploadPhoto(photo.file)
          } catch {
            // Continue even if one photo fails
          }
        }
      }

      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
      onComplete(profile)
    } catch (err) {
      alert('Failed to create profile. Please try again.')
    }
    setSubmitting(false)
  }

  const steps = [
    // Step 0: Basics
    <div key="basics" className="container page">
      <h1 className="page-header">Let's get started</h1>

      <div className="input-group">
        <label>First name</label>
        <input value={form.first_name} onChange={(e) => update('first_name', e.target.value)}
          placeholder="What should we call you?" />
      </div>

      <div className="input-group">
        <label>Date of birth</label>
        <input type="date" value={form.date_of_birth}
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

      <button className="btn btn-primary" style={{ marginTop: '20px' }}
        disabled={!form.first_name || !form.date_of_birth || !form.gender || !form.gender_pref}
        onClick={() => { requestLocation(); setStep(1) }}>
        Continue
      </button>
    </div>,

    // Step 1: Location & Heritage
    <div key="heritage" className="container page">
      <h1 className="page-header">Where are you from?</h1>

      <div className="input-group">
        <label>City</label>
        <input value={form.city} onChange={(e) => update('city', e.target.value)}
          placeholder="Lagos, London, New York..." />
      </div>

      <div className="input-group">
        <label>Country</label>
        <input value={form.country} onChange={(e) => update('country', e.target.value.toUpperCase())}
          placeholder="NG, GB, US..." maxLength={3} />
      </div>

      <div className="input-group">
        <label>Heritage (select all that apply)</label>
        <div className="tag-selector">
          {HERITAGE_OPTIONS.map((h) => (
            <div key={h} className={`tag ${form.heritage.includes(h.toLowerCase()) ? 'selected' : ''}`}
              onClick={() => toggleHeritage(h)}>
              {h}
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

      <button className="btn btn-primary" style={{ marginTop: '20px' }}
        disabled={form.heritage.length === 0 || !form.diaspora_tag}
        onClick={() => setStep(2)}>
        Continue
      </button>
    </div>,

    // Step 2: Intention & Faith
    <div key="intention" className="container page">
      <h1 className="page-header">What are you looking for?</h1>

      <div className="input-group">
        <label>I'm here for</label>
        {['dating', 'friendship', 'both'].map((i) => (
          <div key={i} className={`prompt-option ${form.intention === i ? 'selected' : ''}`}
            onClick={() => update('intention', i)}>
            {{ dating: 'Dating', friendship: 'Friendship', both: 'Both dating & friendship' }[i]}
          </div>
        ))}
      </div>

      <div className="input-group">
        <label>Faith</label>
        <div className="tag-selector">
          {FAITH_OPTIONS.map((f) => (
            <div key={f} className={`tag ${form.faith === f.toLowerCase() ? 'selected' : ''}`}
              onClick={() => update('faith', f.toLowerCase())}>
              {f}
            </div>
          ))}
        </div>
      </div>

      {form.faith && (
        <div className="input-group">
          <label>How important is faith to you?</label>
          {['very_important', 'somewhat', 'not_important'].map((fi) => (
            <div key={fi} className={`prompt-option ${form.faith_importance === fi ? 'selected' : ''}`}
              onClick={() => update('faith_importance', fi)}>
              {{ very_important: 'Very important', somewhat: 'Somewhat important',
                not_important: 'Not important' }[fi]}
            </div>
          ))}
        </div>
      )}

      <button className="btn btn-primary" style={{ marginTop: '20px' }}
        disabled={!form.intention}
        onClick={() => setStep(3)}>
        Continue
      </button>
    </div>,

    // Step 3: Cultural Prompts
    <div key="cultural" className="container page">
      <h1 className="page-header">Show your culture</h1>
      <p style={{ color: 'var(--text-secondary)', marginBottom: '16px' }}>
        Pick 2 prompts and write your answers.
      </p>

      {CULTURAL_PROMPTS.map((prompt) => {
        const isSelected = selectedCulturalPrompts.includes(prompt)
        const existing = form.cultural_prompts.find((p) => p.prompt === prompt)

        return (
          <div key={prompt}>
            <div className={`prompt-option ${isSelected ? 'selected' : ''}`}
              onClick={() => {
                if (isSelected) {
                  setSelectedCulturalPrompts((prev) => prev.filter((p) => p !== prompt))
                  update('cultural_prompts', form.cultural_prompts.filter((p) => p.prompt !== prompt))
                } else if (selectedCulturalPrompts.length < 2) {
                  setSelectedCulturalPrompts((prev) => [...prev, prompt])
                }
              }}>
              {prompt}
            </div>
            {isSelected && (
              <div className="input-group" style={{ marginTop: '4px' }}>
                <textarea value={existing?.answer || ''} placeholder="Your answer..."
                  onChange={(e) => {
                    const updated = form.cultural_prompts.filter((p) => p.prompt !== prompt)
                    updated.push({ prompt, answer: e.target.value })
                    update('cultural_prompts', updated)
                  }} />
              </div>
            )}
          </div>
        )
      })}

      <button className="btn btn-primary" style={{ marginTop: '20px' }}
        disabled={form.cultural_prompts.filter((p) => p.answer).length < 2}
        onClick={() => setStep(4)}>
        Continue
      </button>
    </div>,

    // Step 4: Personality Prompt + Bio
    <div key="personality" className="container page">
      <h1 className="page-header">Almost there</h1>

      <p style={{ color: 'var(--text-secondary)', marginBottom: '16px' }}>
        Pick a prompt that shows your personality.
      </p>

      {PERSONALITY_PROMPTS.map((prompt) => (
        <div key={prompt}>
          <div className={`prompt-option ${selectedPersonalityPrompt === prompt ? 'selected' : ''}`}
            onClick={() => setSelectedPersonalityPrompt(prompt)}>
            {prompt}
          </div>
          {selectedPersonalityPrompt === prompt && (
            <div className="input-group" style={{ marginTop: '4px' }}>
              <textarea placeholder="Your answer..."
                onChange={(e) => {
                  update('personality_prompts', [{ prompt, answer: e.target.value }])
                }} />
            </div>
          )}
        </div>
      ))}

      <div className="input-group" style={{ marginTop: '20px' }}>
        <label>Short bio (optional)</label>
        <textarea value={form.bio} onChange={(e) => update('bio', e.target.value)}
          placeholder="A few words about you..." maxLength={150} />
        <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
          {form.bio.length}/150
        </span>
      </div>

      <button className="btn btn-primary" style={{ marginTop: '20px' }}
        disabled={form.personality_prompts.length === 0 || !form.personality_prompts[0]?.answer}
        onClick={() => setStep(5)}>
        Continue
      </button>
    </div>,

    // Step 5: Photos
    <div key="photos" className="container page">
      <h1 className="page-header">Add your photos</h1>
      <p style={{ color: 'var(--text-secondary)', marginBottom: '16px' }}>
        Add at least 1 photo. The first photo is your primary — make it count!
      </p>

      <PhotoUpload
        photos={photos}
        onAdd={handleAddPhoto}
        onRemove={handleRemovePhoto}
        uploading={uploadingPhoto}
      />

      <button className="btn btn-primary" style={{ marginTop: '24px' }}
        disabled={photos.length === 0 || submitting}
        onClick={submit}>
        {submitting ? 'Creating profile...' : 'Create Profile'}
      </button>

      <button className="btn btn-secondary" style={{ marginTop: '8px' }}
        disabled={submitting}
        onClick={submit}>
        Skip for now
      </button>
    </div>,
  ]

  return (
    <div>
      {/* Progress bar */}
      <div style={{ display: 'flex', gap: '4px', padding: '12px 16px' }}>
        {steps.map((_, i) => (
          <div key={i} style={{
            flex: 1, height: '3px', borderRadius: '2px',
            background: i <= step ? 'var(--primary)' : 'var(--border)',
            transition: 'background 0.3s',
          }} />
        ))}
      </div>

      {steps[step]}
    </div>
  )
}
