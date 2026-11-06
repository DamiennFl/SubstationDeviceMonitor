import { useParams } from "react-router-dom"
import { getDevice, getDeviceAlarms, getDeviceReadings } from "../api/single_device"
import { useState, useEffect } from "react"
import type { Device, RuleSnapshot, AlarmWithAcknowledgement, Reading } from "../types/Types"
import DeviceRules from "../components/rules/Rules"
import "./devices.css"

// Page for a single device. The device ID comes from the URL (/devices/:deviceId).
function SingleDevice() {
    const { deviceId } = useParams<{ deviceId: string }>()
    const [device, setDevice] = useState<Device | null>(null)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        if (!deviceId) return
        let cancelled = false

        async function loadDevice() {
            try {
                const result = await getDevice(deviceId!) // Non-null assertion
                if (!cancelled) {
                    setDevice(result)
                    setError(null)
                }
            } catch (error) {
                if (!cancelled) setError(error instanceof Error ? error.message : "Failed to get device")
            }
        }

        void loadDevice()
        return () => { cancelled = true} // Cleanup function. Will cancel setting device or error if effect runs again, like a new page for a different device.
    }, [deviceId]) // Re-render if deviceId changes. Required to prevent infinite re-renders of page.

    // Early returns after hooks. 3 cases: Device ID is missing, an error occurred, or no device is set (yet).
    if (!deviceId) return <div className="main-page">Missing device ID</div>
    if (error) return <div className="main-page"><h1>{error}</h1></div>
    if (!device) return <div className="main-page"><h1>Loading...</h1></div>

    // Fetch the device by ID and show rules, alarms, and live events.
    return (
        <div className="main-page">
            <DeviceHeader 
                device={device} />
            <DeviceRules device={device} />
            <DeviceAlarms device={device} />
            <DeviceReadings device={device} />
        </div>
    )
}

function DeviceHeader({device}: {device: Device}) {    
    // Read-only info shown plainly on the page background: name as the title, details listed underneath.
    return (
        <div className="single-device-header">
            <h1 className="single-device-header__name">{device.name}</h1>
            <div className="single-device-header__item">IP: {device.ip}</div>
            <div className="single-device-header__item">Type: {device.type.charAt(0).toUpperCase() + device.type.substring(1)}</div>
            <div className="single-device-header__item">Status: {device.status}</div>
            <div className="single-device-header__item">Last Seen: {device.lastSeen ?? "Never"}</div>
        </div>
    )
}




// Describe a rule's limit depending on its type: threshold shows comparison + limit,
// band shows the range, status shows the trip state.
function formatLimit(rule: RuleSnapshot): string {
    switch (rule.type) {
        case "threshold":
            return `${rule.comparison === "GT" ? ">" : "<"} ${rule.limit}`
        case "band":
            return `${rule.lowerLimit} to ${rule.upperLimit}`
        default:
            return rule.tripState ?? "-"
    }
}

