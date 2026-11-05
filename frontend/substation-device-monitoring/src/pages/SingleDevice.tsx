import { useParams } from "react-router-dom"

// Page for a single device. The device ID comes from the URL (/devices/:deviceId).
function SingleDevice() {
    const { deviceId } = useParams<{ deviceId: string }>()

    // TODO: fetch the device by ID and show rules, alarms, and live events.
    return (
        <div className="main-page">
            <h2>Device {deviceId}</h2>
        </div>
    )
}

export default SingleDevice
