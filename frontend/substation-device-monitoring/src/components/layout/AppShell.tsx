import { useState } from 'react'
import HeaderBar from './HeaderBar'

// Main App component
export function AppShell() {
    const [darkMode, setDarkMode] = useState(true)

    const toggleTheme = () => {
        setDarkMode((currentDarkMode) => !currentDarkMode)
    }

    return (
        <div className={`app-shell${darkMode ? ' dark' : ''}`}>
            <HeaderBar
                darkMode={darkMode}
                onToggleTheme={toggleTheme}
            />
        </div>
    )
}

export default AppShell