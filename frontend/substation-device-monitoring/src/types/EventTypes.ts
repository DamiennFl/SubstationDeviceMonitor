
export const EventType = {
    ReadingEvent: "reading",
    AlarmRaisedEvent: "alarmRaised",
    AlarmClearedEvent: "alarmCleared",
    AlarmAckedEvent: "alarmAcked"
} as const

export type EventType = typeof EventType[keyof typeof EventType]

// TODO: Consider renaming this, since React already holds an Event type.
export type Event = {
    eventId: string
    deviceId: string
    eventType: EventType
    eventTimeStamp: string
    data: string
}