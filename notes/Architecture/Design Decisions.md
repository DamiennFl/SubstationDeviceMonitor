Question: Why store an intermediary layer instead of handlers talking to database directly?
 - Alternative: Each handler manages its own queries from database, including validation and errors.
 - Failure: Rules that need all the data at once could fail during concurrent calls. Using an intermediary gives one entry point on each side for the data, with validation in between. This way, we can validate and modify data as it comes from the API, and also validate data before it gets sent to the API, while the API just handles requests. It keeps the scope relative to what each thing should do. API: Handle requests. Database: Store data. Go backend: Call and update database based on API requests.
 - Cost: An extra layer to handle. One more hop to trace when debugging.
 - When it's not worth it: A small CRUD app that does not need to maintain states across records, or a one-off script. Since we have persistent data and it matters when and how it changes, it makes sense to have an extra layer managing that. 

Question: Why validate in Go in the intermediary step? 
 - Alternatives: Validate data only in API, before data is sent to Go backend. Validate data only in database, and reject bad data. Or a mix of both.
 - Failure: Only the Go backend knows the information from both sides. Go can validate data that comes in from the API, and enforce rules that cover multiple different types of data, as well as return clear errors. Without this for example, the API may send a malformed object to the Go backend, which then sends to the database, which is expensive. Ideally we have validation in all three steps: API validates URLs and requests, Go backend validates the data sent through for things like duplicates, malformed data, etc. The database has constraints as well, such as uniqueness constraints, foreign keys and not-null checks, which can catch bugs and also race conditions. Those core rule constraints are the last line of defense before data is updated, so redundancy is acceptable.
 - Cost: Time loss for extra validation, and the rules have to be kept in sync manually.
 - When it's not worth it: When the database stores simple data, and does not care about when it is updated or whether there are cross-data rules. If each data type is self contained in the database, just validation in the API and database could be acceptable, but in some cases may not be enough.

Question: Why use PATCH methods over PUT when updating?
 - PUT completely replaces an item, or creates it if it does not exist. PATCH follows idempotency: multiple identical PATCH updates will not do anything more than the first, while multiple PUT updates will replace each time. Use PUT when you need to update an entire thing, potentially creating it or replacing it entirely, and PATCH for updating pieces while maintaining the state of everything else.

Question: How do you make sure the SSE handler is decoupled from the Store, to maintain separation between business logic and sending events to the frontend?
 - Alternatives: Events are published inside the store. This captures mutation but makes the Store aware of event delivery, which is not its job, thus adding coupling. Another alternative is inside an additional service layer, however then all mutation paths need to go through that service, which adds another level of complexity.
 - Failure: The SSE handler does not independently validate business data. Therefore, the internal event publisher must accept only typed events produced after successful mutations, and the broker must preserve the event envelope contract. meaning bad data could come through back to the frontend. Disciplined data sending to the SSE via event publisher is required.
 - Cost: Requires one shared event publisher which all required handlers map to, which adds another point of failure.
 - When it's not worth it: Not likely. It is best practice to decouple the SSE handler from the store, since the SSE handler does not need to know about any of the business logic (whether a reading is valid, whether it is a duplicate, etc.) It only needs to know: subbing and unsubbing clients, receiving already-formed events to server, serializing them, writing them to all connected clients, and handling disconnected clients. TODO: Make a diagram
 - Solution: All mutation handlers (specifically createReading, acknowledgeAlarm, more?) publishes the full details of what they updated (createReading returns the reading, but also raised and cleared alarms), to an independent broker. The external SSE handler in sse.go subscribes to that broker. This way, it gets all of the relevant information, while the store still exposes the minimum data necessary.The broker only knows what changed and needs to update, not how or why. Note that this is different from having an entire separate service layer, as that solution would coordinate mutation workflows (e.g. get reading event and create alarm events as well), while the implementation we are going for uses just one which arbitrarily takes events, and gives them to the SSE handler via the envelope design pattern.

 General Questions to learn about:
  - When should you .Encode into JSON, and when should you .Marshal JSON and then return it? Is there ever a time where you should? Seems like .Encode uses .Marshal under the hood, based on the docstring, so it's probably the same actually.
  - How does mux.HandleFunc() work explicitly? Where do the ResponseWriter and Request come from?

Limitations:
 - Event broker is very simple. Publishing is best-effort for delivery. Invalid event envelopes still return validation errors, but events may be dropped for slow or disconnected subscribers without reporting a delivery error. This may also de-sync the application from the actual state of the Store (and later the DB). The more important source is the store/DB, thus REST call returns will be the source of truth, while SSE events do their best to stay in sync.
 - Simulator sends readings sequentially, for simplicity.

Improvements
 - Right now, when a device is deleted, or a rule is disabled/deleted, nothing will show up in the live events.
 - Make simulator send readings via concurrent goroutines, instead of sequentially.
 - Use custom error structs and JSON creation for error returns to API, so API does not have to parse.
 - Event Broker was implemented but is only used by Live Events tab, streaming all events. It would be more efficient for a larger scale production to also use it for single device alarms and readings, instead of polling.
 - Use JSON schema for defining JSON. Adds an extra step of validation
 - Use `tygo` package to create typescript types from Go, instead of manually redefining them

Other
 - Tests follow Assemble-Act-Assert unit testing format. Each function has 3 tests: regular case, edge case, invalid case. In a real production environment there would be more, such as testing a failing input for each individually. Tests generally follow this format: Create something, run the function on it, then validate errors (or lack thereof), and validate values if necessary. Some could be improved or expanded upon, e.g. tests which call functions that return JSON. Right now the JSON is not validated in the tests.
 
 - Not implemented right now: On-delay alarm activation, e.g. raise alarm after n-seconds of value out of range. This would mean maintaining a sliding window of values for each measurement for rule for each device. I don't want to do that for this project. So, deadband is only used for clearing alarms, meaning if a value trips because a value went above a number such as 400 with a 5 deadband, then the value must go under 395 before alarm clears.

 - Right now the AddReading function just returns a reading, however since we are implementing SSE events it now needs to also signal which alarms it raised and cleared. Thus we refactor the returns of the Store functions to return structs containing all of the changes, instead of individual objects.

 - In the in-memory store implementation, we were able to dynamically update relationships as they occurred. An example is: A reading comes in, when it is added, we check if its timestamp is newer than the latest reading for its matching device. If it is, we update the device's latest reading to point to that new reading. However for the Postgres database, we can't maintain relationships in that way. Instead of duplicating data into each device each time a reading comes in, which can be slow and cause disagreements, we will instead use lateral lookups to match readings to devices when we get or list devices. This is also faster than simply doing two queries on the separate tables, as Postgres has to then organize and search the readings table. Instead we 