import { useEffect, useState } from "react"
import { loadDeviceRules } from "../api/rules"
import type { Rule } from "../types/Types"

// Device Rules hook. 
// Runs once on startup returns rules list, error state, and functions to update the Rules list.
// Those functions are called in parent component(s) as callback functions once an API
// call succeeds

// Loads a device's rules once and exposes helpers to keep the local list in sync
// after the create / update / delete API calls succeed
export function useDeviceRules(deviceId: string) {
    const [rules, setRules] = useState<Rule[]>([])
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        let cancelled = false

        // Try to load device rules.
        async function load() {
            try {
                const result = await loadDeviceRules({ deviceId })
                if (!cancelled) {
                    setRules(result)
                    setError(null)
                }
            } catch (err) {
                if (!cancelled) setError(err instanceof Error ? err.message : "Failed to load rules")
            }
        }

        // Load once at render
        void load()
        // Return cleanup with cancelled = true to cancel running load on re-render
        return () => { cancelled = true }
    }, [deviceId]) // Reload if the page switches to a different device.

    // Return list of rules, error, and helper functions to use as callbacks once API calls succeed.
    return {
        rules,
        error,
        addRule: (rule: Rule) => setRules((current) => [...current, rule]),
        replaceRule: (updated: Rule) =>
            setRules((current) => current.map((r) => (r.id === updated.id ? updated : r))),
        removeRule: (ruleId: string) =>
            setRules((current) => current.filter((r) => r.id !== ruleId)),
    }
}
