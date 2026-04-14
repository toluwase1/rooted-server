import { useState, useRef, useCallback } from 'react'
import { api } from '../api/client'
import { useNavigate } from 'react-router-dom'

const POSES = [
  { instruction: 'Look straight at the camera and smile', icon: '😊' },
  { instruction: 'Turn your head slightly to the left', icon: '👈' },
  { instruction: 'Give a thumbs up', icon: '👍' },
  { instruction: 'Hold your hand next to your face', icon: '✋' },
]

export default function Verify() {
  const navigate = useNavigate()
  const [step, setStep] = useState<'intro' | 'capture' | 'uploading' | 'success' | 'failed'>('intro')
  const [pose] = useState(() => POSES[Math.floor(Math.random() * POSES.length)])
  const videoRef = useRef<HTMLVideoElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const streamRef = useRef<MediaStream | null>(null)

  const startCamera = useCallback(async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({
        video: { facingMode: 'user', width: 640, height: 480 },
      })
      streamRef.current = stream
      if (videoRef.current) {
        videoRef.current.srcObject = stream
      }
      setStep('capture')
    } catch {
      alert('Camera access is required for verification. Please allow camera access and try again.')
    }
  }, [])

  const stopCamera = () => {
    if (streamRef.current) {
      streamRef.current.getTracks().forEach((t) => t.stop())
      streamRef.current = null
    }
  }

  const captureAndSubmit = async () => {
    if (!videoRef.current || !canvasRef.current) return

    const canvas = canvasRef.current
    const video = videoRef.current
    canvas.width = video.videoWidth
    canvas.height = video.videoHeight
    canvas.getContext('2d')?.drawImage(video, 0, 0)

    stopCamera()
    setStep('uploading')

    // Convert canvas to blob
    canvas.toBlob(async (blob) => {
      if (!blob) {
        setStep('failed')
        return
      }

      try {
        const file = new File([blob], 'verification-selfie.jpg', { type: 'image/jpeg' })
        await api.uploadVerificationSelfie(file)
        window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
        setStep('success')
      } catch {
        setStep('failed')
      }
    }, 'image/jpeg', 0.9)
  }

  return (
    <div className="container page">
      {step === 'intro' && (
        <>
          <h1 className="page-header">Verify your photo</h1>
          <div className="card" style={{ textAlign: 'center', padding: '32px 16px' }}>
            <div style={{ fontSize: '48px', marginBottom: '16px' }}>📸</div>
            <h3 style={{ marginBottom: '8px' }}>Quick selfie check</h3>
            <p style={{ color: 'var(--text-secondary)', marginBottom: '20px' }}>
              We'll ask you to take a selfie with a specific pose to confirm you're real.
              This keeps the community safe and trustworthy.
            </p>
            <p style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              Your selfie is compared to your profile photos and then deleted.
              It's never shown to other users.
            </p>
          </div>
          <button className="btn btn-primary" style={{ marginTop: '16px' }} onClick={startCamera}>
            Start verification
          </button>
          <button className="btn btn-secondary" style={{ marginTop: '8px' }} onClick={() => navigate(-1)}>
            Maybe later
          </button>
        </>
      )}

      {step === 'capture' && (
        <>
          <div style={{
            background: 'var(--card-bg)', borderRadius: 'var(--radius)',
            padding: '16px', textAlign: 'center', marginBottom: '16px',
          }}>
            <span style={{ fontSize: '32px' }}>{pose.icon}</span>
            <p style={{ fontWeight: '600', marginTop: '8px' }}>{pose.instruction}</p>
          </div>

          <div style={{
            borderRadius: 'var(--radius)', overflow: 'hidden',
            background: '#000', marginBottom: '16px',
          }}>
            <video
              ref={videoRef}
              autoPlay
              playsInline
              muted
              style={{ width: '100%', display: 'block', transform: 'scaleX(-1)' }}
            />
          </div>

          <canvas ref={canvasRef} style={{ display: 'none' }} />

          <button className="btn btn-primary" onClick={captureAndSubmit}>
            Take selfie
          </button>
          <button className="btn btn-secondary" style={{ marginTop: '8px' }}
            onClick={() => { stopCamera(); setStep('intro') }}>
            Cancel
          </button>
        </>
      )}

      {step === 'uploading' && (
        <div className="loading">Verifying your photo...</div>
      )}

      {step === 'success' && (
        <div style={{ textAlign: 'center', padding: '40px 20px' }}>
          <div style={{ fontSize: '48px', marginBottom: '16px' }}>✓</div>
          <h2 style={{ color: 'var(--accent)', marginBottom: '8px' }}>Verified!</h2>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '24px' }}>
            Your profile now has a blue verification badge. This helps you
            stand out and builds trust with your matches.
          </p>
          <button className="btn btn-primary" onClick={() => navigate('/settings')}>
            Done
          </button>
        </div>
      )}

      {step === 'failed' && (
        <div style={{ textAlign: 'center', padding: '40px 20px' }}>
          <div style={{ fontSize: '48px', marginBottom: '16px' }}>✕</div>
          <h2 style={{ marginBottom: '8px' }}>Verification failed</h2>
          <p style={{ color: 'var(--text-secondary)', marginBottom: '24px' }}>
            We couldn't match your selfie to your profile photos.
            Make sure your face is clearly visible and try again.
          </p>
          <button className="btn btn-primary" onClick={() => setStep('intro')}>
            Try again
          </button>
        </div>
      )}
    </div>
  )
}
