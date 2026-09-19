# services/

Axios-based API client wrappers. Each file corresponds to a backend resource.
All functions return Promises and handle auth headers automatically.

## Planned service files (Phase 4+)

| File | Exports | Backend route |
|------|---------|---------------|
| `api.js` | Configured Axios instance (base URL, interceptors) | — |
| `authService.js` | `register()`, `login()`, `logout()` | `/api/v1/auth/*` |
| `pollService.js` | `createPoll()`, `getPoll()`, `listMyPolls()`, `deletePoll()` | `/api/v1/polls/*` |
| `voteService.js` | `submitVote()` | `/api/v1/polls/:id/vote` |
| `resultsService.js` | `getResults()` | `/api/v1/polls/:id/results` |
