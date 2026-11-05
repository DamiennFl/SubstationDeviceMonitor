import { useEffect, useState } from "react"
import type { Dispatch, SetStateAction } from "react"
import { DeviceType } from "../types/Types"
import type { CreateDeviceInput, Device } from "../types/Types"
import "./devices.css"

// Main Page component for Device List, for clean page definition.
function Devices() {
    const [devices, setDevices] = useState<Device[]>([]) // Define list of Devices, empty list init
    const [error, setError] = useState<string | null>(null) // Define error, no error init
    const [showCreateDeviceRow, setShowCreateDeviceRow] = useState(false) // Define bool for showing creation row

    return (
        <div className="main-page">
            <DeviceCreateButton setShowCreateDeviceRow={setShowCreateDeviceRow} />
            <DeviceRowHeader />
            {showCreateDeviceRow && ( // Conditionally render device creation row
                <DeviceCreateRow
                    // Append the device returned by the API via onCreated callback prop, then hide the create row
                    onCreated={(device) => {
                        setDevices((currentDevices) => [...currentDevices, device]) // Append device to currentDevices
                        setShowCreateDeviceRow(false) //hide row via setter
                    }}
                    onCancel={() => setShowCreateDeviceRow(false)} // On cancel, just hide row
                />
            )}
            <DeviceList 
                devices={devices} 
                setDevices={setDevices}
                error={error}
                setError={setError}/>
        </div>
    )
}

// Async function which will load 
async function loadDevices({setDevices,setError}: {
    setDevices: Dispatch<SetStateAction<Device[]>> // Explicitly state setter types.
    setError: Dispatch<SetStateAction<string | null>>
}) {
    // Try to fetch devices
    try {
        const response = await fetch("/api/devices") // GET on /api/devices, await response

        // Handle response error
        if (!response.ok) {
            throw new Error(`Failed to load devices: ${response.status}`)
        }

        // Create device list from response. Matching properties are automatically mapped. Assumes JSON is properly formatted.
        // It should be, since we validate it before sending from backend
        const deviceList: Device[] = await response.json()
        // Set list
        setDevices(deviceList)
    } catch (error) {
        // Set error if one occurred
        setError(
            error instanceof Error ? error.message : "Failed to load devices",
        )
    }
}

// Sends a POST to create a new device. Returns the created Device, or throws with the backend's message.
async function createDevice(input: CreateDeviceInput): Promise<Device> {
    // Await device POST. Like curl, fetch defaults to GET, so we define method here.
    const response = await fetch("/api/devices", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        // Backend uses DisallowUnknownFields, so only send name, ip, and type.
        body: JSON.stringify(input),
    })

    // Backend errors come from http.Error, which is plain text, not JSON. e.g. 400 Invalid JSON payload, 409 Duplicate.
    if (!response.ok) {
        //const message = (await response.text()).trim()
        // TODO: Find some way of logging message. For now just
        // throw plain error to obfuscate errors from backend.
        throw new Error(`Failed to create device: ${response.status}`)
    }

    // 201 Created returns the new device as JSON.
    const createdDevice: Device = await response.json()
    return createdDevice
}

// Create Device button definition. Takes in setShowCreateDeviceRow callback.
// Since we define it the same as the underlying function definition for setShowCreateDeviceRow
// from the useState in the Device component, we can pass it as a callback directly.
function DeviceCreateButton({
    setShowCreateDeviceRow,
}: {
    setShowCreateDeviceRow: (show: boolean) => void
}) {
    // Return button with onClick trigger
    return (
        <div className="device-page-button-container">
            <button
                className="device-create-button"
                onClick={() => setShowCreateDeviceRow(true)}
            >
                Create Device
            </button>
        </div>
    )
}

