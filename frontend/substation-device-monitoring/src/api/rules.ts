import type { Rule, CreateRuleInput, UpdateRuleInput } from "../types/Types"

export async function loadDeviceRules( {deviceId }: {deviceId: string }): Promise<Rule[]> {
        // Try to fetch devices
        const response = await fetch(`/api/devices/${deviceId}/rules`) // GET, await response
    
        // Handle response error
        if (!response.ok) {
            throw new Error(`Failed to load device rules: ${response.status}`)
        }
    
        // Create rule list from response. Matching properties are automatically mapped. Assumes JSON is properly formatted.
        // It should be, since we validate it before sending from backend
        const deviceList: Rule[] = await response.json()
        // Set list
        return deviceList
}


export async function createDeviceRule( {deviceId, input}: {deviceId: string, input: CreateRuleInput} ): Promise<Rule> {
    const response = await fetch(`/api/devices/${deviceId}/rules`, {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        body: JSON.stringify(input) // Sends all fields passed in CreateRuleInput. Already validated by function caller.
    })

    if (!response.ok) {
        const message = (await response.text()).trim()
        throw new Error(`Error while creating rule: ${response.status}: ${message}`)
    }

    const rule: Rule = await response.json()
    return rule
}

export async function updateDeviceRule({
    deviceId, 
    ruleId, 
    input
}: {
    deviceId: string, 
    ruleId: string, 
    input: UpdateRuleInput
}): Promise<Rule> {
    // Await device PATCH.
    const response = await fetch(`/api/devices/${deviceId}/rules/${ruleId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(input),
    })

    // Backend errors come from http.Error, which is plain text, not JSON. e.g. 400 Invalid JSON payload, 409 Duplicate.
    if (!response.ok) {
        const message = (await response.text()).trim()
        // TODO: Find some way of logging message. For now just
        // throw plain error to obfuscate errors from backend.
        throw new Error(`Failed to update device: ${response.status + ": " + message}`)
    }

    // 201 Created returns the new device as JSON.
    const updatedRule: Rule = await response.json()
    return updatedRule
}

export async function deleteDeviceRule({
    deviceId, 
    ruleId, 
}: {
    deviceId: string, 
    ruleId: string, 
}): Promise<string> {
    // Await device DELETE.
    const response = await fetch(`/api/devices/${deviceId}/rules/${ruleId}`, {
        method: "DELETE",
        headers: { "Content-Type": "application/json" }
    })

    // Backend errors come from http.Error, which is plain text, not JSON. e.g. 400 Invalid JSON payload, 409 Duplicate.
    if (!response.ok) {
        const message = (await response.text()).trim()
        // TODO: Find some way of logging message.
        throw new Error(`Failed to delete rule: ${response.status + ": " + message}`)
    }

    // 204 No content returned here, 
    return ""
}