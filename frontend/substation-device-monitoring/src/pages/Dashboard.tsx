import { useState } from "react"
import { type Event } from "../types/EventTypes"

function Dashboard( { events }: { events: Event[] } ) {
    return <div>

    </div>
}

function InfoBox() {
    const [IP, setIP] = useState<string>("")
    
    return <div></div>
}

function RecentAlarms( {events}: {events: Event[] }) {
    return <div></div>
}

export default Dashboard