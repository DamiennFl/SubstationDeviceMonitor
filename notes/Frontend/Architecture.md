## Frontend Architecture

### Initialization

npm create vite@latest will scaffold new project using Vite, so I don't have to configure the entire project
myself.

Dependency errors after build. Typescript@7.0.2 was installed, but typescript-eslint@8.72.1 requires <6.1.0. Fixed by npm installing TS 5.9.3, then npm audit fix --force. npm ls produced no dependency errors.

### Visual Scaffolding

What does the frontend need?

Firstly, a main page. This main page will display a few things:
 - Sidebar and top bar for navigation
    - Sidebar contains Dashboard, Devices, Alarms, Live Events, and Settings.
    - Header bar contains logo and app name, alarm count notification, account button on the right.
 - Previewed list of alarms
    - List by newest alarm triggered, with timestamp. and rule which triggered it.
 - Live events, as they come through
    - List by newest live event, with timestamp.
 - An info box at the top containing the IP, device count and current alarm count.

Device Page:
 - List devices, with a create and deleted device button. Delete device is greyed until a device is selected.
 - Devices are clickable, which will display device info, its rules and its active alarms.

Alarm Page:
 - List alarms, with an acknowledge button. Greyed until one alarm or more is selected.
 - Clicking on alarm adds button: See device which brings you to device page for it.

Live Events:
 - For now just one large box of events as they come in.

So, all required Components:
 - Application Shell
 - Router
 - Typescript API module to separate frontend and backend
 - A shared state, perhaps global if necessary
 - Loading, Error, Empty, Confirm, and Notification Components

 - Header Bar
 - Sidebar
    - Button components (separate component or simple buttons)
 - Dashboard Shell Component
    - Info Display
    - Alarm Display
    - Live Events Display
 - Device List Page Shell Component
    - Device Table Component with clickable devices
    - Create Device and Delete Device Buttons

 - Device Page Shell Component
    - Device Info Display
    - Device Rules List
    - Device Rules Create Edit and Delete Buttons, likely on same line as the Rule, and values are editable in line. No need for separate page.
    - Device Edit and Delete Buttons

 - Alarms Page Shell Component
    - Alarms List
    - Acknowledge Button for one or more selected alarms

 - Events Page Shell Component
    - Large Table list of live events. Can click on event and it will display the raw JSON.