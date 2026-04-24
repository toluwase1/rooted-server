import { useNavigate } from 'react-router-dom'

export default function Guidelines() {
  const navigate = useNavigate()

  return (
    <div className="container page" style={{ paddingBottom: '40px' }}>
      <button
        className="btn btn-secondary"
        style={{ marginBottom: '16px', padding: '8px 16px', fontSize: '14px' }}
        onClick={() => navigate(-1)}
      >
        Back
      </button>

      <h1 className="page-header">Community Guidelines</h1>
      <p style={{ color: 'var(--text-secondary)', fontSize: '13px', marginBottom: '24px' }}>
        Last updated: April 22, 2026
      </p>

      <div style={{ fontSize: '14px', lineHeight: '1.7', color: 'var(--text-primary)' }}>
        <p>
          Rooted is a dating and connection platform built for the African diaspora. Our community
          is grounded in respect, cultural pride, and genuine connection. These guidelines exist to
          ensure that every member feels safe, valued, and respected. By using Rooted, you agree to
          follow these guidelines.
        </p>

        <h2>1. Respect and Dignity</h2>
        <ul>
          <li>Treat every member with respect, regardless of their background, heritage, gender, faith, or preferences.</li>
          <li>Communicate with kindness and good intentions. Remember that there is a real person on the other side of every profile.</li>
          <li>Accept rejection gracefully. Not every connection will be mutual, and that is okay.</li>
          <li>Do not pressure anyone into sharing personal information, meeting in person, or continuing a conversation.</li>
        </ul>

        <h2>2. Cultural Sensitivity</h2>
        <ul>
          <li>Rooted celebrates the diversity of the African diaspora. Respect the wide range of cultures, traditions, languages, and experiences represented in our community.</li>
          <li>Do not mock, belittle, or stereotype any culture, ethnicity, nationality, or heritage group.</li>
          <li>Engage with curiosity and openness when connecting with people from different backgrounds within the diaspora.</li>
          <li>Do not make assumptions about someone based on their heritage, diaspora status, or country of origin.</li>
        </ul>

        <h2>3. No Harassment, Abuse, or Threats</h2>
        <ul>
          <li>Harassment of any kind is strictly prohibited. This includes unwanted repeated contact, intimidation, and bullying.</li>
          <li>Do not send threatening, abusive, or hostile messages.</li>
          <li>Do not engage in stalking behavior, whether online or offline.</li>
          <li>Do not share or threaten to share another person's private information or intimate content without their consent.</li>
          <li>Do not make threats of physical violence or harm.</li>
        </ul>

        <h2>4. No Hate Speech</h2>
        <ul>
          <li>Content that promotes hatred, discrimination, or violence against individuals or groups based on race, ethnicity, nationality, religion, gender, sexual orientation, disability, or any other protected characteristic is strictly prohibited.</li>
          <li>This includes slurs, dehumanizing language, and content that promotes supremacist ideologies.</li>
        </ul>

        <h2>5. No Catfishing or Impersonation</h2>
        <ul>
          <li>Use only your own photos. Do not use photos of other people, celebrities, or AI-generated images as your profile photos.</li>
          <li>Do not misrepresent your identity, age, gender, heritage, or any other personal information.</li>
          <li>Do not create fake profiles or impersonate another person.</li>
          <li>Verified profiles (photo-verified) help build trust in our community. We encourage all members to verify their profiles.</li>
        </ul>

        <h2>6. No Explicit Content in Profiles</h2>
        <ul>
          <li>Profile photos must not contain nudity, sexually explicit content, or graphic imagery.</li>
          <li>Bios and prompt responses must not contain sexually explicit language or descriptions.</li>
          <li>Do not solicit explicit content from other users.</li>
          <li>Consensual private conversations between matched adults are your own responsibility, but any non-consensual sharing of intimate content is strictly prohibited and may be reported to law enforcement.</li>
        </ul>

        <h2>7. No Commercial Solicitation</h2>
        <ul>
          <li>Rooted is for genuine personal connections, not business.</li>
          <li>Do not use Rooted to promote products, services, or businesses.</li>
          <li>Do not solicit money, gifts, or financial assistance from other users.</li>
          <li>Do not recruit users for external platforms, apps, or websites.</li>
          <li>Do not engage in or promote scams, fraud, or deceptive schemes.</li>
        </ul>

        <h2>8. Reporting</h2>
        <p>
          If you encounter behavior that violates these guidelines, please report it immediately using the in-app report feature. When reporting:
        </p>
        <ul>
          <li>Select the appropriate category for the violation.</li>
          <li>Provide a clear description of what happened.</li>
          <li>Reports are reviewed by our team and kept confidential.</li>
          <li>We will not reveal your identity to the reported user.</li>
          <li>False reports made with the intent to harm another user may result in action against the reporting account.</li>
        </ul>

        <h2>9. Enforcement</h2>
        <p>
          Violations of these Community Guidelines may result in the following actions, at our sole discretion:
        </p>
        <ul>
          <li><strong>Warning:</strong> a notification that your behavior has violated the guidelines, with a request to correct the behavior.</li>
          <li><strong>Temporary suspension:</strong> temporary restriction of your ability to use certain features or access the service.</li>
          <li><strong>Permanent ban:</strong> permanent removal from the Rooted platform with no option to create a new account.</li>
        </ul>
        <p>
          The severity of the action depends on the nature and frequency of the violation. Repeated violations will result in escalating consequences.
        </p>

        <h2>10. Zero Tolerance for Exploitation</h2>
        <div
          style={{
            background: 'var(--bg-secondary, #f5f5f5)',
            border: '2px solid var(--text-secondary, #999)',
            borderRadius: '8px',
            padding: '16px',
            margin: '12px 0',
          }}
        >
          <p style={{ fontWeight: 'bold', marginBottom: '8px' }}>
            Rooted has a zero tolerance policy for any form of exploitation.
          </p>
          <p>
            This includes, but is not limited to: human trafficking, sexual exploitation, exploitation of minors, financial exploitation, and any form of coercion or manipulation. Any account found to be engaging in exploitative behavior will be permanently banned immediately, and the activity will be reported to the appropriate law enforcement authorities.
          </p>
        </div>

        <h2>Questions or Concerns</h2>
        <p>
          If you have questions about these Community Guidelines or want to report a concern that
          cannot be handled through the in-app reporting feature, contact us at:{' '}
          <a href="mailto:toluwase.dev@gmail.com">toluwase.dev@gmail.com</a>
        </p>
      </div>
    </div>
  )
}
