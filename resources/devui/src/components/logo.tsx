const PUBLIC_BASE = import.meta.env.BASE_URL.endsWith("/")
  ? import.meta.env.BASE_URL
  : `${import.meta.env.BASE_URL}/`

export function Logo({ className }: { className?: string }) {
  return (
    <span className={className} aria-hidden="true">
      <img
        src={`${PUBLIC_BASE}favicon.svg`}
        alt=""
        className="size-full dark:hidden"
        draggable={false}
      />
      <img
        src={`${PUBLIC_BASE}favicon_dark.svg`}
        alt=""
        className="hidden size-full dark:block"
        draggable={false}
      />
    </span>
  )
}

const SSOC_LETTERS = ["S", "S", "O", "C"] as const

export function SSOCText({ className }: { className?: string }) {
  return (
    <span className={className} style={{ userSelect: "none" }}>
      {SSOC_LETTERS.map((letter, i) => (
        <span
          key={i}
          className="sso-letter"
          style={{ animationDelay: `${-i}s` }}
        >
          {letter}
        </span>
      ))}
    </span>
  )
}
