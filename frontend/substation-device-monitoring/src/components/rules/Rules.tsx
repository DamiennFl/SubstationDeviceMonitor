import { createDeviceRule, updateDeviceRule, deleteDeviceRule, loadDeviceRules } from "../../api/rules";
import {
    RuleType, 
    Measurement, 
    Comparison, 
    State } from "../../types/Types";
import type { Device, Rule } from "../../types/Types";
import { useState, useEffect } from "react";


function DeviceRules( {device }: {device: Device}) {
    const [ruleList, setRuleList] = useState<Rule[]>([])
    const [ruleError, setRuleError] = useState<string | null>(null)

    const [showCreateRuleRow, setShowCreateRuleRow] = useState(false) // Define bool for showing creation row
    
    // Wrap async in useEffect, otherwise infinite loop because handleLoading calls setRules
    // which triggers a re-render, which would cause handleLoading to call again.
    useEffect(() =>  {
        async function handleLoading() {
            try {
                const ruleList = await loadDeviceRules({deviceId: device.deviceId })
                setRuleList(ruleList)
                setRuleError(null)
            } catch (error) {
                setRuleError(error instanceof Error ? error.message : "Failed to load devices")
            }
        }

        void handleLoading()
    }, [setRuleList, setRuleError])
    
        return (
            <div>
                <RuleCreateButton setShowCreateRuleRow={setShowCreateRuleRow} />
                <RuleRowHeader />
                {showCreateRuleRow && ( // Conditionally render device creation row
                    <RuleCreateRow
                        deviceId={device.deviceId}
                        // Append the device returned by the API via onCreated callback prop, then hide the create row
                        onCreated={(rule) => {
                            setRuleList((currentRules) => [...currentRules, rule]) // Append device to currentDevices
                            setShowCreateRuleRow(false) //hide row via setter
                        }}
                        onCancel={() => setShowCreateRuleRow(false)} // On cancel, just hide row
                    />
                )}
                <RuleList 
                    rules={ruleList} 
                    error={ruleError}
                    onRuleUpdated={(updated) =>
                    setRuleList((current) =>
                        current.map((r) => (r.id === updated.id ? updated : r)),
                    )}
                    // Filter out deleted device on deletion while updating devices list
                    onRuleDeleted={(deleted) => 
                        setRuleList((current) => current.filter((r) => r.id !== deleted))
                    }
                />
            </div>
        )

}

function RuleList({
    rules,
    error,
    onRuleUpdated,
    onRuleDeleted,
}: {
    rules: Rule[]
    error: string | null
    onRuleUpdated: (rule: Rule) => void
    onRuleDeleted: (ruleId: string) => void
}) {
    // Map all rules in rule rows, if no error. Otherwise return empty div with error.
    return (
        error ? <div>{error}</div> :
        <div className="device-table">
            {rules.map((rule) => (
                <RuleRow
                    key={rule.id}
                    rule={rule}
                    onUpdated={onRuleUpdated}
                    onDeleted={onRuleDeleted}
                />
            ))}
        </div>
    )
}

function RuleCreateButton({
    setShowCreateRuleRow,
}: {
    setShowCreateRuleRow: (show: boolean) => void
}) {
    // Return button with onClick trigger
    return (
        <div className="device-page-button-container">
            <button
                className="device-create-button"
                onClick={() => setShowCreateRuleRow(true)}
            >
                Create Rule
            </button>
        </div>
    )
}