// Conditionally rendered row which handles device creation.
// Takes in two callback functions and returns one or the other.
function DeviceCreateRow({
    onCreated,
    onCancel,
}: {
    onCreated: (device: Device) => void
    onCancel: () => void
}) {
    // Define own function states
    const [name, setName] = useState("") // For name box form
    const [ip, setIp] = useState("") // For IP box form
    // Separate from page error, so a failed create only shows in this row instead of replacing the device list.
    const [createError, setCreateError] = useState<string | null>(null)
    // Prevents double clicks from sending multiple POST requests.
    const [submitting, setSubmitting] = useState(false)

    // Validation and submission async function
    async function handleConfirm() {
        // Validate name. Just don't allow empty name: duplicate names are handled by backend.
        if (name.trim() === "") {
            setCreateError("Name is required")
            return
        }
        if (ip.trim() == "") {
            setCreateError("IP is required")
            return
        }

        // Block multiple submits, and set error to null
        setSubmitting(true)
        setCreateError(null)
        try {
            // Only relays exist right now, so type is fixed.
            // TODO: If new types are added, update this to be dynamic.
            const createdDevice = await createDevice({ // await creating device
                name: name.trim(),
                ip: ip.trim(),
                type: DeviceType.Relay,
            })
            // If no error caught during device creation, onCreated callback gets new device
            onCreated(createdDevice)
        } catch (error) { // Error caught, likely during device creation.
            // set CreateError to error message or fall back to default.
            setCreateError(
                error instanceof Error ? error.message : "Failed to create device",
            )
            setSubmitting(false)
        }
    }

    // Return div. Name and IP forms set Name and IP states accordingly.
    return (
        <div className="device-row device-create-row">
            <div className="device-info">
                <label className="device-create-field">
                    Name
                    <input
                        type="text"
                        value={name}
                        onChange={(event) => setName(event.target.value)}
                        placeholder="Device name"
                    />
                </label>
                <label className="device-create-field">
                    IP
                    <input
                        type="text"
                        value={ip}
                        onChange={(event) => setIp(event.target.value)}
                        placeholder="192.168.1.1"
                    />
                </label>
                {createError !== null && (
                    <div className="device-create-error" role="alert">
                        {createError}
                    </div>
                )}
            </div>
            {/* DeviceCreateButtons handle creating and cancelling. Disable buttons when submitting.*/}
            <DeviceCreateButtons
                onConfirm={() => void handleConfirm()}
                onCancel={onCancel}
                disabled={submitting}
            />
        </div>
    )
}

// List of Devices in a table format
function DeviceList({
    devices,
    setDevices,
    error,
    setError,
}: {
    devices: Device[]
    setDevices: Dispatch<SetStateAction<Device[]>>
    error: string | null
    setError: Dispatch<SetStateAction<string | null>>
}) {
    useEffect(() => { // Run once, on page load. Calls API to get device list via fetch
        void loadDevices( {setDevices, setError} )
    }, [setDevices, setError]) // [] is dependency array. setDevices and setError inside because they arrive as props.
                     // useEffect would then re-run if setDevices or setError state setters changed identity.

    // If error after loading devices (setError was called at some point in loadDevices), render it.
    if (error !== null) {
        return <div>{error}</div>
    }

    // Map all devices in device rows.
    return (
        <div className="device-table">
            {devices.map((device) => (
                <DeviceRow
                    key={device.deviceId}
                    device={device}
                />
            ))}
        </div>
    )
}

// Represents one Device
function DeviceRow({ device }: {device: Device}) {
    return (
        <div className="device-row">
            <DeviceInfo device={device}/>
            <DeviceButtons />
        </div>
    )
}

// Device Table header. Defined separately to allow device creation row to render underneath.
function DeviceRowHeader() {
    return (
        <div className="device-row device-header-row">
            <div className="device-info device-header-info">
                <div className="device-info__header">Name</div>
                <div className="device-info__header">IP</div>
                <div className="device-info__header">Status</div>
            </div>
            <div className="device-buttons" aria-hidden="true" />
        </div>
    )
}

// Displays the device info: Name, IP, and status.
function DeviceInfo({ device }: { device: Device }) {
    return (
        // Use device status as key for displaying status color after
        <div className="device-info">
            <div className="device-info__static">
                {device.name}
            </div>
            <div className="device-info__static">
                {device.ip}
            </div>
            <div className="device-info__status">
                {device.status}
            </div>
        </div>
    )
}

// DeviceButtons defines right-aligned set of buttons for a device row.
function DeviceButtons() {
    return (
        <div className="device-buttons">
            <button className="device-button">
                <svg
                    width="12"
                    height="12"
                    viewBox="0 0 12 12"
                    fill="none"
                    aria-label="Open device page"
                >
                    <path
                        d="M7 2h3v3M10 2 6.25 5.75M9 7.5V10H2V3h2.5"
                        stroke="currentColor"
                        strokeWidth="1.2"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                    />
                </svg>
            </button>
            <button className="device-button">
                <svg
                    width="12"
                    height="12"
                    viewBox="0 0 12 12"
                    fill="none"
                    aria-label="Edit device"
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
            <button className="device-button">
                <svg
                    width="12"
                    height="12"
                    viewBox="0 0 12 12"
                    fill="none"
                    aria-label="Delete device"
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

// Define create row buttons (Confirm and Cancel).
function DeviceCreateButtons({
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
                aria-label="Create device"
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
                aria-label="Cancel creating device"
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

export default Devices