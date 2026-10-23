import { useEffect, useState } from "react"

export const CONFIRM_COUNTDOWN = 3

export function useConfirmCountdown(
  open: boolean,
  seconds = CONFIRM_COUNTDOWN
): number {
  const [remaining, setRemaining] = useState(0)

  useEffect(() => {
    if (!open) {
      setRemaining(0)
      return
    }
    setRemaining(seconds)
    const timer = setInterval(() => {
      setRemaining((value) => {
        if (value <= 1) {
          clearInterval(timer)
          return 0
        }
        return value - 1
      })
    }, 1000)
    return () => clearInterval(timer)
  }, [open, seconds])

  return remaining
}