import { RuleType, Measurement, Comparison, State } from "../../types/Types"
import type { RuleDraft } from "../validation/rule"
import type { SetRuleField } from "../../hooks/useRuleDraft"

// The input fields shared by the create row and the edit row. It owns no state:
// the parent passes the current draft and a setter, so the parent decides what to do on submit.
// `canEditType` is true when creating; an existing rule's type can't be changed.
export function RuleForm({
    draft,
    setField,
    canEditType,
}: {
    draft: RuleDraft
    setField: SetRuleField
    canEditType: boolean
}) {
    return (
        // Define labels
        <>
            <label className="device-create-field">
                Name
                <input
                    type="text"
                    value={draft.name}
                    onChange={(event) => setField("name", event.target.value)}
                    placeholder="Rule name"
                />
            </label>
            <label className="device-create-field">
                Enabled
                <select
                    value={draft.enabled ? "true" : "false"}
                    onChange={(event) => setField("enabled", event.target.value === "true")}
                >
                    <option value="true">True</option>
                    <option value="false">False</option>
                </select>
            </label>
            {canEditType && (
                <EnumSelect
                    label="Type"
                    options={RuleType}
                    value={draft.type}
                    onChange={(v) => setField("type", v)}
                />
            )}
            <EnumSelect
                label="Trip State"
                options={State}
                value={draft.tripState}
                onChange={(v) => setField("tripState", v)}
            />
            <EnumSelect
                label="Measurement"
                options={Measurement}
                value={draft.measurement}
                onChange={(v) => setField("measurement", v)}
            />
            <EnumSelect
                label="Comparison"
                options={Comparison}
                value={draft.comparison}
                onChange={(v) => setField("comparison", v)}
            />
            <NumberField label="Limit" value={draft.limit} onChange={(v) => setField("limit", v)} />
            <NumberField label="Lower Limit" value={draft.lowerLimit} onChange={(v) => setField("lowerLimit", v)} />
            <NumberField label="Upper Limit" value={draft.upperLimit} onChange={(v) => setField("upperLimit", v)} />
            <NumberField label="Deadband" value={draft.deadband} onChange={(v) => setField("deadband", v)} />
        </>
    )
}

// A <select> whose options come from one of the `as const` enum objects in Types.ts
// (RuleType, State, Measurement, Comparison).
function EnumSelect<T extends string>({
    label,
    options,
    value,
    onChange,
}: {
    label: string
    options: Record<string, T>
    value: T | undefined
    onChange: (value: T) => void
}) {
    return (
        <label className="device-create-field">
            {label}
            <select value={value} onChange={(event) => onChange(event.target.value as T)}>
                {Object.values(options).map((t) => (
                    <option key={t} value={t}>{t}</option>
                ))}
            </select>
        </label>
    )
}

// Numeric input. The value stays a string (so the user can type "-" or "1.") and is converted on submit.
function NumberField({
    label,
    value,
    onChange,
}: {
    label: string
    value: string
    onChange: (value: string) => void
}) {
    return (
        <label className="device-create-field">
            {label}
            <input
                type="number"
                step="any"
                value={value}
                onChange={(event) => onChange(event.target.value)}
                placeholder="0"
            />
        </label>
    )
}
