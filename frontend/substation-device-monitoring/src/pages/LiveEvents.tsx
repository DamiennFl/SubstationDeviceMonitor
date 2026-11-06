import { useState, useEffect } from "react"
import { subToEventStream } from "../api/live_events"
import { type Event } from "../types/EventTypes"


// Live Events Component for LiveEvents page
function LiveEvents() {
    // List of events
    const [events, setEvents] = useState<Event[]>([])
    
    // Get event source at startup.
    useEffect(() => {
        // Get event source, callback updates events list by putting newest event up top.
        // Events come from API in order.
        const eventSource = subToEventStream((event) => {
            setEvents(prev => [event, ...prev].slice(0, 200)) // Last 200 events, to prevent infinite memory growth
        })

        // Return cleanup function.
        return () => eventSource.close()
    }, [])

    return (
    <div>
        {events.map(event => ( 
            <li key={event.eventId}>
                {event.eventTimeStamp} {event.eventType} {event.data}
            </li>
        ))}
    </div>
    )
}

export default LiveEvents