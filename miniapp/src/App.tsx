import { useEffect, useState } from 'react'
import { Routes, Route, useNavigate, useLocation } from 'react-router-dom'
import { api } from './api/client'
import Nav from './components/Nav'
import Onboarding from './pages/Onboarding'
import CompleteProfile from './pages/CompleteProfile'
import Circle from './pages/Circle'
import Explore from './pages/Explore'
import Matches from './pages/Matches'
import ProfileView from './pages/ProfileView'
import Settings from './pages/Settings'
import Premium from './pages/Premium'
import Verify from './pages/Verify'
import EditProfile from './pages/EditProfile'
import ChatPage from './pages/Chat'

export default function App() {
  const [user, setUser] = useState<any>(null)
  const [profile, setProfile] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    window.Telegram?.WebApp?.ready()
    window.Telegram?.WebApp?.expand()

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
    navigate('/complete-profile')
  }

  const onProfileCompleted = () => {
    // Reload profile to get updated completeness
    if (user) {
      api.getMe().then((data) => {
        setProfile(data.profile)
      })
    }
    navigate('/')
  }

  if (loading) {
    return <div className="loading">Loading...</div>
  }

  const hiddenNavPaths = ['/onboarding', '/complete-profile', '/verify', '/edit-profile', '/chat/']
  const showNav = !hiddenNavPaths.some(p => location.pathname.startsWith(p))

  return (
    <>
      <Routes>
        <Route path="/" element={<Circle user={user} profile={profile} />} />
        <Route path="/onboarding" element={<Onboarding onComplete={onProfileCreated} />} />
        <Route path="/complete-profile" element={
          <CompleteProfile profile={profile} onDone={onProfileCompleted} />
        } />
        <Route path="/explore" element={<Explore user={user} />} />
        <Route path="/matches" element={<Matches user={user} />} />
        <Route path="/profile/:id" element={<ProfileView currentUserId={user?.id} />} />
        <Route path="/settings" element={<Settings user={user} profile={profile} />} />
        <Route path="/premium" element={<Premium user={user} />} />
        <Route path="/verify" element={<Verify />} />
        <Route path="/chat/:conversationId" element={<ChatPage userId={user?.id || ''} />} />
        <Route path="/edit-profile" element={
          <EditProfile profile={profile} onSaved={onProfileCompleted} />
        } />
      </Routes>
      {showNav && <Nav />}
    </>
  )
}
