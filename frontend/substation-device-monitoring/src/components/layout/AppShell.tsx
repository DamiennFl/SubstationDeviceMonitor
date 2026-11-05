import { useState } from 'react'
import { BrowserRouter, Route, Routes } from 'react-router-dom'
import HeaderBar from './HeaderBar'
import SideBar from './SideBar'
import Devices from '../../pages/Devices'

// Main App component
export function AppShell() {
    const [darkMode, setDarkMode] = useState(true)

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
                </Routes>
            </div>
        </BrowserRouter>
    )
}

export default AppShell