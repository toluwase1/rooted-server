import { useEffect, useState, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { api } from '../api/client'

interface Message {
  id: string
  conversation_id: string
  sender_id: string
  sender_name?: string
  content_type: string
  content: string
  created_at: string
}

interface Props {
  userId: string
}

export default function Chat({ userId }: Props) {
  const { conversationId } = useParams()
  const navigate = useNavigate()
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [otherName, setOtherName] = useState('')
  const [loading, setLoading] = useState(true)
  const [sending, setSending] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    window.Telegram?.WebApp?.BackButton?.show()
    window.Telegram?.WebApp?.BackButton?.onClick(() => navigate('/matches'))

    // Load existing messages
    if (conversationId) {
      api.getMessages(conversationId).then((data) => {
        setMessages((data.messages || []).reverse())
        setLoading(false)
      }).catch(() => setLoading(false))

      // Load conversation info for the other user's name
      api.getConversations().then((data) => {
        const conv = (data.conversations || []).find(
          (c: any) => c.conversation.id === conversationId
        )
        if (conv) {
          setOtherName(conv.other_user_name || 'Match')
        }
      })
    }

    // Connect WebSocket
    const wsUrl = (import.meta.env.VITE_API_URL || '').replace('https://', 'wss://').replace('http://', 'ws://')
    const ws = new WebSocket(`${wsUrl}/ws/chat?user_id=${userId}`)

    ws.onmessage = (event) => {
      const msg = JSON.parse(event.data)
      if (msg.type === 'message' && msg.conversation_id === conversationId) {
        setMessages(prev => [...prev, msg])
        scrollToBottom()
      }
    }

    ws.onerror = () => {
      // WebSocket not available — fall back to polling
      console.log('WebSocket not available, using polling')
    }

    wsRef.current = ws

    // Poll for new messages every 3 seconds as fallback
    const pollInterval = setInterval(() => {
      if (conversationId) {
        api.getMessages(conversationId).then((data) => {
          const serverMsgs = (data.messages || []).reverse()
          setMessages(prev => {
            // Count real messages (not temp/optimistic)
            const realCount = prev.filter(m => !m.id.startsWith('temp-')).length
            if (serverMsgs.length > realCount) {
              return serverMsgs
            }
            return prev
          })
        }).catch(() => {})
      }
    }, 3000)

    return () => {
      window.Telegram?.WebApp?.BackButton?.hide()
      ws.close()
      clearInterval(pollInterval)
    }
  }, [conversationId])

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  const sendMessage = async () => {
    if (!input.trim() || !conversationId || sending) return

    const content = input.trim()
    setInput('')
    setSending(true)
    inputRef.current?.focus()

    // Optimistic update
    const optimistic: Message = {
      id: `temp-${Date.now()}`,
      conversation_id: conversationId,
      sender_id: userId,
      content_type: 'text',
      content,
      created_at: new Date().toISOString(),
    }
    setMessages(prev => [...prev, optimistic])

    // Try WebSocket first
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'message',
        conversation_id: conversationId,
        content_type: 'text',
        content,
      }))
      setSending(false)
      return
    }

    // Fallback: HTTP POST
    try {
      await api.sendMessage(conversationId, content)
    } catch {
      // Message saved via optimistic update, will sync on next poll
    }
    setSending(false)
  }

  const formatTime = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }

  if (loading) return <div className="loading">Loading chat...</div>

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
      {/* Header */}
      <div style={{
        padding: '12px 16px', borderBottom: '1px solid var(--border)',
        display: 'flex', alignItems: 'center', gap: '12px',
        background: 'var(--bg)',
      }}>
        <div style={{
          width: '36px', height: '36px', borderRadius: '50%',
          background: 'linear-gradient(135deg, #E07A5F, #81B29A)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          color: 'white', fontSize: '16px', fontWeight: 'bold',
        }}>
          {otherName[0] || '?'}
        </div>
        <div style={{ fontWeight: '600' }}>{otherName}</div>
      </div>

      {/* Messages */}
      <div style={{
        flex: 1, overflowY: 'auto', padding: '12px 16px',
        display: 'flex', flexDirection: 'column', gap: '6px',
      }}>
        {messages.length === 0 && (
          <div style={{ textAlign: 'center', color: 'var(--text-secondary)', padding: '40px 0', fontSize: '14px' }}>
            Say hello! Start the conversation.
          </div>
        )}

        {messages.map((msg) => {
          const isMine = msg.sender_id === userId
          return (
            <div key={msg.id} style={{
              display: 'flex',
              justifyContent: isMine ? 'flex-end' : 'flex-start',
            }}>
              <div style={{
                maxWidth: '75%',
                padding: '8px 12px',
                borderRadius: isMine ? '16px 16px 4px 16px' : '16px 16px 16px 4px',
                background: isMine ? 'var(--primary)' : 'var(--card-bg)',
                color: isMine ? 'white' : 'var(--text)',
                fontSize: '15px',
                lineHeight: '1.4',
              }}>
                <div>{msg.content}</div>
                <div style={{
                  fontSize: '10px',
                  opacity: 0.6,
                  textAlign: 'right',
                  marginTop: '2px',
                }}>
                  {formatTime(msg.created_at)}
                </div>
              </div>
            </div>
          )
        })}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div style={{
        padding: '8px 12px',
        paddingBottom: 'calc(8px + env(safe-area-inset-bottom, 0px))',
        borderTop: '1px solid var(--border)',
        display: 'flex', gap: '8px', alignItems: 'center',
        background: 'var(--bg)',
      }}>
        <input
          ref={inputRef}
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
          placeholder="Type a message..."
          style={{
            flex: 1, padding: '10px 14px', borderRadius: '20px',
            border: '1px solid var(--border)', background: 'var(--card-bg)',
            fontSize: '15px', color: 'var(--text)', outline: 'none',
          }}
        />
        <button
          onClick={sendMessage}
          disabled={!input.trim() || sending}
          style={{
            width: '40px', height: '40px', borderRadius: '50%',
            background: input.trim() ? 'var(--primary)' : 'var(--border)',
            color: 'white', border: 'none', fontSize: '18px',
            cursor: input.trim() ? 'pointer' : 'default',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}
        >
          →
        </button>
      </div>
    </div>
  )
}
