Develop a Golang client library that communicates with the FastAPI prompt-refinement service. This library should provide:

- A function, e.g., `RefinePrompt(lazyPrompt string) (string, error)`, which sends a POST request to the /refine-prompt endpoint of the API.
- Configuration options for specifying the API base URL and HTTP timeouts.
- Robust error handling, including retry logic if the API call fails.
- JSON marshalling/unmarshalling to handle the request and response payloads.
- Clear documentation of the package and the primary function(s) so that users understand how to integrate it into their projects.

Ensure the library is well-structured, uses idiomatic Go patterns, and includes basic unit tests to verify its behavior.
