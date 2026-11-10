import { createDeviceRule, updateDeviceRule, deleteDeviceRule } from "../../api/rules";
import type { Device, Rule } from "../../types/Types";
import { useState } from "react";
import { useDeviceRules } from "../../hooks/useDeviceRules";
import { useRuleDraft } from "../../hooks/useRuleDraft";
import {
    draftFromRule,
    draftToCreateInput,
    draftToUpdateInput,
    emptyRuleDraft,
    validateRule,
} from "../validation/rule";
import { RuleForm } from "./RuleForm";

function DeviceRules( {device }: {device: Device}) {
    // Loading and list bookkeeping live in the hook.
    const { rules, error, addRule, replaceRule, removeRule } = useDeviceRules(device.deviceId)
    const [showCreateRuleRow, setShowCreateRuleRow] = useState(false) // Define bool for showing creation row

    // Return Div containing Rules for Device page.
    return (
        <div>
            <RuleCreateButton setShowCreateRuleRow={setShowCreateRuleRow} />
            <RuleRowHeader />
            {showCreateRuleRow && ( // Conditionally render rule creation row
                <RuleCreateRow
                    deviceId={device.deviceId}
                    onCreated={(rule) => {
                        addRule(rule)
                        setShowCreateRuleRow(false) //hide row via setter
                    }}
                    onCancel={() => setShowCreateRuleRow(false)} // On cancel, just hide row
                />
            )}
            <RuleList
                rules={rules}
                error={error}
                onRuleUpdated={replaceRule}
                onRuleDeleted={removeRule}
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

// Conditionally rendered row which handles rule creation.
function RuleCreateRow({
    deviceId,
    onCreated,
    onCancel,
}: {
    deviceId: string
    onCreated: (rule: Rule) => void
    onCancel: () => void
}) {
    // Get hooks from useRuleDraft
    const { draft, setField } = useRuleDraft(() => emptyRuleDraft)
    // Separate from page error, so a failed create only shows in this row instead of replacing the rule list.
    const [createError, setCreateError] = useState<string | null>(null)
    // Prevents double clicks from sending multiple POST requests.
    const [submitting, setSubmitting] = useState(false)

    // For confirm button.
    async function handleConfirm() {
        // Validates the values inputted into forms
        const validationError = validateRule(draft)
        if (validationError !== null) {
            setCreateError(validationError)
            return
        }

        setSubmitting(true)
        setCreateError(null)
        // Try to send request
        try {
            const createdRule = await createDeviceRule({
                deviceId,
                input: draftToCreateInput(draft),
            })
            // The row is unmounted by the parent on success, so no need to reset submitting here.
            onCreated(createdRule)
        } catch (error) {
            setCreateError(
                error instanceof Error ? error.message : "Failed to create rule",
            )
            setSubmitting(false)
        }
    }

    // Return div with RuleForm and RuleCreateButtons.
    return (
        <div className="device-row device-create-row">
            <div className="device-info">
                <RuleForm draft={draft} setField={setField} canEditType />
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

// Represents one Rule, either read-only or being edited.
function RuleRow({
    rule,
    onUpdated,
    onDeleted,
}: {
    rule: Rule,
    onUpdated: (rule: Rule) => void
    onDeleted: (ruleId: string) => void
}) {
    // Editing state
    const [editing, setEditing] = useState(false)
    // Draft for rule editing, resetable
    const { draft, setField, reset } = useRuleDraft(() => draftFromRule(rule))
    // Error states
    const [updateError, setUpdateError] = useState<string | null>(null)
    const [deleteError, setDeleteError] = useState<string | null>(null)
    // To prevent duplicates
    const [submitting, setSubmitting] = useState(false)

    // For updates
    async function handleUpdate() {
        // Validate draft before submitting
        const validationError = validateRule(draft)
        if (validationError !== null) {
            setUpdateError(validationError)
            return
        }

        setSubmitting(true)
        setUpdateError(null)
        // Try and update device data via API
        try {
            const editedRule = await updateDeviceRule({
                deviceId: rule.deviceId,
                ruleId: rule.id,
                input: draftToUpdateInput(draft),
            })
            onUpdated(editedRule) // Pass rule back via callback
            setEditing(false)
        } catch (error) {
            setUpdateError(
                error instanceof Error ? error.message : "Failed to update rule",
            )
        } finally {
            setSubmitting(false)
        }
    }

    // For deleting
    async function handleDelete() {
        setSubmitting(true)
        setDeleteError(null)
        // Try and deleted Rule via API
        try {
            await deleteDeviceRule({ deviceId: rule.deviceId, ruleId: rule.id })
            onDeleted(rule.id)
        } catch (error) {
            setDeleteError(
                error instanceof Error ? error.message : "Failed to delete rule",
            )
        } finally {
            setSubmitting(false)
        }
    }

    // Return RuleForm with fields if editing, otherwise just presentational Rule row.
    return editing ? (
        <div className="device-row">
            <div className="device-info">
                <RuleForm draft={draft} setField={setField} canEditType={false} />
            </div>
            {updateError !== null && (
                <div className="device-create-error" role="alert">
                    {updateError}
                </div>
            )}
            {/* RuleCreateButtons handle confirming and cancelling. Disable buttons when submitting.*/}
            <RuleCreateButtons
                onConfirm={() => void handleUpdate()}
                onCancel={() => {
                    reset() // Discard edits, restoring the values from `rule`
                    setUpdateError(null)
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
// TODO: Update this to iterate Rule fields
function RuleInfo({ rule }: { rule: Rule }) {
    return (
    <div className="device-info">
        <div className="device-info__static">
            {rule.name}
        </div>
        <div className="device-info__static">
            {String(rule.enabled)}
        </div>
        <div className="device-info__static">
            {rule.type}
        </div>
        <div className="device-info__static">
            {rule.tripState}
        </div>
        <div className="device-info__static">
            {rule.measurement}
        </div>
        <div className="device-info__static">
            {rule.comparison}
        </div>
        <div className="device-info__static">
            {rule.limit}
        </div>
        <div className="device-info__static">
            {rule.upperLimit}
        </div>
        <div className="device-info__static">
            {rule.lowerLimit}
        </div>
        <div className="device-info__static">
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