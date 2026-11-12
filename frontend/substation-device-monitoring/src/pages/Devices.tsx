import { useEffect, useState } from "react"
import { Link } from "react-router-dom"
import type { Device } from "../types/Types"
import { DeviceType } from "../types/Types" // Not imported as type to reference value device value directly
import { loadDevices, createDevice, updateDevice, deleteDevice } from "../api/devices"
import { useTempMessage } from "../hooks/useTempMessage"
import "./devices.css"

// Main Page component for Device List, for clean page definition.
function Devices() {
    const [devices, setDevices] = useState<Device[]>([]) // Define list of Devices, empty list init
    const [error, setError] = useState<string | null>(null) // Define error, no error init
    const [showCreateDeviceRow, setShowCreateDeviceRow] = useState(false) // Define bool for showing creation row

    // Wrap async in useEffect, otherwise infinite loop because handleLoading calls setDevices
    // which triggers a re-render, which would cause handleLoading to call again.
    useEffect(() =>  {
        async function handleLoading() {
            try {
                const deviceList = await loadDevices()
                setDevices(deviceList)
                setError(null)
            } catch (error) {
                setError(error instanceof Error ? error.message : "Failed to load devices")
            }
        }

        void handleLoading()
    }, [setDevices, setError])

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
                error={error}
                onDeviceUpdated={(updated) =>
                setDevices((current) =>
                    current.map((d) => (d.deviceId === updated.deviceId ? updated : d)),
                )}
                // Filter out deleted device on deletion while updating devices list
                onDeviceDeleted={(deleted) => 
                    setDevices((current) => current.filter((d) => d.deviceId !== deleted))
                }
            />
        </div>
    )
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
    const [createError, setCreateError] = useTempMessage()
    // Prevents double clicks from sending multiple POST requests.
    const [submitting, setSubmitting] = useState(false)

    // Validation and submission async function
    async function handleConfirm() {
        // Validate name. Just don't allow empty name: duplicate names are handled by backend.
        if (name.trim() === "") {
            setCreateError("Name is required")
            return
        }
        if (ip.trim() === "") {
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
                type: DeviceType.Relay
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
    error,
    onDeviceUpdated,
    onDeviceDeleted,
}: {
    devices: Device[]
    error: string | null
    onDeviceUpdated: (device: Device) => void
    onDeviceDeleted: (deviceId: string) => void
}) {
    // Map all devices in device rows, if no error. Otherwise return empty div with error.
    return (
        error ? <div>{error}</div> :
        <div className="device-table">
            {devices.map((device) => (
                <DeviceRow
                    key={device.deviceId}
                    device={device}
                    onUpdated={onDeviceUpdated}
                    onDeleted={onDeviceDeleted}
                />
            ))}
        </div>
    )
}

// Represents one Device
function DeviceRow({ 
    device,
    onUpdated,
    onDeleted,
}: {
    device: Device,
    onUpdated: (device: Device) => void
    onDeleted: (deviceId: string) => void
}) {
    const [editing, setEditing] = useState(false)
    const [name, setName] = useState(device.name)
    const [ip, setIp] = useState(device.ip)
    const [updateError, setUpdateError] = useTempMessage()
    const [deleteError, setDeleteError] = useTempMessage()
    const [submitting, setSubmitting] = useState(false)

    // Validation and submission async function
    async function handleUpdate() {
        // Validate name. Just don't allow empty name: duplicate names are handled by backend.
        if (name.trim() === "") {
            setUpdateError("Name is required")
            return
        }
        if (ip.trim() === "") {
            setUpdateError("IP is required")
            return
        }

        // Block multiple submits, and set error to null
        setSubmitting(true)
        setUpdateError(null)
        try {
            // Update device
            const editedDevice = await updateDevice(device.deviceId,
                { name: name.trim(), ip: ip.trim() }
            )
            onUpdated(editedDevice) // Pass device back via callback
            setEditing(false)
        } catch (error) { // Error caught, likely during device update.
            // set UpdateError to error message or fall back to default.
            setUpdateError(
                error instanceof Error ? error.message : "Failed to update device",
            )
        } finally {
            setSubmitting(false)
        }
    }

    async function handleDelete() {
        setSubmitting(true)
        setDeleteError(null)

        try {
            await deleteDevice(device.deviceId)
            onDeleted(device.deviceId)
        } catch (error) {
            setDeleteError(
                error instanceof Error ? error.message : "Failed to delete device",
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
                        <input
                            type="text"
                            value={name}
                            onChange={(event) => setName(event.target.value)}
                            placeholder="Device name"
                        />
                    </label>
                    <label className="device-create-field">
                        <input
                            type="text"
                            value={ip}
                            onChange={(event) => setIp(event.target.value)}
                            placeholder="192.168.1.1"
                        />
                    </label>
                </div>
                {updateError !== null && (
                    <div className="device-create-error" role="alert">
                        {updateError}
                    </div>
                )}
            {/* DeviceCreateButtons handle creating and cancelling. Disable buttons when submitting.*/}
            <DeviceCreateButtons
                onConfirm={() => {
                    void handleUpdate()
                }}

                onCancel={() => {
                    setName(device.name)
                    setIp(device.ip)
                    setEditing(false)
                }}
                disabled={submitting}
            />
        </div>
    ) : (
        <div className="device-row">
            <DeviceInfo device={device} />
            <DeviceButtons
                deviceId={device.deviceId}
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
    // Use device status as key for displaying status color after
    return (
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
function DeviceButtons({
    deviceId,
    onEdit,
    onDelete
} : {
    deviceId: string
    onEdit: () => void
    onDelete: () => void
}
) {
    return (
        // In order: View Button, Edit Button, Delete Button
        <div className="device-buttons">
            <Link className="device-button" to={`/devices/${deviceId}`}>
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
            </Link>
            <button
                className="device-button"
                onClick={onEdit}
                >
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
            <button 
                className="device-button"
                onClick={onDelete}>
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