function RuleCreateRow({
    deviceId,
    onCreated,
    onCancel,
}: {
    deviceId: string
    onCreated: (rule: Rule) => void
    onCancel: () => void
}) {
    // Define own function states
    const [name, setName] = useState("")
    const [enabled, setEnabled] = useState(true)
    const [ruleType, setRuleType] = useState<RuleType>("status")
    const [measurement, setMeasurement] = useState<Measurement>("currentA")
    const [comparison, setComparison] = useState<Comparison>("GT")
    const [limit, setLimit] = useState("0")
    const [lowerLimit, setLowerLimit] = useState("0")
    const [upperLimit, setUpperLimit] = useState("0")
    const [deadband, setDeadband] = useState("0")
    const [tripState, setTripState] = useState<State>("open")

    // Separate from page error, so a failed create only shows in this row instead of replacing the rule list.
    const [createError, setCreateError] = useState<string | null>(null)
    // Prevents double clicks from sending multiple POST requests.
    const [submitting, setSubmitting] = useState(false)

    // Validation and submission async function
    async function handleConfirm() {
        // Validate name. Just don't allow empty name: duplicate names are handled by backend.
        // TODO: Validate all fields
        if (name.trim() === "") {
            setCreateError("Name is required")
            return
        }

        // Block multiple submits, and set error to null
        setSubmitting(true)
        setCreateError(null)
        try {
            const createdRule = await createDeviceRule({  // await creating device
                deviceId: deviceId,
                input: {
                    name: name.trim(),
                    enabled: enabled,
                    type: ruleType,
                    measurement: measurement,
                    comparison: comparison,
                    limit: Number(limit),
                    lowerLimit: Number(lowerLimit),
                    upperLimit: Number(upperLimit),
                    deadband: Number(deadband),
                    tripState: tripState
                }
            })
            // If no error caught during rule creation, onCreated callback gets new rule
            onCreated(createdRule)
        } catch (error) { // Error caught, likely during rule creation.
            // set CreateError to error message or fall back to default.
            setCreateError(
                error instanceof Error ? error.message : "Failed to create rule",
            )
            setSubmitting(false)
        }
    }

    return (
        <div className="device-row device-create-row">
            <div className="device-info">
                <label className="device-create-field">
                    Name
                    <input
                        type="text"
                        value={name}
                        onChange={(event) => setName(event.target.value)}
                        placeholder="Rule name"
                    />
                </label>
                <label className="device-create-field">
                    Enabled
                    <select
                        value={enabled ? "true" : "false"}
                        onChange={(event) => setEnabled(event.target.value === "true")}
                    >
                        <option value="true">True</option>
                        <option value="false">False</option>
                    </select>
                </label>
                <label className="device-create-field">
                    Type
                    {/* Define selection based on constants defined in RuleType */}
                    <select
                        value={ruleType}
                        onChange={(event) => setRuleType(event.target.value as RuleType)}
                    >
                        {Object.values(RuleType).map((t) => (
                            <option key={t} value={t}>{t}</option>
                        ))}
                    </select>
                </label>
                <label className="device-create-field">
                    Trip State
                    {/* Define selection based on options in State */}
                    <select
                        value={tripState}
                        onChange={(event) => setTripState(event.target.value as State)}
                    >
                        {Object.values(State).map((t) => (
                            <option key={t} value={t}>{t}</option>
                        ))}
                    </select>
                </label>
                <label className="device-create-field">
                    Measurement
                    {/* Define selection based on options in Measurement */}
                    <select
                        value={measurement}
                        onChange={(event) => setMeasurement(event.target.value as Measurement)}
                    >
                        {Object.values(Measurement).map((t) => (
                            <option key={t} value={t}>{t}</option>
                        ))}
                    </select>
                </label>
                <label className="device-create-field">
                    Comparison
                    {/* Define selection based on options in Comparison */}
                    <select
                        value={comparison}
                        onChange={(event) => setComparison(event.target.value as Comparison)}
                    >
                        {Object.values(Comparison).map((t) => (
                            <option key={t} value={t}>{t}</option>
                        ))}
                    </select>
                </label>
                <label className="device-create-field">
                    Limit
                    <input
                        type="number"
                        step="any"
                        value={limit}
                        onChange={(event) => setLimit(event.target.value)}
                        placeholder="0"
                    />
                </label>
                <label className="device-create-field">
                    Lower Limit
                    <input
                        type="number"
                        step="any"
                        value={lowerLimit}
                        onChange={(event) => setLowerLimit(event.target.value)}
                        placeholder="0"
                    />
                </label>
                <label className="device-create-field">
                    Upper Limit
                    <input
                        type="number"
                        step="any"
                        value={upperLimit}
                        onChange={(event) => setUpperLimit(event.target.value)}
                        placeholder="0"
                    />
                </label>
                <label className="device-create-field">
                    Deadband
                    <input
                        type="number"
                        step="any"
                        value={deadband}
                        onChange={(event) => setDeadband(event.target.value)}
                        placeholder="0"
                    />
                </label>
                {createError !== null && (
                    <div className="device-create-error" role="alert">
                        {createError}
                    </div>
                )}
            </div>
            {/* RuleCreateButtons handle creating and cancelling. Disable buttons when submitting.*/}
            <RuleCreateButtons
                onConfirm={() => void handleConfirm()}
                onCancel={onCancel}
                disabled={submitting}
            />
        </div>
    )
}
// Represents one Rule
function RuleRow({ 
    rule,
    onUpdated,
    onDeleted,
}: {
    rule: Rule,
    onUpdated: (rule: Rule) => void
    onDeleted: (ruleId: string) => void
}) {
    const [editing, setEditing] = useState(false)
    const [name, setName] = useState(rule.name)
    const [enabled, setEnabled] = useState(rule.enabled)
    const [ruleType] = useState<RuleType>(rule.type)
    const [measurement, setMeasurement] = useState<Measurement>(rule.measurement)
    const [comparison, setComparison] = useState<Comparison | undefined>(rule.comparison)
    // Numeric inputs are kept as strings (matching the create row) and converted on validation/submit.
    const [limit, setLimit] = useState(String(rule.limit ?? 0))
    const [lowerLimit, setLowerLimit] = useState(String(rule.lowerLimit ?? 0))
    const [upperLimit, setUpperLimit] = useState(String(rule.upperLimit ?? 0))
    const [deadband, setDeadband] = useState(String(rule.deadband ?? 0))
    const [tripState, setTripState] = useState<State | undefined>(rule.tripState)
    const [updateError, setUpdateError] = useState<string | null>(null)
    const [deleteError, setDeleteError] = useState<string | null>(null)
    const [submitting, setSubmitting] = useState(false)

    // Returns an error message if the edited rule is invalid, otherwise null.
    // Required fields depend on rule type, mirroring the fields the backend evaluates.
    function validateRule(): string | null {
        // Just don't allow empty name: duplicate names are handled by backend.
        if (name.trim() === "") {
            return "Name is required"
        }
        if (!Object.values(RuleType).includes(ruleType)) {
            return "Invalid rule type"
        }
        if (!Object.values(Measurement).includes(measurement)) {
            return "Invalid measurement"
        }

        // Number("") is 0, so check for blank strings explicitly before converting.
        const parse = (value: string) => (value.trim() === "" ? NaN : Number(value))
        const deadbandNum = parse(deadband)
        if (!Number.isFinite(deadbandNum) || deadbandNum < 0) {
            return "Deadband must be a number that is 0 or greater"
        }

        switch (ruleType) {
            case RuleType.Status:
                // Status rules watch breaker state and only need a trip state.
                if (measurement !== Measurement.BreakerStatus) {
                    return "Status rules must use the breakerStatus measurement"
                }
                if (tripState === undefined || !Object.values(State).includes(tripState)) {
                    return "Trip state is required for status rules"
                }
                break
            case RuleType.Threshold:
                if (measurement === Measurement.BreakerStatus) {
                    return "Threshold rules cannot use the breakerStatus measurement"
                }
                if (comparison === undefined || !Object.values(Comparison).includes(comparison)) {
                    return "Comparison is required for threshold rules"
                }
                if (!Number.isFinite(parse(limit))) {
                    return "Limit must be a valid number"
                }
                break
            case RuleType.Band: {
                if (measurement === Measurement.BreakerStatus) {
                    return "Band rules cannot use the breakerStatus measurement"
                }
                const lower = parse(lowerLimit)
                const upper = parse(upperLimit)
                if (!Number.isFinite(lower) || !Number.isFinite(upper)) {
                    return "Lower and upper limits must be valid numbers"
                }
                if (lower >= upper) {
                    return "Lower limit must be less than upper limit"
                }
                break
            }
        }
        return null
    }

    // Validation and submission async function
    async function handleUpdate() {
        const validationError = validateRule()
        if (validationError !== null) {
            setUpdateError(validationError)
            return
        }

        // Block multiple submits, and set error to null
        setSubmitting(true)
        setUpdateError(null)
        try {
            // Update rule
            const editedRule = await updateDeviceRule({
                deviceId: rule.deviceId,
                ruleId: rule.id,
                input: {
                    name: name.trim(),
                    enabled: enabled,
                    measurement: measurement,
                    comparison: comparison,
                    limit: Number(limit),
                    lowerLimit: Number(lowerLimit),
                    upperLimit: Number(upperLimit),
                    deadband: Number(deadband),
                    tripState: tripState
                }
            })
            onUpdated(editedRule) // Pass rule back via callback
            setEditing(false)
        } catch (error) { // Error caught, likely during rule update.
            // set UpdateError to error message or fall back to default.
            setUpdateError(
                error instanceof Error ? error.message : "Failed to update rule",
            )
        } finally {
            setSubmitting(false)
        }
    }

    async function handleDelete() {
        setSubmitting(true)
        setDeleteError(null)
        try {
            await deleteDeviceRule({ deviceId: rule.deviceId, ruleId: rule.id })
            onDeleted(rule.id)
        } catch (error) {
            setDeleteError(
                error instanceof Error ? error.message : "Failed to delete rule",
            )
        }
        finally {
            setSubmitting(false)
        }
    }

   return editing ? (
        <div className="device-row">
            <div className="device-info">
                <label className="device-create-field">
                    Name
                    <input
                        type="text"
                        value={name}
                        onChange={(event) => setName(event.target.value)}
                        placeholder="Rule name"
                    />
                </label>
                <label className="device-create-field">
                    Enabled
                    <select
                        value={enabled ? "true" : "false"}
                        onChange={(event) => setEnabled(event.target.value === "true")}
                    >
                        <option value="true">True</option>
                        <option value="false">False</option>
                    </select>
                </label>
                <label className="device-create-field">
                    Trip State
                    {/* Define selection based on options in State */}
                    <select
                        value={tripState}
                        onChange={(event) => setTripState(event.target.value as State)}
                    >
                        {Object.values(State).map((t) => (
                            <option key={t} value={t}>{t}</option>
                        ))}
                    </select>
                </label>
                <label className="device-create-field">
                    Measurement
                    {/* Define selection based on options in Measurement */}
                    <select
                        value={measurement}
                        onChange={(event) => setMeasurement(event.target.value as Measurement)}
                    >
                        {Object.values(Measurement).map((t) => (
                            <option key={t} value={t}>{t}</option>
                        ))}
                    </select>
                </label>
                <label className="device-create-field">
                    Comparison
                    {/* Define selection based on options in Comparison */}
                    <select
                        value={comparison}
                        onChange={(event) => setComparison(event.target.value as Comparison)}
                    >
                        {Object.values(Comparison).map((t) => (
                            <option key={t} value={t}>{t}</option>
                        ))}
                    </select>
                </label>
                <label className="device-create-field">
                    Limit
                    <input
                        type="number"
                        step="any"
                        value={limit}
                        onChange={(event) => setLimit(event.target.value)}
                        placeholder="0"
                    />
                </label>
                <label className="device-create-field">
                    Lower Limit
                    <input
                        type="number"
                        step="any"
                        value={lowerLimit}
                        onChange={(event) => setLowerLimit(event.target.value)}
                        placeholder="0"
                    />
                </label>
                <label className="device-create-field">
                    Upper Limit
                    <input
                        type="number"
                        step="any"
                        value={upperLimit}
                        onChange={(event) => setUpperLimit(event.target.value)}
                        placeholder="0"
                    />
                </label>
                <label className="device-create-field">
                    Deadband
                    <input
                        type="number"
                        step="any"
                        value={deadband}
                        onChange={(event) => setDeadband(event.target.value)}
                        placeholder="0"
                    />
                </label>
                </div>
                {updateError !== null && (
                    <div className="device-create-error" role="alert">
                        {updateError}
                    </div>
                )}
            {/* RuleCreateButtons handle creating and cancelling. Disable buttons when submitting.*/}
            <RuleCreateButtons
                onConfirm={() => {
                    void handleUpdate()
                }}

                onCancel={() => {
                    // Reset all values. TODO: Is there better way to do this?
                    setName(rule.name)
                    setEnabled(rule.enabled)
                    setTripState(rule.tripState)
                    setMeasurement(rule.measurement)
                    setComparison(rule.comparison)
                    setLimit(String(rule.limit))
                    setLowerLimit(String(rule.lowerLimit))
                    setUpperLimit(String(rule.upperLimit))
                    setDeadband(String(rule.deadband))
                    setEditing(false)
                }}
                disabled={submitting}
            />
        </div>
    ) : (
        <div className="device-row">
            <RuleInfo rule={rule} />
            <RuleButtons
                onEdit={() => setEditing(true)}
                onDelete={() => void handleDelete()}/>
                {deleteError !== null && (
                    <div className="device-create-error" role="alert">
                        {deleteError}
                    </div>
                )}
        </div>
    );
}

