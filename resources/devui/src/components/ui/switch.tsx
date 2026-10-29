import { cn } from "@/lib/utils"

export function Switch({
  checked,
  onCheckedChange,
  label,
  disabled = false,
  className,
}: {
  checked: boolean
  onCheckedChange: (checked: boolean) => void
  label: string
  disabled?: boolean
  className?: string
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-label={label}
      aria-checked={checked}
      data-state={checked ? "checked" : "unchecked"}
      disabled={disabled}
      className={cn(
        "border-input data-[state=checked]:bg-primary data-[state=unchecked]:bg-input focus-visible:ring-ring inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border transition-colors focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
      onClick={() => onCheckedChange(!checked)}
    >
      <span
        className="bg-background pointer-events-none block size-4 rounded-full shadow-sm transition-transform data-[state=checked]:translate-x-4 data-[state=unchecked]:translate-x-0"
        data-state={checked ? "checked" : "unchecked"}
      />
    </button>
  )
}