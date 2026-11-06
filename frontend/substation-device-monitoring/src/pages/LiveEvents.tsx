import { useState, useEffect } from "react"
import { loadDevices } from "../api/devices"
import { type Event, EventType } from "../types/EventTypes"
import "./liveEvents.css"


// Formats an event's JSON data string into a readable summary.
// Falls back to the raw string if the payload can't be parsed.
function formatEventData(event: Event): string {
    try {
        // Parse Event Data. May be reading data of alarm data.
        const parsed = JSON.parse(event.data)
        // Switch based on the event's type
        switch (event.eventType) {
            case EventType.ReadingEvent: {
                // Just list the values, comma separated (skip identity/timestamp fields).
                // This call creates key value pairs from the JSON
                const { deviceId, timestamp, ...values } = parsed
                void deviceId
                void timestamp
                // Map key value pairs together in one line
                return Object.entries(values).map(([key, value]) => `${key}=${value}`).join(", ")
            }
            // If alarm raised or cleared, just display the rule which triggered it and the value
            case EventType.AlarmRaisedEvent:
                return `${parsed.ruleSnapshot?.name}: ${parsed.triggerValue}`
            case EventType.AlarmClearedEvent:
                return `${parsed.ruleSnapshot?.name}: ${parsed.triggerValue}`
            default:
                return event.data
        }
    } catch {
        return event.data
    }
}

// Live Events Component for LiveEvents page
function LiveEvents({ events }: { events: Event[] }) {
    // Map of deviceId -> device name, so events show names instead of UUIDs.
    const [deviceNames, setDeviceNames] = useState<Record<string, string>>({})

    // Load devices once at startup (single GET /api/devices) and build the lookup map.
    useEffect(() => {
        loadDevices() // Use .then because loadDevices does not take a callback directly
                      // but we do not need a wrapper function
            .then(devices => {
                setDeviceNames(Object.fromEntries(devices.map(d => [d.deviceId, d.name])))
            })
            .catch(() => {
                // Non-fatal: rows fall back to showing the raw device ID.
            })
    }, [])

    // Return simple table
    return (
    <table className="events-table">
        <thead>
            <tr>
                <th>Event Type</th>
                <th>Device</th>
                <th>Data</th>
                <th>Timestamp</th>
            </tr>
        </thead>
        <tbody>
            {events.map(event => (
                <tr key={event.eventId}>
                    <td>{event.eventType}</td>
                    <td>{deviceNames[event.deviceId] ?? event.deviceId}</td>
                    <td className="events-data">{formatEventData(event)}</td>
                    <td>{event.eventTimeStamp}</td>
                </tr>
            ))}
        </tbody>
    </table>
    )
}

export default LiveEvents
