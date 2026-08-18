# Multi-Account Profile Handling (`auth_user`)

When multiple Google accounts are signed into the same browser session (e.g. Profile 0 = Personal, Profile 1 = Work/Workspace), Google assigns an integer index to each account.

---

## 🔍 How to Determine Your Account Index

Look at the URL in your browser when visiting Google Gemini:

| Gemini Web URL | `auth_user` Index | Account Type |
| :--- | :--- | :--- |
| `https://gemini.google.com/app` | `"0"` (default) | Primary profile |
| `https://gemini.google.com/u/1/app` | `"1"` | Secondary profile |
| `https://gemini.google.com/u/2/app` | `"2"` | Tertiary profile |

---

## ⚙️ Configuring `auth_user`

### In `config.json`
```json
{
  "auth_user": "1",
  "cookie_file": "./cookie.txt"
}
```

### Via Environment Variable
```bash
export BOB_GEMINI_FREE_AUTH_USER="1"
./bob-gemini-free
```

### What BOB Does Behind the Scenes
1. Dispatches `X-Goog-AuthUser: 1` header in upstream RPC calls.
2. Prefixes all request paths with `/u/1/` (e.g. `https://gemini.google.com/u/1/_/BardChatUi/...`).
3. Ensures requests route to the exact authenticated profile and Pro subscription tier.
