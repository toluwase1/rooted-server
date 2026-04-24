import { useNavigate } from 'react-router-dom'

export default function Privacy() {
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

      <h1 className="page-header">Privacy Policy</h1>
      <p style={{ color: 'var(--text-secondary)', fontSize: '13px', marginBottom: '24px' }}>
        Last updated: April 22, 2026
      </p>

      <div style={{ fontSize: '14px', lineHeight: '1.7', color: 'var(--text-primary)' }}>
        <p>
          This Privacy Policy explains how Rooted ("we," "us," or "our") collects, uses, shares, and
          protects your personal data when you use the Rooted application, a Telegram Mini App for
          dating and connection within the African diaspora.
        </p>

        <h2>1. Data Controller</h2>
        <p>
          The data controller responsible for your personal data is:
        </p>
        <ul>
          <li><strong>Entity:</strong> Rooted</li>
          <li><strong>Contact:</strong> <a href="mailto:toluwase.dev@gmail.com">toluwase.dev@gmail.com</a></li>
        </ul>

        <h2>2. Data We Collect</h2>

        <h3>2.1 Data You Provide</h3>
        <ul>
          <li><strong>Profile information:</strong> name, date of birth, gender, gender preference, heritage, diaspora status, intention (dating/friendship), faith, faith importance, bio, city, country.</li>
          <li><strong>Photos:</strong> profile photos and verification selfies you upload.</li>
          <li><strong>Messages:</strong> content of messages you send through the in-app chat.</li>
          <li><strong>Preferences:</strong> matching preferences, settings, and configurations.</li>
          <li><strong>Reports and feedback:</strong> information you provide when reporting other users or contacting support.</li>
        </ul>

        <h3>2.2 Data Received from Telegram</h3>
        <p>When you open Rooted through Telegram, we receive:</p>
        <ul>
          <li>Your Telegram user ID (a numeric identifier).</li>
          <li>Your Telegram display name and username.</li>
        </ul>
        <p>
          <strong>We do NOT access your Telegram phone number, contacts, or message history.</strong> Telegram's Mini App platform does not provide us with your phone number.
        </p>

        <h3>2.3 Data Collected Automatically</h3>
        <ul>
          <li>Device information (browser type, operating system).</li>
          <li>Usage data (features used, actions taken, session duration).</li>
          <li>Approximate location (based on city/country you provide, or coordinates if you grant location access).</li>
          <li>Timestamps of activity (last active, profile creation date).</li>
        </ul>

        <h2>3. Special Category Data</h2>
        <div
          style={{
            background: 'var(--bg-secondary, #f5f5f5)',
            border: '2px solid var(--text-secondary, #999)',
            borderRadius: '8px',
            padding: '16px',
            margin: '12px 0',
          }}
        >
          <p>
            Rooted collects data that may be considered <strong>special category data</strong> under data protection laws such as GDPR (Article 9) and Nigeria's NDPA:
          </p>
          <ul>
            <li><strong>Ethnicity/Heritage:</strong> used for cultural matching within the African diaspora community.</li>
            <li><strong>Religious beliefs (Faith):</strong> used for compatibility matching when you indicate faith is important.</li>
          </ul>
          <p>
            We process this data based on your <strong>explicit consent</strong>, which you provide when creating your profile. You can choose not to disclose faith information by selecting "Prefer not to say." Heritage information is required for the core matching functionality of the service.
          </p>
        </div>

        <h2>4. Legal Bases for Processing</h2>

        <h3>GDPR Article 6 (General Processing)</h3>
        <ul>
          <li><strong>Consent (Art. 6(1)(a)):</strong> for processing special category data, sending optional notifications.</li>
          <li><strong>Contract (Art. 6(1)(b)):</strong> for providing the core matching and messaging service as described in our Terms of Service.</li>
          <li><strong>Legitimate Interest (Art. 6(1)(f)):</strong> for security, fraud prevention, service improvement, and analytics.</li>
          <li><strong>Legal Obligation (Art. 6(1)(c)):</strong> for compliance with applicable laws and regulations.</li>
        </ul>

        <h3>GDPR Article 9 (Special Category Data)</h3>
        <ul>
          <li><strong>Explicit Consent (Art. 9(2)(a)):</strong> you explicitly consent to the processing of your heritage and faith data when you create your profile and agree to this Privacy Policy.</li>
        </ul>

        <h2>5. How We Use Your Data</h2>
        <ul>
          <li><strong>Matching:</strong> to suggest compatible profiles based on heritage, location, preferences, faith, and other factors.</li>
          <li><strong>Messaging:</strong> to facilitate communication between matched users.</li>
          <li><strong>Moderation:</strong> to enforce our Community Guidelines and protect user safety.</li>
          <li><strong>Analytics:</strong> to understand how the service is used and to improve features.</li>
          <li><strong>Notifications:</strong> to inform you of matches, messages, and relevant activity.</li>
          <li><strong>Account management:</strong> to maintain your account, process payments, and provide support.</li>
        </ul>

        <h2>6. Data Sharing</h2>
        <p>
          <strong>We do NOT sell your personal data.</strong> We share data only in these limited circumstances:
        </p>
        <ul>
          <li><strong>Other users:</strong> your profile information (name, photos, bio, heritage, faith, city, etc.) is visible to other users as part of the matching experience. Messages are shared only with the intended recipient.</li>
          <li><strong>Cloud infrastructure providers:</strong> Google Cloud Platform (GCP) and Cloudflare, which host our servers and deliver content.</li>
          <li><strong>Telegram:</strong> as required for the Mini App platform to function. We do not send your Rooted profile data back to Telegram.</li>
          <li><strong>Law enforcement:</strong> if required by valid legal process (court order, subpoena, or applicable law).</li>
          <li><strong>Business transfers:</strong> in connection with a merger, acquisition, or sale of assets, with notice to affected users.</li>
        </ul>

        <h2>7. International Data Transfers</h2>
        <p>
          Your data may be processed and stored in:
        </p>
        <ul>
          <li>United States (Google Cloud Platform servers)</li>
          <li>Nigeria (development and administration)</li>
        </ul>
        <p>
          For transfers from the European Economic Area (EEA) or the United Kingdom, we rely on Standard Contractual Clauses (SCCs) approved by the European Commission, or other lawful transfer mechanisms. For transfers from Nigeria, we comply with the requirements of the Nigeria Data Protection Act (NDPA).
        </p>

        <h2>8. Data Retention</h2>
        <ul>
          <li><strong>Active account:</strong> we retain your data for as long as your account is active.</li>
          <li><strong>Account deletion:</strong> when you delete your account, we remove your profile data within 30 days. Some data may be retained longer if required by law or for legitimate business purposes (such as resolving disputes or preventing fraud).</li>
          <li><strong>Messages:</strong> deleted when both users' accounts are deleted or the conversation is removed.</li>
          <li><strong>Reports:</strong> retained for safety purposes even after account deletion.</li>
        </ul>

        <h2>9. Your Rights</h2>

        <h3>For EEA/UK Residents (GDPR)</h3>
        <p>You have the right to:</p>
        <ul>
          <li>Access your personal data.</li>
          <li>Rectify inaccurate data.</li>
          <li>Erase your data ("right to be forgotten").</li>
          <li>Restrict processing in certain circumstances.</li>
          <li>Data portability (receive your data in a machine-readable format).</li>
          <li>Object to processing based on legitimate interests.</li>
          <li>Withdraw consent at any time (without affecting the lawfulness of prior processing).</li>
          <li>Lodge a complaint with your local data protection authority.</li>
        </ul>

        <h3>For Nigerian Residents (NDPA)</h3>
        <p>You have the right to:</p>
        <ul>
          <li>Be informed about the processing of your personal data.</li>
          <li>Access your personal data.</li>
          <li>Rectify inaccurate data.</li>
          <li>Erase your data.</li>
          <li>Restrict processing.</li>
          <li>Data portability.</li>
          <li>Object to processing.</li>
          <li>Not be subject to decisions based solely on automated processing.</li>
          <li>Lodge a complaint with the Nigeria Data Protection Commission (NDPC).</li>
        </ul>

        <h3>For California Residents (CCPA/CPRA)</h3>
        <p>You have the right to:</p>
        <ul>
          <li>Know what personal information is collected, used, shared, or sold.</li>
          <li>Delete personal information held by us.</li>
          <li>Opt out of the sale of personal information (we do not sell your data).</li>
          <li>Non-discrimination for exercising your privacy rights.</li>
          <li>Correct inaccurate personal information.</li>
        </ul>

        <p>
          To exercise any of these rights, contact us at <a href="mailto:toluwase.dev@gmail.com">toluwase.dev@gmail.com</a>. We will respond within 30 days (or sooner where required by applicable law).
        </p>

        <h2>10. Phone Number Privacy</h2>
        <p>
          Rooted operates as a Telegram Mini App. <strong>We do not access, collect, or store your Telegram phone number.</strong> The Telegram Mini App platform does not provide phone numbers to Mini Apps. Your phone number remains private to Telegram.
        </p>

        <h2>11. Children's Privacy</h2>
        <p>
          Rooted is strictly for users aged 18 and older. We do not knowingly collect personal data from anyone under the age of 18. If we become aware that a user is under 18, we will promptly delete their account and associated data.
        </p>

        <h2>12. Security Measures</h2>
        <p>
          We implement appropriate technical and organizational measures to protect your personal data, including:
        </p>
        <ul>
          <li>Encryption of data in transit (TLS/HTTPS).</li>
          <li>Encrypted data storage.</li>
          <li>Access controls and authentication.</li>
          <li>Regular security reviews.</li>
          <li>Secure cloud infrastructure (Google Cloud Platform).</li>
        </ul>
        <p>
          While we take reasonable precautions, no method of transmission or storage is 100% secure. We cannot guarantee absolute security of your data.
        </p>

        <h2>13. Changes to This Policy</h2>
        <p>
          We may update this Privacy Policy from time to time. When we make material changes, we will notify you through the app or via Telegram. Your continued use of Rooted after such changes constitutes acceptance of the updated policy.
        </p>

        <h2>14. Contact Us</h2>
        <p>
          For privacy-related questions, data access requests, or concerns, contact us at:
        </p>
        <ul>
          <li><strong>Email:</strong> <a href="mailto:toluwase.dev@gmail.com">toluwase.dev@gmail.com</a></li>
        </ul>
      </div>
    </div>
  )
}
