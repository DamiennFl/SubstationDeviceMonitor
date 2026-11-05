## Typescript Notes

### Callbacks
A function is assignable to a callback type if it can accept everything the callback type promises to pass in.

### Useful built-in types

`Dispatch<SetStateAction<T>>` is the required type for useState setters passed in as props. `useState<T>` return `[T, Dispatch<SetStateAction<T>>]`.
Better design pattern is to pass a narrow callback, instead of the raw setter, if the child does not care about the state. Example: `onDevicesLoaded: (d: []Device) => void` would replace `setDevices: Dispatch<SetStateAction<Device[]>>` in Devices.tsx `loadDevices` function, which would decouple loadDevices from actually updating the list of devices, and instead just return the list for the Devices component, which owns the device list and the setter, to set itself. See `onCreated` in the same file for an example of this implemented.
