import { useRef } from 'react'

interface UploadedPhoto {
  id: string
  url: string
  file?: File
}

interface Props {
  photos: UploadedPhoto[]
  onAdd: (file: File) => void
  onRemove: (id: string) => void
  maxPhotos?: number
  uploading?: boolean
}

export default function PhotoUpload({ photos, onAdd, onRemove, maxPhotos = 6, uploading = false }: Props) {
  const inputRef = useRef<HTMLInputElement>(null)

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return

    // Validate type
    if (!file.type.startsWith('image/')) {
      alert('Please select an image file')
      return
    }

    // Validate size (5MB)
    if (file.size > 5 * 1024 * 1024) {
      alert('Photo must be under 5MB')
      return
    }

    onAdd(file)

    // Reset input so same file can be re-selected
    if (inputRef.current) inputRef.current.value = ''
  }

  const slots = Array.from({ length: maxPhotos }, (_, i) => photos[i] || null)

  return (
    <div>
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        style={{ display: 'none' }}
        onChange={handleFileSelect}
      />

      <div className="photo-grid">
        {slots.map((photo, i) => (
          <div
            key={i}
            className="photo-slot"
            onClick={() => {
              if (photo) return // clicking existing photo does nothing (long press to delete)
              if (photos.length < maxPhotos) inputRef.current?.click()
            }}
            style={i === 0 ? { gridColumn: 'span 2', gridRow: 'span 2' } : {}}
          >
            {photo ? (
              <div style={{ position: 'relative', width: '100%', height: '100%' }}>
                <img src={photo.url} alt={`Photo ${i + 1}`} />
                <button
                  onClick={(e) => { e.stopPropagation(); onRemove(photo.id) }}
                  style={{
                    position: 'absolute', top: '4px', right: '4px',
                    width: '24px', height: '24px', borderRadius: '50%',
                    background: 'rgba(0,0,0,0.6)', color: 'white',
                    border: 'none', fontSize: '14px', cursor: 'pointer',
                    display: 'flex', alignItems: 'center', justifyContent: 'center',
                  }}
                >
                  ✕
                </button>
                {i === 0 && (
                  <span style={{
                    position: 'absolute', bottom: '4px', left: '4px',
                    background: 'var(--primary)', color: 'white',
                    padding: '2px 8px', borderRadius: '10px', fontSize: '10px',
                  }}>
                    Primary
                  </span>
                )}
              </div>
            ) : (
              <div className="photo-slot-add">
                {uploading && i === photos.length ? '...' : '+'}
              </div>
            )}
          </div>
        ))}
      </div>

      <p style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '8px', textAlign: 'center' }}>
        {photos.length}/6 photos · First photo is your primary
      </p>
    </div>
  )
}
