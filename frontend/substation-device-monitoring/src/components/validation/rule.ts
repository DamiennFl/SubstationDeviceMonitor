import { Comparison, Measurement, RuleType, State, type Rule, type CreateRuleInput, type UpdateRuleInput } from "../../types/Types"

// Draft object matching Rule. Used for temporary draft state when creating or editing rule
export type RuleDraft = {
    name: string
    enabled: boolean
    type: RuleType
    measurement: Measurement
    comparison: Comparison | undefined
    limit: string
    lowerLimit: string
    upperLimit: string
    deadband: string
    tripState: State | undefined
}

// Starting values for the create form. Type, measurement and trip state are chosen so the
// default draft is itself valid (status rules must use breakerStatus)
export const emptyRuleDraft: RuleDraft = {
    name: "",
    enabled: true,
    type: RuleType.Status,
    measurement: Measurement.BreakerStatus,
    comparison: Comparison.GreaterThan,
    limit: "0",
    lowerLimit: "0",
    upperLimit: "0",
    deadband: "0",
    tripState: State.Open,
}

// Conversion Types

// Converts a validated draft into the API payload for creating a rule.
export function draftToCreateInput(draft: RuleDraft): CreateRuleInput {
    return {
        name: draft.name.trim(),
        enabled: draft.enabled,
        type: draft.type,
        measurement: draft.measurement,
        comparison: draft.comparison,
        limit: Number(draft.limit),
        lowerLimit: Number(draft.lowerLimit),
        upperLimit: Number(draft.upperLimit),
        deadband: Number(draft.deadband),
        tripState: draft.tripState,
    }
}

// Same as above for updates. The rule type can't be changed, so it is left out.
export function draftToUpdateInput(draft: RuleDraft): UpdateRuleInput {
    const input: UpdateRuleInput & { type?: RuleType } = draftToCreateInput(draft)
    delete input.type
    return input
}

// Create draft from Rule. Used for editing Rule
export function draftFromRule(rule: Rule): RuleDraft {
    return {
        name: rule.name,
        enabled: rule.enabled,
        type: rule.type,
        measurement: rule.measurement,
        comparison: rule.comparison,
        limit: String(rule.limit ?? 0),
        lowerLimit: String(rule.lowerLimit ?? 0),
        upperLimit: String(rule.upperLimit ?? 0),
        deadband: String(rule.deadband ?? 0),
        tripState: rule.tripState,
    }
}


// Returns an error message if the edited rule is invalid, otherwise null.
// Required fields depend on rule type, mirroring the fields the backend evaluates.
export function validateRule(draft: RuleDraft): string | null {
    // Just don't allow empty name: duplicate names are handled by backend.
    if (draft.name.trim() === "") {
        return "Name is required"
    }
    if (!Object.values(RuleType).includes(draft.type)) {
        return "Invalid rule type"
    }
    if (!Object.values(Measurement).includes(draft.measurement)) {
        return "Invalid measurement"
    }

    // Number("") is 0, so check for blank strings explicitly before converting.
    const parse = (value: string) => (value.trim() === "" ? NaN : Number(value))
    const deadbandNum = parse(String(draft.deadband))
    if (!Number.isFinite(deadbandNum) || deadbandNum < 0) {
        return "Deadband must be a number that is 0 or greater"
    }

    switch (draft.type) {
        case RuleType.Status:
            // Status rules watch breaker state and only need a trip state.
            if (draft.measurement !== Measurement.BreakerStatus) {
                return "Status rules must use the breakerStatus measurement"
            }
            if (draft.tripState === undefined || !Object.values(State).includes(draft.tripState)) {
                return "Trip state is required for status rules"
            }
            break
        case RuleType.Threshold:
            // Threshold validation: Must not check BreakerStatus, Comparison must be defined,
            // Limit must be a valid number.
            if (draft.measurement === Measurement.BreakerStatus) {
                return "Threshold rules cannot use the breakerStatus measurement"
            }
            if (draft.comparison === undefined || !Object.values(Comparison).includes(draft.comparison)) {
                return "Comparison is required for threshold rules"
            }
            if (!Number.isFinite(parse(draft.limit))) {
                return "Limit must be a valid number"
            }
            break
        case RuleType.Band: {
            // Band validation: Must not check BreakerStatus, lower and upper limit must be defined,
            // lower must be less than upper.
            if (draft.measurement === Measurement.BreakerStatus) {
                return "Band rules cannot use the breakerStatus measurement"
            }
            const lower = parse(draft.lowerLimit)
            const upper = parse(draft.upperLimit)
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