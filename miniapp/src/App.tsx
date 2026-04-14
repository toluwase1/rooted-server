import { useEffect, useState } from 'react'
import { Routes, Route, useNavigate, useLocation } from 'react-router-dom'
import { api } from './api/client'
import Nav from './components/Nav'
import Onboarding from './pages/Onboarding'
import Circle from './pages/Circle'
import Explore from './pages/Explore'
import Matches from './pages/Matches'
import ProfileView from './pages/ProfileView'
import Settings from './pages/Settings'
import Premium from './pages/Premium'
import Verify from './pages/Verify'

export default function App() {
  const [user, setUser] = useState<any>(null)
  const [profile, setProfile] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    // Tell Telegram the Mini App is ready
    window.Telegram?.WebApp?.ready()
    window.Telegram?.WebApp?.expand()

    // Load user data
    api.getMe().then((data) => {
      setUser(data.user)
      setProfile(data.profile)
      setLoading(false)

      if (data.is_new || !data.profile) {
        navigate('/onboarding')
      }
    }).catch(() => {
      setLoading(false)
    })
  }, [])

  const onProfileCreated = (newProfile: any) => {
    setProfile(newProfile)
    navigate('/')
  }

  if (loading) {
    return <div className="loading">Loading...</div>
  }

  const showNav = !['/onboarding', '/profile/'].some(p => location.pathname.startsWith(p))
    && location.pathname !== '/onboarding'

  return (
    <>
      <Routes>
        <Route path="/" element={<Circle user={user} />} />
        <Route path="/onboarding" element={<Onboarding onComplete={onProfileCreated} />} />
        <Route path="/explore" element={<Explore user={user} />} />
        <Route path="/matches" element={<Matches user={user} />} />
        <Route path="/profile/:id" element={<ProfileView currentUserId={user?.id} />} />
        <Route path="/settings" element={<Settings user={user} profile={profile} />} />
        <Route path="/premium" element={<Premium user={user} />} />
        <Route path="/verify" element={<Verify />} />
      </Routes>
      {showNav && <Nav />}
    </>
  )
}
