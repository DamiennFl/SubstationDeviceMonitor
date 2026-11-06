import { useState, useEffect } from "react"
import { subToEventStream } from "../api/live_events"
import { type Event } from "../types/EventTypes"


function LiveEvents() {
    const [events, setEvents] = useState<Event[]>([])
    
    useEffect(() => {
        const eventSource = subToEventStream((event) => {
            setEvents(prev => [event, ...prev].slice(0, 200)) // Last 200 events, to prevent infinite memory growth
        })

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