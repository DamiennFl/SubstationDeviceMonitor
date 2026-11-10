import { useEffect, useRef, useState } from "react"

// Calls `fetcher` immediately and then every `intervalMs`, keeping the latest result and error.
// `key` identifies what is being fetched (e.g. a device ID): when it changes, polling restarts.
// `data` stays null until the first success. A failed poll sets `error` but keeps the last `data`.
export function usePolling<T>(
    fetcher: () => Promise<T>,
    intervalMs: number,
    key: string | undefined, // Undefined makes usePolling safe in early return cases
    fallbackError = "Request failed",
) {
    const [data, setData] = useState<T | null>(null)
    const [error, setError] = useState<string | null>(null)

    // Keep the newest fetcher in a ref so the effect below doesn't restart on every render
    // (an inline arrow function would be a new object each time).
    const fetcherRef = useRef(fetcher)
    // Update fetcher ref each re-render, to prevent fetcher from being stale.
    useEffect(() => { fetcherRef.current = fetcher })

    // 
    useEffect(() => {
        if (!key) return
        let cancelled = false

        async function poll() {
            try {
                const result = await fetcherRef.current()
                if (!cancelled) {
                    setData(result)
                    setError(null)
                }
            } catch (err) {
                if (!cancelled) setError(err instanceof Error ? err.message : fallbackError)
            }
        }

        // Call poll initially, not in interval so it is immediate
        void poll()
        // Each  tick calls fetcherRef.current rather than fetcher directly. usePolling
        // callers pass new inline arrow functions every render, so having fetcher in this effect's
        // deps would restart the interval and poll() on every render, looping forever. Reading
        // the fetcher function from ref lets the inerval run uninterrupted while still calling the right
        // function object.
        const intervalId = setInterval(() => void poll(), intervalMs)

        // Cleanup: stop the timer and ignore any in-flight response (e.g. navigating to another device).
        return () => {
            cancelled = true
            clearInterval(intervalId)
        }
    }, [key, intervalMs, fallbackError])

    return { data, error }
}
