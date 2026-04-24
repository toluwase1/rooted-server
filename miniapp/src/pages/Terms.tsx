import { useNavigate } from 'react-router-dom'

export default function Terms() {
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

      <h1 className="page-header">Terms of Service</h1>
      <p style={{ color: 'var(--text-secondary)', fontSize: '13px', marginBottom: '24px' }}>
        Last updated: April 22, 2026
      </p>

      <div style={{ fontSize: '14px', lineHeight: '1.7', color: 'var(--text-primary)' }}>
        <p>
          Welcome to Rooted ("we," "us," or "our"). These Terms of Service ("Terms") govern your
          access to and use of the Rooted application, a Telegram Mini App available through the
          Telegram platform. By creating an account or using Rooted, you agree to be bound by these
          Terms. If you do not agree, do not use the service.
        </p>

        <h2>1. Eligibility</h2>
        <p>To use Rooted, you must:</p>
        <ul>
          <li>Be at least 18 years of age.</li>
          <li>Not be a person required to register on any sex offender registry.</li>
          <li>Not have been previously banned from Rooted.</li>
          <li>Have the legal capacity to enter into a binding agreement.</li>
          <li>Comply with all applicable laws in your jurisdiction.</li>
        </ul>
        <p>
          By creating an account, you represent and warrant that you meet all of the above eligibility
          requirements.
        </p>

        <h2>2. Account Responsibilities</h2>
        <p>
          You are responsible for maintaining the confidentiality of your account and for all
          activities that occur under your account. You agree to:
        </p>
        <ul>
          <li>Provide accurate, current, and complete information during registration.</li>
          <li>Keep your profile information up to date.</li>
          <li>Not create more than one account.</li>
          <li>Not transfer your account to another person.</li>
          <li>Notify us immediately if you suspect unauthorized access to your account.</li>
        </ul>

        <h2>3. Telegram Mini App Acknowledgment</h2>
        <p>
          Rooted operates as a Telegram Mini App. By using Rooted, you acknowledge and agree that:
        </p>
        <ul>
          <li>
            <strong>Rooted is NOT affiliated with, endorsed by, or sponsored by Telegram.</strong>
          </li>
          <li>
            Telegram's own terms of service and privacy policy apply separately to your use of the
            Telegram platform.
          </li>
          <li>
            Certain data (such as your Telegram user ID, display name, and username) is received from
            Telegram when you access Rooted.
          </li>
          <li>
            We are not responsible for the availability, functionality, or security of the Telegram
            platform itself.
          </li>
          <li>
            Payments processed through Telegram Stars are subject to Telegram's payment terms and
            policies.
          </li>
        </ul>

        <h2>4. User Conduct</h2>
        <p>You agree NOT to:</p>
        <ul>
          <li>Harass, abuse, threaten, stalk, or intimidate any other user.</li>
          <li>Post or transmit content that is hateful, discriminatory, or promotes violence.</li>
          <li>Impersonate any person or entity, or create a fake identity.</li>
          <li>Use Rooted for any commercial purpose, solicitation, or advertising.</li>
          <li>Attempt to gain unauthorized access to other users' accounts or data.</li>
          <li>Use automated systems, bots, or scrapers to access the service.</li>
          <li>Engage in any activity that interferes with or disrupts the service.</li>
          <li>Violate any applicable law or regulation.</li>
          <li>Solicit money or financial information from other users.</li>
          <li>Send unsolicited sexual content or engage in sexual harassment.</li>
        </ul>

        <h2>5. Content Standards</h2>
        <p>All content you post on Rooted must:</p>
        <ul>
          <li>Be accurate and represent you truthfully.</li>
          <li>Not contain nudity, sexually explicit material, or pornography.</li>
          <li>Not contain violent, graphic, or disturbing imagery.</li>
          <li>Not infringe upon any third party's intellectual property rights.</li>
          <li>Not contain personal contact information of others without their consent.</li>
        </ul>
        <p>
          We reserve the right to remove any content that violates these standards at our sole
          discretion, without prior notice.
        </p>

        <h2>6. No Background Checks</h2>
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
            ROOTED DOES NOT CONDUCT CRIMINAL BACKGROUND CHECKS, IDENTITY VERIFICATION BEYOND PHOTO
            MATCHING, OR SEX OFFENDER REGISTRY SEARCHES ON ITS USERS.
          </p>
          <p>
            You are solely responsible for taking appropriate precautions when interacting with other
            users. We make no representations or warranties regarding the conduct, identity, health,
            intentions, legitimacy, or veracity of any user.
          </p>
        </div>

        <h2>7. No Guarantee of Outcomes</h2>
        <p>
          Rooted makes <strong>NO GUARANTEE</strong> that you will receive matches, dates,
          relationships, or any particular outcome from using the service. The quality and quantity of
          matches depend on many factors, including your profile completeness, location, preferences,
          and the current user base.
        </p>

        <h2>8. Safety and In-Person Meetings</h2>
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
            ROOTED IS NOT LIABLE FOR ANY IN-PERSON MEETINGS, INTERACTIONS, INJURY, HARM, DEATH, OR
            DAMAGES OF ANY KIND ARISING FROM CONNECTIONS MADE THROUGH THE SERVICE.
          </p>
          <p>
            You assume all risk when choosing to meet another user in person. We strongly recommend
            meeting in public places, informing someone you trust about your plans, and exercising
            caution at all times.
          </p>
        </div>

        <h2>9. User Content</h2>
        <p>
          You retain ownership of the content you post on Rooted. By posting content, you grant us a
          non-exclusive, worldwide, royalty-free, transferable license to use, reproduce, modify,
          distribute, and display your content in connection with operating and promoting the service.
        </p>
        <p>
          You are solely responsible for the content you post. We do not endorse or guarantee the
          accuracy of any user content and are not liable for any content posted by users.
        </p>

        <h2>10. Service Provided "AS IS"</h2>
        <p>
          ROOTED IS PROVIDED ON AN "AS IS" AND "AS AVAILABLE" BASIS, WITHOUT ANY WARRANTIES OF ANY
          KIND, WHETHER EXPRESS, IMPLIED, OR STATUTORY. WE SPECIFICALLY DISCLAIM ALL IMPLIED
          WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, TITLE, AND
          NON-INFRINGEMENT. WE DO NOT WARRANT THAT THE SERVICE WILL BE UNINTERRUPTED, ERROR-FREE,
          SECURE, OR FREE OF VIRUSES OR OTHER HARMFUL COMPONENTS.
        </p>

        <h2>11. Limitation of Liability</h2>
        <p>
          TO THE MAXIMUM EXTENT PERMITTED BY APPLICABLE LAW, IN NO EVENT SHALL ROOTED, ITS
          AFFILIATES, OFFICERS, DIRECTORS, EMPLOYEES, OR AGENTS BE LIABLE FOR ANY INDIRECT,
          INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES, OR ANY LOSS OF PROFITS, DATA,
          USE, GOODWILL, OR OTHER INTANGIBLE LOSSES, RESULTING FROM:
        </p>
        <ul>
          <li>YOUR ACCESS TO, USE OF, OR INABILITY TO USE THE SERVICE;</li>
          <li>ANY CONDUCT OR CONTENT OF ANY USER OR THIRD PARTY ON THE SERVICE;</li>
          <li>UNAUTHORIZED ACCESS TO OR ALTERATION OF YOUR DATA;</li>
          <li>ANY IN-PERSON MEETING OR INTERACTION WITH ANOTHER USER.</li>
        </ul>
        <p>
          OUR TOTAL AGGREGATE LIABILITY SHALL NOT EXCEED THE GREATER OF (A) THE AMOUNT YOU PAID TO
          ROOTED IN THE TWELVE (12) MONTHS PRECEDING THE CLAIM, OR (B) ONE HUNDRED US DOLLARS
          ($100.00).
        </p>

        <h2>12. Indemnification</h2>
        <p>
          You agree to indemnify, defend, and hold harmless Rooted and its officers, directors,
          employees, and agents from any claims, liabilities, damages, losses, and expenses
          (including reasonable attorneys' fees) arising out of or related to: (a) your use of the
          service; (b) your violation of these Terms; (c) your violation of any rights of another
          person or entity; or (d) any content you post on the service.
        </p>

        <h2>13. Dispute Resolution and Arbitration</h2>
        <p>
          <strong>For users in the United States:</strong> You and Rooted agree that any dispute,
          claim, or controversy arising out of or relating to these Terms or the service shall be
          resolved through binding individual arbitration, rather than in court, except that either
          party may seek equitable relief in court for infringement or misuse of intellectual property
          rights.
        </p>
        <p>
          <strong>CLASS ACTION WAIVER:</strong> YOU AND ROOTED AGREE THAT EACH MAY BRING CLAIMS
          AGAINST THE OTHER ONLY IN YOUR OR ITS INDIVIDUAL CAPACITY, AND NOT AS A PLAINTIFF OR CLASS
          MEMBER IN ANY PURPORTED CLASS, CONSOLIDATED, OR REPRESENTATIVE ACTION.
        </p>
        <p>
          <strong>30-Day Opt-Out:</strong> You may opt out of this arbitration agreement by sending
          written notice to toluwase.dev@gmail.com within 30 days of first accepting these Terms. Your
          notice must include your name, Telegram username, and a clear statement that you wish to opt
          out of the arbitration clause.
        </p>
        <p>
          For users outside the United States, disputes shall be resolved in accordance with the laws
          of your jurisdiction, subject to the governing law provision below.
        </p>

        <h2>14. Termination</h2>
        <p>
          We reserve the right to suspend or terminate your account at any time, for any reason or no
          reason, with or without notice. Grounds for termination include, but are not limited to:
        </p>
        <ul>
          <li>Violation of these Terms or our Community Guidelines.</li>
          <li>Conduct that we determine is harmful to other users, the service, or third parties.</li>
          <li>Fraudulent, abusive, or illegal activity.</li>
          <li>Failure to meet eligibility requirements.</li>
        </ul>
        <p>
          You may delete your account at any time through the app settings. Upon termination, your
          right to use the service ceases immediately.
        </p>

        <h2>15. Governing Law</h2>
        <p>
          These Terms shall be governed by and construed in accordance with the laws of the State of
          Delaware, United States, without regard to conflict of law principles, except where
          overridden by mandatory local consumer protection laws (such as GDPR for EU residents or
          NDPA for Nigerian residents).
        </p>

        <h2>16. Changes to These Terms</h2>
        <p>
          We may update these Terms from time to time. When we make material changes, we will notify
          you through the app or via Telegram. Your continued use of Rooted after such changes
          constitutes your acceptance of the updated Terms. If you do not agree to the changes, you
          must stop using the service and delete your account.
        </p>

        <h2>17. Contact Us</h2>
        <p>
          If you have questions about these Terms of Service, please contact us at:{' '}
          <a href="mailto:toluwase.dev@gmail.com">toluwase.dev@gmail.com</a>
        </p>
      </div>
    </div>
  )
}
