import { useEffect, useState } from "react"

// Hook to display message for a specified amount of time before
// returning to null state, clearing the message.
export function useTempMessage(ms = 5000) {
    const [message, setMessage] = useState<string | null>(null)

    useEffect(() => {
        if (message === null) return

        const timer = setTimeout(() => setMessage(null), ms)

        return () => clearTimeout(timer)
    }, [message, ms])

    return [message, setMessage] as const
}