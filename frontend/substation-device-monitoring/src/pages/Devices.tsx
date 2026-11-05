import { useEffect, useState } from "react"
import type { Device } from "../types/Types"
import "./devices.css"


// Main Page component for Device List, for clean page definition.
function Devices() {
    const [showCreateDeviceRow, setShowCreateDeviceRow] = useState(false)

    return (
        <div className="main-page">
            <DeviceCreateButton setShowCreateDeviceRow={setShowCreateDeviceRow} />
            <DeviceRowHeader />
            {showCreateDeviceRow && <DeviceCreateRow />}
            <DeviceList />
        </div>
    )
}

function DeviceCreateButton({
    setShowCreateDeviceRow,
}: {
    setShowCreateDeviceRow: (show: boolean) => void
}) {
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

function DeviceCreateRow() {
    const [name, setName] = useState("")
    const [ip, setIp] = useState("")

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
            </div>
            <DeviceCreateButtons />
        </div>
    )
}

// List of Devices in a table format
function DeviceList() {
    const [devices, setDevices] = useState<Device[]>([]) // Define list of Devices, empty list init
    const [error, setError] = useState<string | null>(null) // Define error, no error init

    useEffect(() => { // Run once, on page load. Calls API to get device list via fetch
        async function loadDevices() {
            // Try to fetch devices
            try {
                const response = await fetch("/api/devices")

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
                setError(
                    error instanceof Error ? error.message : "Failed to load devices",
                )
            }
        }

        // Discard returned Promise, since React only expects useEffect's callback to return nothing or a cleanup function.
        void loadDevices()
    }, []) // [] is dependency array. Empty because we only load once when DeviceList is mounted.

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

function DeviceCreateButtons() {
    return (
        <div className="device-buttons">
            <button className="device-button">
                <svg
                    width="12"
                    height="12"
                    viewBox="0 0 12 12"
                    fill="none"
                    aria-label="Create device"
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
            <button className="device-button">
                <svg
                    width="12"
                    height="12"
                    viewBox="0 0 12 12"
                    fill="none"
                    aria-label="Cancel creating device"
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