function DeviceAlarms({device}: {device: Device}) {
    const [alarmList, setAlarmList] = useState<AlarmWithAcknowledgement[] | null>(null)
    const [alarmError, setAlarmError] = useState<string | null>(null)
    
    useEffect(() => {
        if (!device.deviceId) return
        let cancelled = false

        async function loadDeviceAlarms() {
            try {
                const result = await getDeviceAlarms(device.deviceId) // 
                if (!cancelled) {
                    setAlarmList(result)
                    setAlarmError(null) // Handle Alarm error in here, so it does not error the entire page.
                }
            } catch (error) {
                if (!cancelled) setAlarmError(error instanceof Error ? error.message : "Failed to get device alarms")
            }
        }

        void loadDeviceAlarms()
        const intervalId = setInterval(() => loadDeviceAlarms(), 5000) // Pass function which calls loadDeviceAlarms()

        return () => { 
            cancelled = true
            clearInterval(intervalId)} // Cleanup function. Will cancel setting device or error if effect runs again, like a new page for a different device.
    }, [device.deviceId]) // Re-render if deviceId changes. Required to prevent infinite re-renders of page.

    if (alarmList === null) {
        return alarmError
        ? <div className="device-create-error">{alarmError}</div>
        : <div>Loading Alarms...</div>
    }
    if (alarmList.length === 0) return <div><h1>No alarms</h1></div>

    return (
        <div>
            {alarmError &&
                <div className="device-create-error">
                    Getting device alarm list failed.
                </div>
            }
            <h2>Alarms</h2>
            {/* Scroll box: CSS caps the height at 5 rows, anything more scrolls. List arrives newest first. */}
            <div className="alarm-scroll">
                <table className="alarm-table">
                    <thead>
                        <tr>
                            <th>Rule</th>
                            <th>Type</th>
                            <th>Measurement</th>
                            <th>Limit</th>
                            <th>Trigger Value</th>
                            <th>Raised</th>
                        </tr>
                    </thead>
                    {!alarmError &&
                    <tbody>
                        {alarmList.map((alarm) => (
                            // Active alarm (no clearedAt) is red, cleared alarm is grey.
                            <tr
                                key={alarm.alarm.id}
                                className={alarm.alarm.clearedAt ? "alarm-row alarm-row--cleared" : "alarm-row alarm-row--active"}
                            >
                                <td>{alarm.alarm.ruleSnapshot.name}</td>
                                <td>{alarm.alarm.ruleSnapshot.type}</td>
                                <td>{alarm.alarm.ruleSnapshot.measurement}</td>
                                <td>{formatLimit(alarm.alarm.ruleSnapshot)}</td>
                                <td>{alarm.alarm.triggerValue}</td>
                                <td>{new Date(alarm.alarm.raisedAt).toLocaleString()}</td>
                            </tr>
                        ))}
                    </tbody>}
                </table>
            </div>
        </div>
    )
}

function DeviceReadings({device}: {device: Device}) {
    const [readingList, setReadingList] = useState<Reading[] | null>(null)
    const [readingError, setReadingError] = useState<string | null>(null)

    useEffect(() => {
        if (!device.deviceId) return
        let cancelled = false

        async function loadDeviceReadings() {
            try {
                const result = await getDeviceReadings(device.deviceId) // 
                if (!cancelled) {
                    setReadingList(result)
                    setReadingError(null) // Handle Alarm error in here, so it does not error the entire page.
                }
            } catch (error) {
                if (!cancelled) setReadingError(error instanceof Error ? error.message : "Failed to get device readings")
            }
        }

        void loadDeviceReadings()
        const intervalId = setInterval(() => loadDeviceReadings(), 1000) // Pass function which calls loadDeviceAlarms()

        return () => { 
            cancelled = true
            clearInterval(intervalId)} // Cleanup function. Will cancel setting device or error if effect runs again, like a new page for a different device.
    }, [device.deviceId]) // Re-render if deviceId changes. Required to prevent infinite re-renders of page.

    if (readingList === null) {
        return readingError && <div className="device-create-error">{readingError}</div>
    }
    if (readingList.length === 0) return <div>No readings</div>

    return (
        <div>
            {readingError &&
                <div className="device-create-error">
                    Getting device reading list failed.
                </div>
            }
            <h2>Readings</h2>
            {/* Scroll box: CSS caps the height at 5 rows, anything more scrolls. List arrives newest first. */}
            <div className="alarm-scroll">
                <table className="alarm-table">
                    <thead>
                        <tr>
                            <th>Breaker Status</th>
                            <th>Current A</th>
                            <th>Current B</th>
                            <th>Current C</th>
                            <th>Voltage A</th>
                            <th>Voltage B</th>
                            <th>Voltage C</th>
                            <th>Frequency</th>
                            <th>Real Power</th>
                            <th>Reactive Power</th>
                            <th>Timestamp</th>
                        </tr>
                    </thead>
                    {!readingError &&
                    <tbody>
                        {readingList.map((reading) => (
                            <tr className="alarm-row alarm-row--cleared">
                                <td>{reading.breakerStatus}</td>
                                <td>{reading.currentA}</td>
                                <td>{reading.currentB}</td>
                                <td>{reading.currentC}</td>
                                <td>{reading.voltageA}</td>
                                <td>{reading.voltageB}</td>
                                <td>{reading.voltageC}</td>
                                <td>{reading.frequency}</td>
                                <td>{reading.realPower}</td>
                                <td>{reading.reactivePower}</td>
                                <td>{reading.timestamp}</td>
                            </tr>
                        ))}
                    </tbody>}
                </table>
            </div>
        </div>
    )
}

export default SingleDevice
