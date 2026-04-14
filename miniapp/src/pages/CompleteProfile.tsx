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

interface Props {
  profile: any
  onDone: () => void
}

export default function CompleteProfile({ profile, onDone }: Props) {
  const [selectedSection, setSelectedSection] = useState<string | null>(null)
  const [culturalPrompts, setCulturalPrompts] = useState<string[]>([])
  const [culturalAnswers, setCulturalAnswers] = useState<{ prompt: string; answer: string }[]>(
    profile?.cultural_prompts || []
  )
  const [personalityPrompt, setPersonalityPrompt] = useState('')
  const [personalityAnswer, setPersonalityAnswer] = useState(
    profile?.personality_prompts?.[0]?.answer || ''
  )
  const [photos, setPhotos] = useState<{ id: string; url: string; file?: File }[]>([])
  const [saving, setSaving] = useState(false)

  const hasCulturalPrompts = (profile?.cultural_prompts?.length || 0) >= 2
  const hasPersonalityPrompt = (profile?.personality_prompts?.length || 0) >= 1
  const photoCount = (profile?.photos?.length || 0)

  const completeness = profile?.completeness || 0

  const savePrompts = async () => {
    setSaving(true)
    try {
      await api.updateProfile({
        cultural_prompts: culturalAnswers.length > 0 ? culturalAnswers : undefined,
        personality_prompts: personalityAnswer
          ? [{ prompt: personalityPrompt, answer: personalityAnswer }]
          : undefined,
      })

      for (const photo of photos) {
        if (photo.file) {
          try { await api.uploadPhoto(photo.file) } catch { /* continue */ }
        }
      }

      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
      onDone()
    } catch {
      alert('Failed to save. Please try again.')
    }
    setSaving(false)
  }

  return (
    <div className="container page">
      <h1 className="page-header">Boost your profile</h1>
      <p style={{ color: 'var(--text-secondary)', marginBottom: '4px' }}>
        Profiles with prompts get 3x more likes. Add yours now.
      </p>

      {/* Completeness bar */}
      <div style={{ marginBottom: '24px' }}>
        <div style={{
          width: '100%', height: '6px', borderRadius: '3px',
          background: 'var(--border)', overflow: 'hidden',
        }}>
          <div style={{
            width: `${completeness}%`, height: '100%',
            background: completeness >= 80 ? 'var(--accent)' : 'var(--primary)',
            borderRadius: '3px', transition: 'width 0.3s',
          }} />
        </div>
        <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
          {completeness}% complete
        </span>
      </div>

      {/* Cultural prompts */}
      <div className="card" onClick={() => setSelectedSection(selectedSection === 'cultural' ? null : 'cultural')}
        style={{ cursor: 'pointer' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <div style={{ fontWeight: '600' }}>Cultural prompts</div>
            <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              {hasCulturalPrompts ? 'Done' : 'Show what makes your culture special'}
            </div>
          </div>
          <span style={{ color: hasCulturalPrompts ? 'var(--accent)' : 'var(--primary)' }}>
            {hasCulturalPrompts ? '✓' : '+'}
          </span>
        </div>
      </div>

      {selectedSection === 'cultural' && (
        <div style={{ marginBottom: '16px' }}>
          <p style={{ fontSize: '13px', color: 'var(--text-secondary)', marginBottom: '12px' }}>
            Pick 2 prompts and write your answers.
          </p>
          {CULTURAL_PROMPTS.map((prompt) => {
            const isSelected = culturalPrompts.includes(prompt)
            const existing = culturalAnswers.find((p) => p.prompt === prompt)

            return (
              <div key={prompt}>
                <div className={`prompt-option ${isSelected ? 'selected' : ''}`}
                  onClick={() => {
                    if (isSelected) {
                      setCulturalPrompts((prev) => prev.filter((p) => p !== prompt))
                      setCulturalAnswers((prev) => prev.filter((p) => p.prompt !== prompt))
                    } else if (culturalPrompts.length < 2) {
                      setCulturalPrompts((prev) => [...prev, prompt])
                    }
                  }}>
                  {prompt}
                </div>
                {isSelected && (
                  <div className="input-group" style={{ marginTop: '4px' }}>
                    <textarea value={existing?.answer || ''} placeholder="Your answer..."
                      onChange={(e) => {
                        const updated = culturalAnswers.filter((p) => p.prompt !== prompt)
                        updated.push({ prompt, answer: e.target.value })
                        setCulturalAnswers(updated)
                      }} />
                  </div>
                )}
              </div>
            )
          })}
        </div>
      )}

      {/* Personality prompt */}
      <div className="card" onClick={() => setSelectedSection(selectedSection === 'personality' ? null : 'personality')}
        style={{ cursor: 'pointer' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <div style={{ fontWeight: '600' }}>Personality prompt</div>
            <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              {hasPersonalityPrompt ? 'Done' : 'Let your personality shine'}
            </div>
          </div>
          <span style={{ color: hasPersonalityPrompt ? 'var(--accent)' : 'var(--primary)' }}>
            {hasPersonalityPrompt ? '✓' : '+'}
          </span>
        </div>
      </div>

      {selectedSection === 'personality' && (
        <div style={{ marginBottom: '16px' }}>
          {PERSONALITY_PROMPTS.map((prompt) => (
            <div key={prompt}>
              <div className={`prompt-option ${personalityPrompt === prompt ? 'selected' : ''}`}
                onClick={() => setPersonalityPrompt(prompt)}>
                {prompt}
              </div>
              {personalityPrompt === prompt && (
                <div className="input-group" style={{ marginTop: '4px' }}>
                  <textarea value={personalityAnswer} placeholder="Your answer..."
                    onChange={(e) => setPersonalityAnswer(e.target.value)} />
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* More photos */}
      <div className="card" onClick={() => setSelectedSection(selectedSection === 'photos' ? null : 'photos')}
        style={{ cursor: 'pointer' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <div style={{ fontWeight: '600' }}>Add more photos</div>
            <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              {photoCount}/6 photos — profiles with 4+ get 60% more matches
            </div>
          </div>
          <span style={{ color: photoCount >= 4 ? 'var(--accent)' : 'var(--primary)' }}>
            {photoCount >= 4 ? '✓' : '+'}
          </span>
        </div>
      </div>

      {selectedSection === 'photos' && (
        <div style={{ marginBottom: '16px' }}>
          <PhotoUpload
            photos={photos}
            onAdd={(file) => {
              const localURL = URL.createObjectURL(file)
              setPhotos((prev) => [...prev, { id: `temp-${Date.now()}`, url: localURL, file }])
            }}
            onRemove={(id) => setPhotos((prev) => prev.filter((p) => p.id !== id))}
            maxPhotos={6 - photoCount}
          />
        </div>
      )}

      {/* Save */}
      <button className="btn btn-primary" style={{ marginTop: '8px' }}
        disabled={saving}
        onClick={savePrompts}>
        {saving ? 'Saving...' : 'Save & continue'}
      </button>

      <button className="btn btn-secondary" style={{ marginTop: '8px' }}
        onClick={onDone}>
        Skip for now
      </button>
    </div>
  )
}
