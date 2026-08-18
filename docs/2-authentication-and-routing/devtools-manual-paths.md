# Manual DevTools Cookie Extraction (3 Visual Paths)

For headless servers, remote cloud VMs, or power users who prefer extracting cookies directly from their existing browser session.

---

## 3 Visual Extraction Paths

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Chrome DevTools (F12 or Cmd+Opt+I)                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│  [Network] Tab  Filter: [ app                   ] [X] Preserve log          │
│  ─────────────────────────────────────────────────────────────────────────  │
│  Name                                  Status   Type      Size              │
│  📄 app?eom=1&awwd=1&em=2&...          200      document  22.2 kB  <── CLICK│
│  ⚙️ batchexecute?rpcids=...            200      xhr       14.5 kB           │
│  ─────────────────────────────────────────────────────────────────────────  │
│  [Headers] [Payload] [Preview] [Response]                                   │
│  ▼ Request Headers                                                          │
│    :authority: gemini.google.com                                            │
│    Cookie: __Secure-BUCKET=...; SID=...; SAPISID=...;  <── RIGHT CLICK & COPY│
└─────────────────────────────────────────────────────────────────────────────┘
```

### Path A: The Instant 1-Click Method (No Chat Needed)
1. Open [gemini.google.com](https://gemini.google.com) in Chrome, Arc, Edge, or Brave.
2. Press **`F12`** (or **`Cmd + Option + I`** on macOS) to open Developer Tools.
3. Click the **Network** tab at the top.
4. Refresh the page (**`Cmd + R`** or **`F5`**).
5. Click on the document request named **`app?eom=1...`** (or any top **`batchexecute`** request).
6. In the right panel, select **Headers** and scroll to **Request Headers**.
7. Right-click on the **`cookie:`** value and click **Copy value**.

---

### Path B: The 1-Word Chat Method (`StreamGenerate`)
1. In DevTools **Network** tab, type **`StreamGenerate`** in the filter search box.
2. In Gemini, type any 1-word prompt (e.g. *"hi"*) and press Enter.
3. The request **`StreamGenerate`** will instantly appear in the Network list.
4. Click **`StreamGenerate`** $\rightarrow$ **Headers** $\rightarrow$ **Request Headers** $\rightarrow$ right-click **`cookie:`** $\rightarrow$ **Copy value**.

---

### Path C: Application Storage Tab
1. In DevTools, click the **Application** tab (or click `>>` to find Application).
2. Expand **Storage** $\rightarrow$ **Cookies** $\rightarrow$ select `https://gemini.google.com`.
3. Highlight all cookie rows and copy.

---

## Ingesting the Copied Cookie

### Method 1: Automated Interactive Helper
```bash
./bob-gemini-free --setup-cookie
```
Paste the string and press **Enter**.

### Method 2: Zero-Config Local `cookie.txt`
```bash
cat << 'EOF' > cookie.txt
PASTE_YOUR_COOKIE_STRING_HERE
EOF
chmod 600 cookie.txt
./bob-gemini-free
```
