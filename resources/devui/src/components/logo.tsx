export function Logo({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 16 16"
      className={className}
      aria-hidden="true"
    >
      <g fill="none" fillRule="evenodd" transform="translate(2 2)">
        <rect width="5.5" height="5.5" fill="#499C54" />
        <rect width="5.5" height="5.5" x="6.5" fill="#F0A732" />
        <rect width="5.5" height="5.5" y="6.5" fill="#3592C4" />
        <rect width="5.5" height="5.5" x="6.5" y="6.5" fill="#C75450" />
      </g>
    </svg>
  )
}

// SSOC 四个字母与 favicon 四色方块一一对应（按从左到右、从上到下的阅读顺序）。
const SSOC_LETTERS = [
  { char: "S", color: "#499C54" }, // 绿
  { char: "S", color: "#F0A732" }, // 橙
  { char: "O", color: "#3592C4" }, // 蓝
  { char: "C", color: "#C75450" }, // 红
] as const

export function SSOCText({ className }: { className?: string }) {
  return (
    <span className={className} style={{ userSelect: "none" }}>
      {SSOC_LETTERS.map((letter, i) => (
        <span
          key={i}
          className="sso-letter"
          style={{ animationDelay: `${-i}s` }}
        >
          {letter.char}
        </span>
      ))}
    </span>
  )
}