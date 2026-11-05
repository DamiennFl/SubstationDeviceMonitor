import type { Device, CreateDeviceInput } from "../types/Types"

// Async function which will load devices
export async function loadDevices(): Promise<Device[]> {
    // Try to fetch devices
    const response = await fetch("/api/devices") // GET on /api/devices, await response

    // Handle response error
    if (!response.ok) {
        throw new Error(`Failed to load devices: ${response.status}`)
    }

    // Create device list from response. Matching properties are automatically mapped. Assumes JSON is properly formatted.
    // It should be, since we validate it before sending from backend
    const deviceList: Device[] = await response.json()
    // Set list
    return deviceList
}

// Sends a POST to create a new device. Returns the created Device, or throws with the backend's message.
export async function createDevice(input: CreateDeviceInput): Promise<Device> {
    // Await device POST. Like curl, fetch defaults to GET, so we define method here.
    const response = await fetch("/api/devices", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        // Backend uses DisallowUnknownFields, so only send name, ip, and type.
        body: JSON.stringify(input),
    })

    // Backend errors come from http.Error, which is plain text, not JSON. e.g. 400 Invalid JSON payload, 409 Duplicate.
    if (!response.ok) {
        const message = (await response.text()).trim()
        // TODO: Find some way of logging message. For now just
        // throw plain error to obfuscate errors from backend.
        throw new Error(`Failed to create device: ${response.status + ": " + message}`)
    }

    // 201 Created returns the new device as JSON.
    const createdDevice: Device = await response.json()
    return createdDevice
}