import { useState } from "react"
import type { RuleDraft } from "../components/validation/rule"

// Holds the form values for a rule being created or edited.
// `setField` updates one field; `reset` restores the starting values (used by Cancel).
export function useRuleDraft(initial: () => RuleDraft) {
    // Maintain draft state. Initial will either be an empty draft for creation or
    // a draft of the current rule for editing.
    const [draft, setDraft] = useState<RuleDraft>(initial)

    // Update one field.
    // Can be read as: setField on K where K is a key of RuleDraft. Pass in the key
    // as text, and the value as whatever the expected type of K is.
    function setField<K extends keyof RuleDraft>(key: K, value: RuleDraft[K]) {
        // Set just key value, keep the rest
        setDraft((d) => ({ ...d, [key]: value }))
    }

    // Reset all draft values to initial.
    function reset() {
        setDraft(initial())
    }

    // Return draft and the functions
    return { draft, setField, reset }
}

// Export generic SetRuleField for use with any field defined by RuleDraft
export type SetRuleField = ReturnType<typeof useRuleDraft>["setField"]
