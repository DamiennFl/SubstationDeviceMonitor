import { useState } from 'react'
import { BrowserRouter } from 'react-router-dom'
import HeaderBar from './HeaderBar'
import SideBar from './SideBar'

// Main App component
export function AppShell() {
    const [darkMode, setDarkMode] = useState(true)

    const toggleTheme = () => {
        setDarkMode((currentDarkMode) => !currentDarkMode)
    }

    return (
        <BrowserRouter>
            <div className={`app-shell${darkMode ? ' dark' : ''}`}>
                <HeaderBar
                    darkMode={darkMode}
                    onToggleTheme={toggleTheme}
                />
                <SideBar />
            </div>
        </BrowserRouter>
    )
}

export default AppShell