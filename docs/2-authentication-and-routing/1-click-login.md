# 1-Click Interactive Login Window (`--login`)

The **1-Click Interactive Sign-In Window** (`./bob-gemini-free --login`) is a
best-effort local browser flow for capturing the session values needed by some
Google web capabilities. Login success and downstream entitlements must still
be verified on the target device/account.

---

## 🌟 Why Use 1-Click Login?

- 🚫 **No Developer Tools needed**
- 🚫 **No Manual Copy-Pasting headers**
- 🚫 **No Scary macOS Keychain password prompts**
- 🚫 **Works across macOS, Windows, and Linux**
- 🔒 **Locks `cookie.txt` with POSIX `0600` permissions**

---

## 🚀 How It Works

```
┌─────────────────────────────────────────────────────────────┐
│  🌐 Sign in to Google Gemini (BOB Login Window)             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                    Google                                   │
│                    Sign in with Google                      │
│                    [ your-email@gmail.com ]                 │
│                    [ Enter Password / Passkey ]             │
│                                                             │
│                    [ Next ]                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼ (User logs in once)
┌─────────────────────────────────────────────────────────────┐
│  [✔] Required session values detected                       │
│  [✔] Saved to ./cookie.txt and ~/.config/bob-gemini-free/   │
│  [i] Provider capabilities remain session-dependent         │
└─────────────────────────────────────────────────────────────┘
```

1. Run the login command:
   ```bash
   ./bob-gemini-free --login
   ```
2. A standalone Google Gemini window opens automatically.
3. Sign into your Google account (supports standard passwords, 2FA prompts, SMS, and Passkeys).
4. The moment login completes and reaches `gemini.google.com/app`:
   - BOB captures the configured session values through the Chrome DevTools Protocol (CDP).
   - Validates the required auth fields when they are present.
   - Saves the cookie string to `./cookie.txt` and `~/.config/bob-gemini-free/cookie.txt` with locked `0600` permissions.
   - Closes the browser window after the capture attempt.
5. Pro routing, vision uploads, Imagen, and Nano Banana remain provider/session-dependent; run the live conformance checks before relying on them.