// Rule Table header. Defined separately to allow rule creation row to render underneath.
function RuleRowHeader() {
    return (
        <div className="device-row device-header-row">
            <div className="device-info device-header-info">
                <div className="device-info__header">Name</div>
                <div className="device-info__header">Enabled</div>
                <div className="device-info__header">Type</div>
                <div className="device-info__header">Trip State</div>
                <div className="device-info__header">Measurement</div>
                <div className="device-info__header">Comparison</div>
                <div className="device-info__header">Limit</div>
                <div className="device-info__header">Upper Limit</div>
                <div className="device-info__header">Lower Limit</div>
                <div className="device-info__header">Deadband</div>
            </div>
            <div className="device-buttons" aria-hidden="true" />
        </div>
    )
}

// Displays the rule info.
function RuleInfo({ rule }: { rule: Rule }) {
    return (
    <div className="device-info">
        <div className="device-info__static">
            {rule.name}
        </div>
        <div className="device-info__static">
            {String(rule.enabled)}
        </div>
        <div className="device-info__status">
            {rule.type}
        </div>
        <div className="device-info__status">
            {rule.tripState}
        </div>
        <div className="device-info__status">
            {rule.measurement}
        </div>
        <div className="device-info__status">
            {rule.comparison}
        </div>
        <div className="device-info__status">
            {rule.limit}
        </div>
        <div className="device-info__status">
            {rule.upperLimit}
        </div>
        <div className="device-info__status">
            {rule.lowerLimit}
        </div>
        <div className="device-info__status">
            {rule.deadband}
        </div>
    </div>
    )
}

