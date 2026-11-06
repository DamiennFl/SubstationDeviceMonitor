import { EventType, type Event } from "../types/EventTypes"

// Event Stream function for live events from "/api/events"
export function subToEventStream(
    onEvent: (event: Event) => void
): EventSource {
    // Define event source streaming from URL
    const eventSource = new EventSource("/api/events")

    // Define common handle for events, since each type
    // follows the Event type structure 
    const handle = (e: MessageEvent<string>) => {
        // Parse into Event object
        onEvent(JSON.parse(e.data) as Event)
    }

    // Iterate EventType, register shared handler for each named
    // event type. Consumers can filter on event.eventType
    Object.values(EventType).forEach((type) => {
        eventSource.addEventListener(type, handle)
    })

    // Handle stream errors or closed connections
    eventSource.onerror = (error) => {
        console.error("SSE connection failed:", error)
        eventSource.close(); // Close to prevent infinite auto-reconnect loops
                             // TODO: Write reconnect logic in LiveEvents.tsx
    }

    return eventSource
}
