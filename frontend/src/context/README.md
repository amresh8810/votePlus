# context/

React Context providers for global application state.

## Planned contexts (Phase 2+)

| File | Provider | State held |
|------|----------|-----------|
| `AuthContext.jsx` | `AuthProvider` | `user`, `token`, `login()`, `logout()` — JWT stored in `localStorage` and refreshed on mount |

### Usage pattern

```jsx
// Wrap the entire app (in main.jsx):
<AuthProvider>
  <App />
</AuthProvider>

// Consume anywhere:
const { user, token, logout } = useContext(AuthContext);
```
