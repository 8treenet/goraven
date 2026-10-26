export function IconGrid({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" className={className}>
      <rect width="8" height="8" x="3" y="3" rx="2" />
      <rect width="8" height="8" x="13" y="3" rx="2" />
      <rect width="8" height="8" x="3" y="13" rx="2" />
      <rect width="8" height="8" x="13" y="13" rx="2" />
    </svg>
  )
}

export function IconFolderFilled({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" className={className}>
      <path d="m6 14 1.5-2.9A2 2 0 0 1 9.24 10H20a2 2 0 0 1 1.94 2.5l-1.54 6a2 2 0 0 1-1.95 1.5H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h3.9a2 2 0 0 1 1.69.9l.81 1.2a2 2 0 0 0 1.67.9H18a2 2 0 0 1 2 2v2" />
    </svg>
  )
}

export function IconWrench({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" className={className}>
      <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" />
    </svg>
  )
}

export function IconBot({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className={className}>
      <rect x="4" y="8" width="16" height="12" rx="2" fill="currentColor" />
      <path d="M12 8V4" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" />
      <circle cx="12" cy="3" r="1.5" fill="currentColor" />
      <rect x="8.8" y="12" width="2.4" height="3.4" rx="1.1" fill="#fff" />
      <rect x="12.8" y="12" width="2.4" height="3.4" rx="1.1" fill="#fff" />
    </svg>
  )
}

export function IconInfo({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className={className}>
      <circle cx="12" cy="12" r="10" fill="currentColor" />
      <rect x="11" y="11" width="2" height="6" rx="1" fill="#fff" />
      <circle cx="12" cy="7" r="1.5" fill="#fff" />
    </svg>
  )
}

export function IconUsers({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className={className}>
      <circle cx="9" cy="8" r="3.4" fill="currentColor" />
      <path d="M2.5 20c0-3.6 2.9-5.8 6.5-5.8s6.5 2.2 6.5 5.8v.2a.3.3 0 0 1-.3.3H2.8a.3.3 0 0 1-.3-.3z" fill="currentColor" />
      <circle cx="17" cy="9" r="2.6" fill="currentColor" />
      <path d="M14.5 14.5c1 .5 1.9 1.3 2.5 2.4.6 1 .9 2 .9 3.1h-3.4z" fill="currentColor" />
    </svg>
  )
}

export function IconSettings({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill-rule="evenodd" aria-hidden="true" className={className}>
      <path
        fill="currentColor"
        d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"
      />
      <circle cx="12" cy="12" r="3.2" fill="#fff" />
    </svg>
  )
}

export function IconShare({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className={className}>
      <circle cx="18" cy="5" r="3" fill="currentColor" />
      <circle cx="6" cy="12" r="3" fill="currentColor" />
      <circle cx="18" cy="19" r="3" fill="currentColor" />
      <path
        d="m8.59 13.51 6.83 3.98M15.41 6.51l-6.82 3.98"
        stroke="currentColor"
        strokeWidth="2"
        fill="none"
        strokeLinecap="round"
      />
    </svg>
  )
}

export function IconCpu({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className={className}>
      <rect x="6" y="6" width="12" height="12" rx="2" fill="currentColor" />
      <rect x="10" y="10" width="4" height="4" rx="1" fill="#fff" />
      <path
        d="M9 2v2M15 2v2M9 20v2M15 20v2M2 9h2M2 15h2M20 9h2M20 15h2"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        fill="none"
      />
    </svg>
  )
}

export function IconPlug({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className={className}>
      <rect x="6" y="8" width="12" height="9" rx="2" fill="currentColor" />
      <path d="M9 8V3M15 8V3" stroke="currentColor" strokeWidth="2" strokeLinecap="round" fill="none" />
      <path d="M12 17v4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" fill="none" />
    </svg>
  )
}

export function IconAutomation({ className }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className={className}>
      <circle cx="12" cy="13" r="8.5" fill="currentColor" />
      <path d="M12 8.5V13l3 1.8" stroke="#fff" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" fill="none" />
      <path d="M12 1.5v2M19.8 4.2l-1.4 1.4M4.2 4.2l1.4 1.4" stroke="currentColor" strokeWidth="2" strokeLinecap="round" fill="none" />
    </svg>
  )
}
