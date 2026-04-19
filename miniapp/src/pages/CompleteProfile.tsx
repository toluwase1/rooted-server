import { useState, useEffect } from 'react'
import { api } from '../api/client'

const CULTURAL_PROMPTS = [
  'My culture means...', 'A tradition I love is...', 'Home is...',
  'My family would describe me as...', "I'm proudly...",
  'The food I\'d cook for you is...', 'My Sunday looks like...',
  'I switch between ___ and ___ cultures when...',
  'The song that defines me is...', 'What I miss most about home is...',
]

const PERSONALITY_PROMPTS = [
  'My ideal weekend is...', "I'm looking for someone who...",
  'The way to my heart is...', 'I geek out about...',
  'My friends would say I\'m...', 'A perfect first date is...',
  "The hill I'll die on is...", 'I recently discovered...',
]

interface Props {
  profile: any
  onDone: () => void
}

export default function CompleteProfile({ profile, onDone }: Props) {
  const [selectedSection, setSelectedSection] = useState<string | null>(null)

  // Cultural prompts state
  const [culturalPrompts, setCulturalPrompts] = useState<string[]>(
    (profile?.cultural_prompts || []).map((p: any) => p.prompt)
  )
  const [culturalAnswers, setCulturalAnswers] = useState<{ prompt: string; answer: string }[]>(
    profile?.cultural_prompts || []
  )

  // Personality prompt state
  const [personalityPrompt, setPersonalityPrompt] = useState(
    profile?.personality_prompts?.[0]?.prompt || ''
  )
  const [personalityAnswer, setPersonalityAnswer] = useState(
    profile?.personality_prompts?.[0]?.answer || ''
  )

  // Photos state — load existing from profile
  const [photos, setPhotos] = useState<{ id: string; url: string; isExisting: boolean }[]>([])
  const [uploading, setUploading] = useState(false)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (profile?.photos) {
      setPhotos(profile.photos.map((p: any) => ({
        id: p.id,
        url: p.url_medium || p.url_thumbnail,
        isExisting: true,
      })))
    }
  }, [profile])

  const hasCulturalPrompts = culturalAnswers.filter(p => p.answer).length >= 2
  const hasPersonalityPrompt = !!personalityAnswer
  const completeness = profile?.completeness || 0

  // Upload photo immediately when selected
  const handleAddPhoto = async (file: File) => {
    if (photos.length >= 6) return
    setUploading(true)
    try {
      const result = await api.uploadPhoto(file)
      setPhotos(prev => [...prev, { id: result.id, url: URL.createObjectURL(file), isExisting: true }])
      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    } catch (err: any) {
      alert(err.message || 'Failed to upload photo')
    }
    setUploading(false)
  }

  // Delete photo immediately
  const handleDeletePhoto = async (photoId: string) => {
    try {
      await api.deletePhoto(photoId)
      setPhotos(prev => prev.filter(p => p.id !== photoId))
    } catch {
      alert('Failed to delete photo')
    }
  }

  const savePrompts = async () => {
    setSaving(true)
    try {
      const updates: any = {}
      if (culturalAnswers.filter(p => p.answer).length > 0) {
        updates.cultural_prompts = culturalAnswers.filter(p => p.answer)
      }
      if (personalityAnswer && personalityPrompt) {
        updates.personality_prompts = [{ prompt: personalityPrompt, answer: personalityAnswer }]
      }
      if (Object.keys(updates).length > 0) {
        await api.updateProfile(updates)
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
        <span style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>{completeness}% complete</span>
      </div>

      {/* Photos — always visible, not collapsible */}
      <div className="card">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '12px' }}>
          <div style={{ fontWeight: '600' }}>Photos ({photos.length}/6)</div>
          {photos.length < 6 && (
            <label style={{
              color: 'var(--primary)', fontSize: '14px', fontWeight: '600', cursor: 'pointer',
            }}>
              {uploading ? 'Uploading...' : '+ Add'}
              <input type="file" accept="image/*" style={{ display: 'none' }}
                onChange={(e) => {
                  const file = e.target.files?.[0]
                  if (file) handleAddPhoto(file)
                  e.target.value = ''
                }} />
            </label>
          )}
        </div>

        {photos.length === 0 ? (
          <label style={{
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            height: '120px', border: '2px dashed var(--border)', borderRadius: 'var(--radius-sm)',
            color: 'var(--text-secondary)', cursor: 'pointer', fontSize: '14px',
          }}>
            {uploading ? 'Uploading...' : 'Tap to add your first photo'}
            <input type="file" accept="image/*" style={{ display: 'none' }}
              onChange={(e) => {
                const file = e.target.files?.[0]
                if (file) handleAddPhoto(file)
                e.target.value = ''
              }} />
          </label>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: '8px' }}>
            {photos.map((photo, i) => (
              <div key={photo.id} style={{
                aspectRatio: '1', borderRadius: 'var(--radius-sm)', overflow: 'hidden',
                position: 'relative',
              }}>
                <img src={photo.url} alt="" style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                {i === 0 && (
                  <span style={{
                    position: 'absolute', bottom: '4px', left: '4px',
                    background: 'var(--primary)', color: 'white',
                    padding: '2px 6px', borderRadius: '6px', fontSize: '9px',
                  }}>Primary</span>
                )}
                <button onClick={() => handleDeletePhoto(photo.id)} style={{
                  position: 'absolute', top: '4px', right: '4px',
                  width: '22px', height: '22px', borderRadius: '50%',
                  background: 'rgba(0,0,0,0.6)', color: 'white',
                  border: 'none', fontSize: '12px', cursor: 'pointer',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                }}>✕</button>
              </div>
            ))}

            {photos.length < 6 && (
              <label style={{
                aspectRatio: '1', borderRadius: 'var(--radius-sm)',
                border: '2px dashed var(--border)', display: 'flex',
                alignItems: 'center', justifyContent: 'center',
                cursor: 'pointer', fontSize: '24px', color: 'var(--text-secondary)',
              }}>
                {uploading ? '...' : '+'}
                <input type="file" accept="image/*" style={{ display: 'none' }}
                  onChange={(e) => {
                    const file = e.target.files?.[0]
                    if (file) handleAddPhoto(file)
                    e.target.value = ''
                  }} />
              </label>
            )}
          </div>
        )}

        <p style={{ fontSize: '11px', color: 'var(--text-secondary)', marginTop: '8px' }}>
          Profiles with 4+ photos get 60% more matches
        </p>
      </div>

      {/* Cultural prompts */}
      <div className="card" onClick={() => setSelectedSection(selectedSection === 'cultural' ? null : 'cultural')}
        style={{ cursor: 'pointer' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <div style={{ fontWeight: '600' }}>Cultural prompts</div>
            <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              {hasCulturalPrompts ? `${culturalAnswers.filter(p => p.answer).length} answered` : 'Show what makes your culture special'}
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
                      setCulturalPrompts(prev => prev.filter(p => p !== prompt))
                      setCulturalAnswers(prev => prev.filter(p => p.prompt !== prompt))
                    } else if (culturalPrompts.length < 2) {
                      setCulturalPrompts(prev => [...prev, prompt])
                    }
                  }}>
                  {prompt}
                </div>
                {isSelected && (
                  <div className="input-group" style={{ marginTop: '4px' }}>
                    <textarea value={existing?.answer || ''} placeholder="Your answer..."
                      onChange={(e) => {
                        const updated = culturalAnswers.filter(p => p.prompt !== prompt)
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

      {/* Save */}
      <button className="btn btn-primary" style={{ marginTop: '8px' }}
        disabled={saving}
        onClick={savePrompts}>
        {saving ? 'Saving...' : 'Save & continue'}
      </button>

      <button className="btn btn-secondary" style={{ marginTop: '8px' }}
        onClick={onDone}>
        {photos.length > 0 || hasCulturalPrompts || hasPersonalityPrompt ? 'Done' : 'Skip for now'}
      </button>
    </div>
  )
}
