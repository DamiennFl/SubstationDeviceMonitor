Think of things in terms of shape, function, and state, for each component.
Example: A device component on the device list page. It's shape: horizontal item in list, with basic information, and buttons to view, modify, and delete it (each are components as well, since they will be reused). The function: Show device, allow modification, deletion, and routing to full device page view.

Backend sends errors via http.Error, which is plain text and not JSON. I'm choosing to keep this structure so I do not have to add error structs and JSON creation in the backend, as this is a lot of extra work at this time. Logged as a future improvement.

It is cleaner to extract functions which hit API endpoints and put them in a separate API module. Decouple them by never passing in setters, insead just returning a Promise. The component calling the function should then deal with the valid return or error.

For props and states, think about the nearest ancestor to whatever is dealing with the state and data. It likely should live there.