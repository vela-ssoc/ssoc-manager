import { OTPInput, type SlotProps } from "input-otp"

import { cn } from "@/lib/utils"

export function InputOTP({
  id,
  value,
  onChange,
  disabled,
  "aria-label": ariaLabel,
}: {
  id?: string
  value: string
  onChange: (value: string) => void
  disabled?: boolean
  "aria-label"?: string
}) {
  return (
    <OTPInput
      maxLength={6}
      id={id}
      value={value}
      onChange={onChange}
      inputMode="numeric"
      pattern="[0-9]*"
      autoComplete="one-time-code"
      aria-label={ariaLabel}
      disabled={disabled}
      containerClassName="flex w-full items-center justify-center"
      render={({ slots }) => (
        <div className="flex items-center justify-center gap-2">
          {slots.map((slot, index) => (
            <InputOTPSlot key={index} {...slot} disabled={disabled} />
          ))}
        </div>
      )}
    />
  )
}

function InputOTPSlot({
  char,
  isActive,
  disabled,
}: SlotProps & { disabled?: boolean }) {
  return (
    <div
      className={cn(
        "bg-background text-foreground flex size-10 items-center justify-center rounded-md border font-mono text-lg tabular-nums transition-colors",
        isActive && "border-ring ring-2 ring-ring/30",
        disabled && "opacity-50"
      )}
    >
      {char}
    </div>
  )
}
