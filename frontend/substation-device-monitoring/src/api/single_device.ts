import type { Device, AlarmWithAcknowledgement, Reading } from "../types/Types"

// getDevice gets a single device
export async function getDevice(deviceId: string): Promise<Device> {
    // Fetch with device ID
    const response = await fetch(`/api/devices/${deviceId}`, {
        headers: {"Content-Type": "application/json"}
    })

    // If bad response, print error with status and message
    if (!response.ok) {
        const message = (await response.text()).trim()
        throw new Error(`Failed to GET device ${deviceId}: ` + response.status + ": " + message)
    }

    // Create device object and return it
    const device: Device = await response.json()
    return device
}

// getDeviceAlarms gets alarms
export async function getDeviceAlarms(deviceId: string): Promise<AlarmWithAcknowledgement[]> {
    // Fetch with Device ID
    const response = await fetch(`/api/devices/${deviceId}/alarms`, {
        headers: {"Content-Type": "application/json"}
    })

    // If bad response, print error with status and message
    if (!response.ok) {
        const message = (await response.text()).trim()
        throw new Error(`Failed to GET device ${deviceId} alarms: ` + response.status + ": " + message)
    }

    // Return Alarm with Acknowledgment. We return this wrapper type so
    // acknowledgements can be read and added.
    const alarmList: AlarmWithAcknowledgement[] = await response.json()
    return alarmList
}

// getDeviceReadings gets a list of device readings. For now just gets the last hour,
// however could support "from" and "to" input to define a larger or smaller range.
export async function getDeviceReadings(deviceId: string): Promise<Reading[]> {
    // Statically define range
    const to = new Date()
    const from = new Date(to.getTime() - (60 * 60 * 1000)) // One hour before to

    // Set from and to as search params: from=...&to=...
    const params = new URLSearchParams({
        from: from.toISOString(), // use toISOString which matches Go's format: RFC3339
        to: to.toISOString()
    })
    
    // Fetch device readings
    const response = await fetch(`/api/devices/${deviceId}/readings?${params}`, {
        headers: {"Content-Type": "application/json"}
    })

    // If bad response, print error with status and message
    if (!response.ok) {
        const message = (await response.text()).trim()
        throw new Error(`Failed to GET device ${deviceId} readings: ` + response.status + ": " + message)
    }

    // Return list of readings from JSON response
    const readingList: Reading[] = await response.json()
    return readingList
}