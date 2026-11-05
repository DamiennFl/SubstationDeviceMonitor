import { NavLink } from 'react-router-dom'

function SideBar() {
    return (
        // Map side bar items to SideBarButtons.
        <aside className="side-bar">
            {sideBarItems.map((item) => (
                <SideBarButton
                  key={item.route}
                  label={item.label}
                  route={item.route}
                />
            ))}
        </aside>
    )
}

// Common structure for sidebar buttons, for easier defining
type SideBarItem = {
    label: string
    route: string
}

// List of sidebar buttons. Update this list as necessary to add or remove buttons.
const sideBarItems: SideBarItem[] = [
    {label: 'Dashboard', route: '/' },
    {label: 'Devices', route: '/devices' },
    //{label: 'Alarms', route: '/alarms'}, // TODO: Add an endpoint which gets alarms list. Does not exist.
    {label: 'Live Events', route: '/events'},
    {label: 'Settings', route: '/settings'} // Does not actually exist, will be used for PAGE NOT FOUND learning
]

type SideBarButtonProps = {
    label: string
    route: string
}

// Common SideBarButton
function SideBarButton({ label, route }: SideBarButtonProps) {
    return (
        <NavLink className="side-bar__button" to={route}>
            {label}
        </NavLink>
    )
}

export default SideBar