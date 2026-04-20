import { useLocation } from '../hooks/useLocation'

interface Props {
  city: string
  country: string
  onUpdate: (city: string, country: string, lat: number, lon: number) => void
}

export default function LocationPicker({ city, country, onUpdate }: Props) {
  const { detect, loading, error } = useLocation()

  const handleDetect = async () => {
    const result = await detect()
    if (result) {
      onUpdate(
        result.city || city,
        result.countryCode || country,
        result.latitude,
        result.longitude,
      )
    }
  }

  return (
    <div>
      <div style={{ display: 'flex', gap: '8px', marginBottom: '8px' }}>
        <div className="input-group" style={{ flex: 2, marginBottom: 0 }}>
          <label>City</label>
          <input value={city} onChange={(e) => onUpdate(e.target.value, country, 0, 0)}
            placeholder="Lagos, London..." />
        </div>
        <div className="input-group" style={{ flex: 1, marginBottom: 0 }}>
          <label>Country</label>
          <input value={country} onChange={(e) => onUpdate(city, e.target.value.toUpperCase(), 0, 0)}
            placeholder="NG" maxLength={3} />
        </div>
      </div>

      <button type="button" className="btn btn-secondary"
        style={{ fontSize: '13px', padding: '8px 14px' }}
        disabled={loading}
        onClick={handleDetect}>
        {loading ? 'Detecting...' : 'Use my location'}
      </button>

      {error && (
        <p style={{ fontSize: '12px', color: 'var(--text-secondary)', marginTop: '4px' }}>{error}</p>
      )}
    </div>
  )
}
