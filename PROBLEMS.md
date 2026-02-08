When actually running a query in the home (http://localhost:8090/):
Error: Failed to construct 'Worker': Script at 'https://cdn.jsdelivr.net/npm/@duckdb/duckdb-wasm@latest/dist/duckdb-browser-mvp.worker.js' cannot be accessed from origin 'http://localhost:8090'.

---

2026/02/07 21:51:46 "GET http://localhost:8090/scoreboard HTTP/1.1" from [::1]:41316 - 500 27B in 969.721µs

---

Basically none of the pages work, likely because of the features requiring DuckDB to work, but it did not. That being the case, I cannot login/register, and none of the menus shows any data in the SQL console.

---

Code doesn't handle signals gracefully (e.g. Ctrl + C when in `task run` execution)

---

TO Dos (I'll let it be at your discretion to decide when to implement (phase 2, 3, later etc))(PS: Some of these might already be mapped for later phases):

- Change the logs to be JSON structured instead of the apache style you used. Apply this to all kinds of logs in the platform, with a origin from where in the code generated the log.
- Add a button in the UI to switch between dark and light themes
- Generate an Open API Spec/swagger
- Instead of creating a fixed password in the docker, create a new one and output to stdout every time
- I feel the @DOCKER.md file is too big but provide too little value. I think most of it could be removed or consolidated in one single document about usage in general. Also, looking at all the docs, it seems that there are a few that say things that are very similar. Make docs that are very clear, but more concise. I dont want users reading for 15+ minutes to be able to fully understand how things work. Also, I miss a documentation that details more about configuration.
- Add for the latest phase a way for markdown files become full blown challenges/question pages instead of a simple text. This would make it so that an entire CTF platform could run in it, instead of relying in other things for the content itself.
- Make it so if I have version 1.1.0 and want to update, migrations will run automatically so the user will not have problems updating from one version to another. Also add an automatic backup before automatic migrations.
- One of the features states "duplicate solve prevention", does that mean that no 2 people will be able to solve the same challenge? If so, that is a problem and should be removed (or at least configurable), e.g., it will not only serve competitions but be up 24/7 serving my blog's content, so 2 different viewers should be able to solve the same challenge, so they can learn (instead of competing)
- `POST /api/auth/logout` - User logout - This doesnt seem to be an API that should be able to be called publicly
- Change docs recommendation from using nginx and use Caddy instead
- Implement a metrics system. The idea is that we'll be able to plug it in a observability system like clickstack/datadog and understand metrics about how people use the system, performance and any other thing that might make sense for production deployments.
- Add a Bruno (https://www.usebruno.com/) collection with all the APIs to be tested with a specific API client, but still being fully as code.
- Add some kind of metric for page load times for example, and with that create tests that understand if there were regression in page load times
- Make it so that if someone doesnt want to interact with the SQL parts of the platform, they can pass a queryParam that will not download the WASM duckdb to save bandwidth, and with that no page will download it.
- Add some kind of stress tests (perhaps with k6?) to understand until where the application will go healthy
- Create a command in the binary to create new admins or update accounts (e.g., change password)
- Add a way to inject code into every page (e.g. Adding Umami tracking javascript code on header/footer for all pages)
