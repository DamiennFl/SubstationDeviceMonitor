import { useState, useEffect } from 'react'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import HeaderBar from './HeaderBar'
import SideBar from './SideBar'
import Devices from '../../pages/Devices'
import SingleDevice from '../../pages/SingleDevice'
import LiveEvents from '../../pages/LiveEvents'
import { subToEventStream } from '../../api/live_events'
import type { Event } from '../../types/EventTypes'

// Main App component
export function AppShell() {
    const [darkMode, setDarkMode] = useState(true)
    // Live events live here (not in LiveEvents) so history survives navigating between pages.
    const [events, setEvents] = useState<Event[]>([])

    // Subscribe once for the lifetime of the app. Newest event goes first.
    useEffect(() => {
        const eventSource = subToEventStream((event) => {
            setEvents(prev => [event, ...prev].slice(0, 200)) // Last 200 events, to prevent infinite memory growth
        })
        return () => eventSource.close()
    }, [])

    const toggleTheme = () => {
        setDarkMode((currentDarkMode) => !currentDarkMode)
    }

    return (
        <BrowserRouter>
        {/* Put Header and Sidebar outside of main app pane, since they do not change. */}
            <HeaderBar
                darkMode={darkMode}
                onToggleTheme={toggleTheme}
            />
            <SideBar />
            <div className={`app-shell${darkMode ? ' dark' : ''}`}>
                <Routes>
                    {/* Route to Devices page when /devices is linked via a NavLink somewhere else. */}
                    <Route path="/devices" element={<Devices />} />
                    {/* :deviceId is read in SingleDevice via useParams. */}
                    <Route path="/devices/:deviceId" element={<SingleDevice />} />
                    <Route path="/events" element={<LiveEvents events={events} />} />
                </Routes>
            </div>
        </BrowserRouter>
    )
}

export default AppShell