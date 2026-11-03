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


 General Questions to learn about:
  - When should you .Encode into JSON, and when should you .Marshal JSON and then return it? Is there ever a time where you should? Seems like .Encode uses .Marshal under the hood, based on the docstring, so it's probably the same actually.
  - How does mux.HandleFunc() work explicitly? Where do the ResponseWriter and Request come from?


Other
 - Tests follow Assemble-Act-Assert unit testing format. Each function has 3 tests: regular case, edge case, invalid case. In a real production environment there would be more, such as testing a failing input for each individually. Tests generally follow this format: Create something, run the function on it, then validate errors (or lack thereof), and validate values if necessary. Some could be improved or expanded upon, e.g. tests which call functions that return JSON. Right now the JSON is not validated in the tests.