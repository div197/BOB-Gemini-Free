# 1-Click Interactive Login Window (`--login`)

The **1-Click Interactive Sign-In Window** (`./bob-gemini-free --login`) provides a completely automated, zero-friction session setup for non-technical users and developers.

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
│  [✔] Verified 19 session tokens!                            │
│  [✔] Saved to ./cookie.txt and ~/.config/bob-gemini-free/   │
│  [✔] Gemini Pro model (gemini-3.1-pro) & Imagen 3 unlocked! │
└─────────────────────────────────────────────────────────────┘
```

1. Run the login command:
   ```bash
   ./bob-gemini-free --login
   ```
2. A standalone Google Gemini window opens automatically.
3. Sign into your Google account (supports standard passwords, 2FA prompts, SMS, and Passkeys).
4. The moment login completes and reaches `gemini.google.com/app`:
   - BOB Gemini Free intercepts all 19+ session tokens via the Chrome DevTools Protocol (CDP).
   - Validates `SAPISID`, `__Secure-1PSIDTS`, `__Secure-1PSID`, `__Secure-3PSID`, and `SID`.
   - Saves the cookie string to `./cookie.txt` and `~/.config/bob-gemini-free/cookie.txt` with locked `0600` permissions.
   - Gracefully closes the browser window.
5. Pro routing (`gemini-3.1-pro`), Multimodal Image Vision Analysis (Scotty storage uploads), Imagen 3, and Gemini Nano Banana are instantly active!
