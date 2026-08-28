# Android & Mobile Deployment Guide (`pkg/mobile`)

**BOB Gemini Free** supports running natively on **Android (ARM64)** and **iOS (Swift / Apple Silicon)** via the embedded `pkg/mobile` package and `gomobile bind`.

This architecture turns any standard mobile phone or tablet into a **sovereign, self-contained AI coding workstation** with zero cloud API billing and zero external framework bloat.

---

## 📱 Mobile Architecture Overview

```
                                  ┌─────────────────────────────────────────────────────────┐
                                  │               MOBILE APP (Android / iOS)                │
                                  ├─────────────────────────────────────────────────────────┤
                                  │  UI Layer: Jetpack Compose (Kotlin) / SwiftUI / Flutter │
                                  └──────────────────────────┬──────────────────────────────┘
                                                             │
                                   ┌─────────────────────────┴─────────────────────────┐
                                   │                                                   │
                                   ▼                                                   ▼
                    ┌──────────────────────────────┐                   ┌──────────────────────────────┐
                    │   1-Tap Native Login Sheet   │                   │  Embedded Core (`gomobile`)  │
                    │   (WebView / WKWebView)      │                   │  • In-Memory Stream Deltas   │
                    │   • CookieManager (Android)  │ ──(Extracted)───► │  • Zero Socket Latency (0ms) │
                    │   • WKCookieStore (iOS)      │    Cookies        │  • 4-Stage Invariant Refiner │
                    │   • Hardware Keystore AES-256│                   │  • BPE Subword Tokenizer     │
                    └──────────────────────────────┘                   └───────────────┬──────────────┘
                                                                                       │
                                                                                       ▼ (Direct HTTPS)
                                                                       ┌──────────────────────────────┐
                                                                       │    Google Gemini Web RPC     │
                                                                       │    • streamGenerateContent   │
                                                                       │    • batchexecute            │
                                                                       └──────────────────────────────┘
```

---

## 🤖 Android Integration (Kotlin & Jetpack Compose)

### 1. Build the Android Archive (`.aar`)
Compile the `pkg/mobile` package into a standard Android `.aar` library:

```bash
# Prerequisites: Go 1.26+ and Android NDK
go install golang.org/x/mobile/cmd/gomobile@latest
gomobile init

# Compile BOB Mobile core
gomobile bind -target=android -o bob-mobile.aar ./pkg/mobile
```

Place `bob-mobile.aar` into your Android project's `app/libs/` folder.

---

### 2. Boot Gateway In-Process in Android (`MainActivity.kt`)

```kotlin
package com.abcsteps.bobgeminifree

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import mobile.Mobile
import mobile.MobileGateway
import mobile.StreamCallback

class MainActivity : ComponentActivity() {
    private val gateway: MobileGateway = Mobile.getDefaultGateway()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // 1. Start in-process gateway on 127.0.0.1 (random high port or 9610)
        val endpointUrl = gateway.start(9610, "127.0.0.1", "")

        setContent {
            MobileStudioScreen(gateway = gateway)
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        gateway.stop()
    }
}
```

---

### 3. 1-Tap Google Sign-In with Android `CookieManager`

Android provides programmatic access to web session cookies without requiring developer flags:

```kotlin
import android.webkit.CookieManager
import android.webkit.WebView
import android.webkit.WebViewClient

fun setupGoogleLoginWebView(webView: WebView, onSessionExtracted: (String) -> Unit) {
    webView.settings.javaScriptEnabled = true
    webView.webViewClient = object : WebViewClient() {
        override fun onPageFinished(view: WebView?, url: String?) {
            super.onPageFinished(view, url)
            if (url?.contains("gemini.google.com") == true) {
                val cookies = CookieManager.getInstance().getCookie("https://gemini.google.com")
                if (cookies != null && cookies.contains("__Secure-1PSID")) {
                    onSessionExtracted(cookies)
                }
            }
        }
    }
    webView.loadUrl("https://gemini.google.com/app")
}
```

---

### 4. Real-Time Streaming in Jetpack Compose

```kotlin
fun streamResponse(prompt: String, onTextUpdate: (String) -> Unit) {
    val callback = object : StreamCallback {
        override fun onDelta(chunk: String) {
            runOnUiThread {
                onTextUpdate(chunk)
            }
        }

        override fun onComplete(totalTokens: Long, errStr: String) {
            if (errStr.isNotEmpty()) {
                println("Stream error: $errStr")
            }
        }
    }

    Thread {
        gateway.generateStream(prompt, "gemini-3.7-flash", callback)
    }.start()
}
```

---

## 🛠️ Mobile IDE Integration (Acode & Termux)

Students on Android can use BOB directly inside mobile code editors:

### Acode Editor Configuration:
1. Open **Acode** $\rightarrow$ **Settings** $\rightarrow$ **AI Assistant Plugin**.
2. Set **Base URL**: `http://127.0.0.1:9610/v1`
3. Set **API Key**: `none`
4. Set **Model**: `gemini-3.7-flash` or `gemini-3.7-flash-thinking`

### Termux CLI Configuration:
```bash
export OPENAI_BASE_URL=http://127.0.0.1:9610/v1
export OPENAI_API_KEY=none
export ANTHROPIC_BASE_URL=http://127.0.0.1:9610
export ANTHROPIC_API_KEY=none

# Test instant CLI completion on Android
curl -s http://127.0.0.1:9610/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "gemini-3.7-flash", "messages": [{"role": "user", "content": "Hello from Android Termux!"}]}'
```

---

## 🍏 iOS Integration (Swift & SwiftUI)

1. Compile the Apple `XCFramework`:
   ```bash
   gomobile bind -target=ios -o BOBMobile.xcframework ./pkg/mobile
   ```
2. Drag `BOBMobile.xcframework` into your Xcode project.
3. Import and initialize in SwiftUI:
   ```swift
   import SwiftUI
   import BOBMobile

   @main
   struct BOBApp: App {
       let gateway = MobileGetDefaultGateway()!

       init() {
           try? gateway.start(9610, host: "127.0.0.1", cookieContent: "")
       }

       var body: some Scene {
           WindowGroup {
               ContentView(gateway: gateway)
           }
       }
   }
   ```
