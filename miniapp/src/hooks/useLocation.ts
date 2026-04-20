import { useState } from 'react'

interface LocationResult {
  city: string
  country: string      // full name
  countryCode: string   // ISO 3166-1 alpha-2 (NG, GB, US)
  latitude: number
  longitude: number
}

export function useLocation() {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const detect = (): Promise<LocationResult | null> => {
    return new Promise((resolve) => {
      if (!navigator.geolocation) {
        setError('Location not available on this device')
        resolve(null)
        return
      }

      setLoading(true)
      setError('')

      navigator.geolocation.getCurrentPosition(
        async (pos) => {
          const { latitude, longitude } = pos.coords

          // Try reverse geocoding
          try {
            const res = await fetch(
              `https://api.bigdatacloud.net/data/reverse-geocode-client?latitude=${latitude}&longitude=${longitude}&localityLanguage=en`
            )
            if (res.ok) {
              const data = await res.json()
              setLoading(false)
              resolve({
                city: data.city || data.locality || '',
                country: data.countryName || '',
                countryCode: data.countryCode || '',
                latitude,
                longitude,
              })
              return
            }
          } catch {
            // API failed — fall back to coordinates only
          }

          setLoading(false)
          resolve({ city: '', country: '', countryCode: '', latitude, longitude })
        },
        () => {
          setLoading(false)
          setError('Location permission denied')
          resolve(null)
        },
        { enableHighAccuracy: false, timeout: 10000 }
      )
    })
  }

  return { detect, loading, error }
}