// RuleButtons defines right-aligned set of buttons for a rule row.
function RuleButtons({
    onEdit,
    onDelete
} : {
    onEdit: () => void
    onDelete: () => void
}
) {
    return (
        // In order: View Button, Edit Button, Delete Button
        <div className="device-buttons">
            <button
                className="device-button"
                onClick={onEdit}
                >
                <svg
                    width="12"
                    height="12"
                    viewBox="0 0 12 12"
                    fill="none"
                    aria-label="Edit rule"
                >
                    <path
                        d="M8.85 1.65a1.06 1.06 0 0 1 1.5 1.5L4.1 9.4l-2.1.6.6-2.1 6.25-6.25ZM7.9 2.6l1.5 1.5"
                        stroke="currentColor"
                        strokeWidth="1.2"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                    />
                </svg>
            </button>
            <button 
                className="device-button"
                onClick={onDelete}>
                <svg
                    width="12"
                    height="12"
                    viewBox="0 0 12 12"
                    fill="none"
                    aria-label="Delete rule"
                >
                    <path
                        d="m3 3 6 6M9 3 3 9"
                        stroke="currentColor"
                        strokeWidth="1.2"
                        strokeLinecap="round"
                    />
                </svg>
            </button>
        </div>
    )
}

// Define rule row buttons (Confirm and Cancel).
function RuleCreateButtons({
    onConfirm,
    onCancel,
    disabled,
}: {
    onConfirm: () => void
    onCancel: () => void
    disabled: boolean
}) {
    return (
        <div className="device-buttons">
            <button
                type="button"
                className="device-button"
                onClick={onConfirm}
                disabled={disabled}
                aria-label="Create rule"
            >
                <svg
                    width="12"
                    height="12"
                    viewBox="0 0 12 12"
                    fill="none"
                    aria-hidden="true"
                >
                    <path
                        d="m2.5 6.25 2.25 2.25L9.5 3.75"
                        stroke="currentColor"
                        strokeWidth="1.2"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                    />
                </svg>
            </button>
            <button
                type="button"
                className="device-button"
                onClick={onCancel}
                aria-label="Cancel creating rule"
            >
                <svg
                    width="12"
                    height="12"
                    viewBox="0 0 12 12"
                    fill="none"
                    aria-hidden="true"
                >
                    <path
                        d="m3 3 6 6M9 3 3 9"
                        stroke="currentColor"
                        strokeWidth="1.2"
                        strokeLinecap="round"
                    />
                </svg>
            </button>
        </div>
    )
}

export default DeviceRules  