<p align="center">
  <img src="assets/bob-gemini-free-banner.jpg" alt="BOB Gemini Free Banner" width="100%">
</p>

<h1 align="center">BOB Gemini Free (बॉब जेमिनी फ्री)</h1>

<p align="center">
  <strong>यूनिवर्सल 3-इन-1 AI गेटवे इंजन</strong><br>
  <em>डेवलपर्स और एजेंट्स के लिए ड्रॉप-इन OpenAI, Anthropic, और Google Gemini API</em>
</p>

<p align="center">
  <a href="https://abcsteps.com/"><img src="https://img.shields.io/badge/Powered%20by-ABCsteps.com-2563eb?style=flat-square" alt="ABCsteps"></a>
  <a href="https://github.com/div197/bob-gemini-free"><img src="https://img.shields.io/badge/Author-Divyanshu%20Singh%20Chouhan-16a34a?style=flat-square" alt="Author"></a>
  <img src="https://img.shields.io/badge/Release-v0.1.2-7c3aed?style=flat-square" alt="Release">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/Protocols-OpenAI%20%7C%20Anthropic%20%7C%20Gemini-059669?style=flat-square" alt="Protocols">
  <img src="https://img.shields.io/badge/License-MIT-f59e0b?style=flat-square" alt="License">
</p>

<p align="center">
  <a href="README.md"><strong>English Documentation</strong></a> &nbsp;•&nbsp;
  <a href="README.hi.md"><strong>हिंदी गाइड (Hindi)</strong></a> &nbsp;•&nbsp;
  <a href="CHANGELOG.md"><strong>Changelog</strong></a>
</p>

---

**BOB Gemini Free** (बॉब जेमिनी फ्री), [**ABCsteps.com**](https://abcsteps.com/) — जोधपुर, राजस्थान (भारत) में **दिव्यांशु सिंह चौहान** ([@div197](https://github.com/div197)) द्वारा स्थापित ऑनलाइन एआई इंजीनियरिंग स्कूल — की **BOB सीरीज़** (*Break Ordinary Boundaries*) का एक प्रमुख उत्पाद है।

---

## 🌟 दर्शन: "Break Ordinary Boundaries" क्यों?

आज के दौर में एआई सीखने वाले छात्रों, स्वतंत्र रचनाकारों और डेवलपर्स के सामने **तीन साधारण सीमाएँ (Ordinary Boundaries)** आती हैं:

1. 💸 **आर्थिक सीमा (महँगी API कीमतें)**: उन्नत एआई मॉडल्स के लिए महँगे क्रेडिट कार्ड्स और $20-$100 प्रति माह के बिल चुकाने पड़ते हैं।
2. 🔒 **इकोसिस्टम सीमा (प्लेटफ़ॉर्म बंधन)**: एंथ्रोपिक टूल्स केवल एंथ्रोपिक से बात करते हैं, ओपनएआई केवल अपने से, और जेमिनी का मुफ़्त वेब एक्सेस ब्राउज़र टैब तक सीमित रहता है।
3. ⚙️ **जटिलता की सीमा (इंस्टॉलेशन का झंझट)**: प्रॉक्सी सेटअप करने के लिए पायथन, गो, नोड या जटिल टूल्स इंस्टॉल करने पड़ते हैं।

**BOB इन तीनों सीमाओं को एक साथ तोड़ता है**:
- ✨ **शून्य लागत (Zero Cost)**: हर गूगल खाते को **Gemini 3.7 Flash**, **Flash Thinking**, **3.1 Pro**, **Imagen 3**, और **Gemini Nano Banana** का मुफ़्त उपयोग देता है।
- 🔓 **"API-Less AI" आर्किटेक्चर**: न कोई क्रेडिट कार्ड, न क्लाउड बिलिंग खाता, और न API Key चोरी होने का कोई डर। आपका लोकल सेशन ही पूरी सुरक्षा से सब कुछ संभालता है।
- 🌉 **यूनिवर्सल 4-इन-1 प्रोटोकॉल**: एक ही सिंगल गेटवे से **OpenAI** (`/v1/chat/completions`, `/v1/responses`, `/v1/tokens/count`), **Anthropic Messages (Claude Code CLI)**, **Google Gemini v1beta** (`:countTokens`), और **Embedded Go लाइब्रेरी** (`pkg/gateway`) का सीधा अनुवाद।
- ⚡ **शून्य निर्भरता (Zero Dependencies)**: न Go, न Python, न कोई रनटाइम — बस एक सिंगल बाइनरी और **1-क्लिक लॉगिन (`--login`)**।

---

## 🚀 "API-Less AI" क्रांति: डेवलपर्स के लिए सच्ची आज़ादी

पारंपरिक क्लाउड कंपनियाँ डेवलपर्स को क्रेडिट कार्ड और प्रति-टोकन बिलिंग के चक्रव्यूह में उलझा देती हैं। **BOB Gemini Free पेश करता है बिना-API का मुफ़्त एआई मॉडल**:

| पारंपरिक क्लाउड API मॉडल | BOB का API-Less आर्किटेक्चर |
| :--- | :--- |
| 💳 क्रेडिट कार्ड और क्लाउड बिलिंग खाता अनिवार्य | **₹0.00 / शून्य क्रेडिट कार्ड की ज़रूरत** |
| 💸 हर मिलियन टोकन और रीज़निंग स्टेप का महँगा बिल | **Flash और Deep Thinking पर असीमित कोडिंग** |
| 🔑 API Key लीक होने पर लाखों का नुकसान | **सुरक्षित लोकल सेशन (`0600` फ़ाइल परमिशन)** |
| 🔒 किसी एक कंपनी के CLI या टूल में क़ैद | **यूनिवर्सल 4-इन-1 ट्रांसलेशन (OpenAI, Claude, Google, Go)** |
| 📊 महीने के अंत में चौंकाने वाला इनवॉइस | **स्क्रीन पर लाइव टोकन व डॉलर बचत ट्रैकिंग (`GET /`)** |

---

<p align="center">
  <img src="assets/bob-gemini-free-universal-gateway.png" alt="BOB Gemini Free यूनिवर्सल AI गेटवे आर्किटेक्चर" width="100%">
</p>

---

## 💡 यह कैसे काम करता है? (सरल हिंदी में)

**BOB Gemini Free** को अपने कंप्यूटर पर चलने वाले एक तेज़, सुरक्षित और प्राइवेट **यूनिवर्सल ट्रांसलेटर** की तरह समझें:

```
┌─────────────────────────────────────────────────────────────┐
│  डीप एजेंटिक कोडिंग टूल्स व स्वायत्त फ़्रेमवर्क्स           │
│  (Codex CLI, Claude Code CLI, Cursor, Grok Build, Agent SDK)│
└──────────────────────────────┬──────────────────────────────┘
                               │ OpenAI / Anthropic स्टैंडर्ड में बात करता है
                               ▼
┌─────────────────────────────────────────────────────────────┐
│  ⚡ BOB Gemini Free (आपके कंप्यूटर पर चलने वाला लोकल गेटवे)  │
│  लोकल और सुरक्षित रूप से रिक्वेस्ट ट्रांसलेट करता है         │
└──────────────────────────────┬──────────────────────────────┘
                               │ Google Web RPC स्ट्रीम में बात करता है
                               ▼
┌─────────────────────────────────────────────────────────────┐
│  🌐 Google Gemini Web (Flash 3.7 / Thinking / Pro / Images)  │
└─────────────────────────────┘
```

जब **OpenAI Codex CLI**, **Claude Code CLI**, **Cursor**, या **Grok Build** किसी जवाब की माँग करता है, तो BOB उसे जेमिनी के वेब फ़ॉर्मेट में बदलता है, पूरा रीज़निंग और विज़न उत्तर लाता है, और मिलीसेकंड्स में वापस सौंप देता है।

---

## BOB सीरीज़ क्या है? (*Break Ordinary Boundaries*)

ABCsteps की **BOB सीरीज़** का उद्देश्य डेवलपर्स, छात्रों और एआई इंजीनियरों के लिए बिना किसी पेवॉल या महँगे सब्सक्रिप्शन के शक्तिशाली टूल्स और रनटाइम्स उपलब्ध कराना है:

* ⚡ [**BOB Gemini Free**](https://github.com/div197/bob-gemini-free) — OpenAI, Anthropic, Google REST और Go SDK को सपोर्ट करने वाला यूनिवर्सल 4-इन-1 गेटवे।
* 🎥 [**BOB YouTube**](https://github.com/div197/BOB-Youtube) — डेवलपर्स और एआई एजेंट्स के लिए डॉकर-फर्स्ट यूट्यूब डेटा इंजेक्शन और प्रोसेसिंग टूल।

---

## मुख्य विशेषताएँ

* **प्रत्येक जीमेल (Gmail) यूज़र के लिए मुफ़्त**: दुनिया के हर व्यक्ति के पास जीमेल के साथ जेमिनी का मुफ़्त एक्सेस होता है। बिना किसी अतिरिक्त शुल्क के Flash 3.7, Flash Lite और Deep Thinking (20,000+ अक्षरों की रीज़निंग) का उपयोग करें।
* **यूनिवर्सल 3-इन-1 प्रोटोकॉल**: OpenAI SDKs, Claude Code CLI, Anthropic SDKs, Codex CLI, और Google GenAI SDKs के साथ सीधे कम्पैटिबल।
* **1-क्लिक नेटिव लॉगिन (`--login`)**: बिना DevTools या चाबी (Keychain) पॉपअप के एक क्लिक में गूगल अकाउंट जोड़ें।
* **मल्टी-अकाउंट कुकी पूल (`cookie_pool`)**: कई गूगल खातों में लोड-बैलेंसिंग और 429 एरर पर ऑटोमैटिक फ़ेलओवर।
* **Gemini Advanced ($20/माह) प्रो अनलॉक**: लोकल कुकी जोड़कर गूगल के फ्लैगशिप **Pro** (`gemini-3.1-pro`) और **Imagen 3 / Nano Banana** को सक्रिय करें।
* **Claude 3.7 / 3.5 नेटिव थिंकिंग**: Claude Code CLI के लिए थिंकिंग ब्लॉक्स का सीधा समर्थन।
* **मल्टीमॉडल विज़न (Vision)**: OpenAI फॉर्मेट में Base64 इमेज या इमेज लिंक्स भेजें — ऑटोमैटिक कम्प्रेशन के साथ।
* **शून्य लागत, पूर्ण प्राइवेसी**: सिंगल स्टैटिक बाइनरी, <15MB मेमोरी, और ज़ीरो टेलीमेट्री (`127.0.0.1`)।
## समर्थित टूल्स और इकोसिस्टम (Supported Tools & Ecosystem)

BOB Gemini Free बिना किसी अतिरिक्त कॉन्फ़िगरेशन के सभी प्रमुख AI कोडिंग टूल्स और एजेंटिक फ़्रेमवर्क्स के साथ काम करता है:

| श्रेणी (Category) | समर्थित टूल्स व क्लाइंट्स | कनेक्शन एंडपॉइंट |
| :--- | :--- | :--- |
| **टर्मिनल कोडिंग एजेंट्स** | OpenAI Codex CLI (`codex`), Claude Code CLI (`claude`), Gemini CLI (`gemini`) | नेटिव बेस यूआरएल (Base URLs) |
| **एजेंटिक IDEs व एक्सटेंशन्स** | Cursor (Agent Mode), Windsurf, VS Code (Continue, Roo Code, Cline) | `http://127.0.0.1:8081/v1` |
| **स्वायत्त एजेंट फ़्रेमवर्क्स** | Grok Build, LangChain, CrewAI, AutoGen, OpenAI Agents SDK | `http://127.0.0.1:8081/v1` |
| **रूटर्स व लोकल प्रॉक्सीज़** | LiteLLM, OneAPI, NewAPI, Portkey, OpenRouter | `http://127.0.0.1:8081/v1` |
| **ऑफिशियल SDKs** | OpenAI (Python/JS/Go/.NET/Java), Anthropic (Python/TypeScript), Google GenAI | लोकल बेस यूआरएल |

---

## त्वरित शुरुआत (Quick Start)

### सबसे आसान तरीका (Super Simple - बिना किसी कॉन्फ़िग के)

```bash
# 1. गेटवे शुरू करें
./bob-gemini-free

# 2. किसी भी टूल में सेट करें:
# Base URL: http://127.0.0.1:8081/v1
# API Key:  none
```

---

### ऑटोमैटिक 13-पॉइंट टेस्ट किट (Diagnostic Test Kit)

सभी 13 मॉडल्स, स्ट्रीमिंग, टोकन काउंटिंग, और API फ़ॉर्मेट्स की लाइव जाँच के लिए:

```bash
# डिफ़ॉल्ट लोकल गेटवे का परीक्षण
./bob-gemini-free --test

# या कस्टम पोर्ट / API Key के साथ
./bob-gemini-free --test --test-url http://127.0.0.1:8081 --test-key your_api_key

# या स्क्रिप्ट चलाएँ
./test-kit.sh
```

### लाइव स्टेटस व डॉलर बचत ट्रैकर (`--status`)

टर्मिनल में लाइव अनुरोध, टोकन थ्रूपुट और डॉलर बचत की वास्तविक जानकारी देखें:

```bash
./bob-gemini-free --status
```

---

### विकल्प A: डॉकर (Docker) द्वारा चलाएँ

```bash
docker compose up -d
```

---

### विकल्प B: ऑटोमैटिक इंस्टॉलर स्क्रिप्ट (macOS / Linux)

```bash
chmod +x install.sh
./install.sh
```

---

### विकल्प C: इंटरैक्टिव वेब प्लेग्राउंड व टेलीमेट्री डैशबोर्ड (`/playground` व `/ui`)

अपने किसी भी वेब ब्राउज़र में `http://127.0.0.1:8081/playground` खोलें और इनबिल्ट विज़ुअल इंटरफ़ेस का आनंद लें:

* 🧠 **लाइव रीज़निंग विज़ुअलाइज़र**: Gemini 3.7 Flash Thinking के रीज़निंग टोकन्स को रीयल-टाइम में प्रवाहित होते हुए देखें।
* ⚡ **मॉडल व थिंकिंग स्विचर**: तेज़ मॉडल्स, डीप रीज़निंग, या Imagen 3 इमेज सिंथेसिस को एक क्लिक में टेस्ट करें।
* 📊 **लाइव टेलीमेट्री व डॉलर बचत ट्रैकर**: अपटाइम, प्रोसेस्ड टोकन्स और कुल डॉलर बचत का लाइव मीटर।
* 📋 **मल्टी-प्रोटोकॉल स्निपेट जनरेटर**: Python, Claude Code CLI, और cURL कोड स्निपेट्स को तुरंत कॉपी करें।

<p align="center">
  <img src="assets/bob-gemini-free-playground.png" alt="BOB Gemini Free वेब प्लेग्राउंड व टेलीमेट्री डैशबोर्ड" width="100%">
</p>

---

### विकल्प D: सोर्स कोड से निर्माण (Build from Source - Go 1.22+)

```bash
make build
./bob-gemini-free --port 8081
```

---

## 📂 मल्टी-लैंग्वेज कोड उदाहरण और SDK इंटीग्रेशन

विभिन्न भाषाओं के रेडी-टू-रन कोड उदाहरण [`examples/`](examples/) डायरेक्टरी में उपलब्ध हैं:

* 🐍 **Python**:
  * [`examples/python/openai_chat.py`](examples/python/openai_chat.py) — OpenAI SDK द्वारा रीज़निंग टोकन स्ट्रीमिंग।
  * [`examples/python/anthropic_messages.py`](examples/python/anthropic_messages.py) — Anthropic SDK Messages API और एक्सटेंडेड थिंकिंग।
* 🟨 **Node.js / TypeScript**:
  * [`examples/nodejs/openai_chat.mjs`](examples/nodejs/openai_chat.mjs) — OpenAI npm SDK स्ट्रीम कंस्यूमर।
  * [`examples/nodejs/anthropic_messages.mjs`](examples/nodejs/anthropic_messages.mjs) — `@anthropic-ai/sdk` Messages API क्लाइंट।
* 🔷 **Go (एंबेडेड इंजन)**:
  * [`examples/go/embedded_sdk.go`](examples/go/embedded_sdk.go) — डायरेक्ट इन-प्रोसेस Go प्रोग्रामैटिक इन्फरेंस (`pkg/gateway.NewEngine()`)।
* 🐚 **cURL व शेल स्क्रिप्ट्स**:
  * [`examples/curl/chat.sh`](examples/curl/chat.sh) — स्टैंडर्ड चैट कंप्लीशन।
  * [`examples/curl/stream_thinking.sh`](examples/curl/stream_thinking.sh) — रीयल-टाइम रीज़निंग स्ट्रीम।
  * [`examples/curl/anthropic.sh`](examples/curl/anthropic.sh) — Anthropic messages एंडपॉइंट।
  * [`examples/curl/responses_codex.sh`](examples/curl/responses_codex.sh) — OpenAI Codex CLI `/v1/responses`।

---

## क्लाइंट में कैसे जोड़ें?

### OpenAI Python SDK

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:8081/v1",
    api_key="none"
)

response = client.chat.completions.create(
    model="gemini-3.5-flash-thinking",
    messages=[
        {"role": "user", "content": "क्वांटम कंप्यूटिंग को आसान हिंदी में समझाइए।"}
    ]
)
print(response.choices[0].message.content)
```

### मल्टीमॉडल विज़न (Image Input)

```python
import base64
from openai import OpenAI

client = OpenAI(base_url="http://127.0.0.1:8081/v1", api_key="none")

with open("image.png", "rb") as f:
    b64 = base64.b64encode(f.read()).decode("utf-8")

response = client.chat.completions.create(
    model="gemini-3.6-flash",
    messages=[
        {
            "role": "user",
            "content": [
                {"type": "text", "text": "इस चित्र में क्या है?"},
                {"type": "image_url", "image_url": {"url": f"data:image/png;base64,{b64}"}}
            ]
        }
    ]
)
print(response.choices[0].message.content)
```

### Claude Code CLI और Anthropic SDK (सीधा सपोर्ट)

BOB Gemini Free में Anthropic Messages API प्रोटोकॉल (`POST /v1/messages`) का इनबिल्ट नेटिव सपोर्ट है।

#### Claude Code CLI सेटअप

```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:8081
export ANTHROPIC_API_KEY=none
claude
```

#### Anthropic Python SDK

```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://127.0.0.1:8081",
    api_key="none"
)

message = client.messages.create(
    model="gemini-3.7-flash-thinking",
    max_tokens=4096,
    messages=[
        {"role": "user", "content": "एक समवर्ती Go वर्कर पूल लिखें।"}
    ]
)
print(message.content[0].text)
```

### OpenAI Codex CLI

```bash
export OPENAI_BASE_URL=http://127.0.0.1:8081/v1
export OPENAI_API_KEY=none
codex
```

### OpenAI इमेज जेनरेशन (`POST /v1/images/generations`)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:8081/v1",
    api_key="none"
)

response = client.images.generate(
    model="dall-e-3",
    prompt="A futuristic neon cybernetic logo for BOB Gemini Free, digital art, 8k",
    size="1024x1024",
    n=1
)
print("Generated Image URL:", response.data[0].url)
```

> [!NOTE]
> **Imagen इमेज जेनरेशन प्रमाणीकरण**: गूगल इमेज जेनरेशन के लिए सक्रिय रूप से साइन-इन किए गए गूगल खाते की मांग करता है। अनाम (बिना कुकी) मोड में चलने पर जेमिनी नीति स्पष्टीकरण लौटाता है। इमेज जेनरेशन का उपयोग करने के लिए `--cookie-file cookie.txt` या `--setup-cookie` द्वारा अपना सेशन कुकी जोड़ें।

### Go पैकेज / इन-प्रोसेस लाइब्रेरी इम्पोर्ट (Embedded Go Library)

BOB Gemini Free को अपने किसी भी Go प्रोजेक्ट, माइक्रो-सर्विस या AI एजेंट में सीधे इन-प्रोसेस लाइब्रेरी के रूप में इम्पोर्ट करें:

```go
package main

import (
	"net/http"
	"github.com/div197/bob-gemini-free/pkg/gateway"
)

func main() {
	handler := gateway.NewHandler(
		gateway.WithDefaultModel("gemini-3.7-flash"),
		gateway.WithCookieFile("cookie.txt"), // ऐच्छिक (Optional)
	)

	http.ListenAndServe("127.0.0.1:8081", handler)
}
```

---

## विस्तृत आर्किटेक्चर तुलना: Google AI Studio बनाम BOB Gemini Free

### आधिकारिक सीमाएँ और फ़ीचर मैट्रिक्स (मुफ़्त टियर बिना पेड बिलिंग)

| पैमाना / विशेषता | Google AI Studio (Free Tier) | **BOB Gemini Free (लोकल गेटवे)** |
| :--- | :--- | :--- |
| **Flash दैनिक सीमा (RPD)** | **1,500 अनुरोध / दिन** (दैनिक सीमा पार होते ही बंद) | **व्यावहारिक रूप से असीमित इंटरैक्टिव प्रश्न** |
| **Flash प्रति मिनट अनुरोध (RPM)** | **15 RPM** (`429 RESOURCE_EXHAUSTED`) | **हाई-थ्रूपुट वेब सेशन (30+ RPM बर्स्ट)** |
| **Flash टोकन दर (TPM)** | 1,000,000 टोकन / मिनट | बफ़र्ड हाई-स्पीड थ्रूपुट |
| **Pro दैनिक सीमा (RPD)** | **50 अनुरोध / दिन** (अत्यंत सीमित) | **व्यावहारिक रूप से असीमित Pro प्रश्न** *(Advanced कुकी के साथ)* |
| **Pro प्रति मिनट अनुरोध (RPM)** | **2 RPM** (भारी थ्रॉटलिंग) | **सामान्य इंटरैक्टिव वेब कंकरेंसी** |
| **रीज़निंग / थिंकिंग गहराई** | मुफ़्त कुंजियों पर सीमित | **20,000+ अक्षरों का गहरा रीज़निंग आउटपुट** |
| **OpenAI प्रोटोकॉल सपोर्ट** | ❌ शून्य (कस्टम SDK कोड चाहिए) | **✅ 100% ड्रॉप-इन (`/v1/chat/completions`, `/v1/responses`)** |
| **Developer रोल सपोर्ट** | ❌ नहीं | **✅ पूर्ण OpenAI `developer` व `system` पार्सिंग** |
| **रीज़निंग टोकन एक्सपोर्ट** | ❌ प्रोप्राइटरी फ़ॉर्मेट | **✅ स्टैंडर्ड `reasoning_content` (Cursor/OpenWebUI में कार्ड)** |
| **डेटा प्राइवेसी व ट्रेनिंग** | ⚠️ **गूगल समीक्षकों द्वारा डेटा ट्रेनिंग के लिए लॉग किया जाता है** | **🛡️ 100% लोकल गेटवे (`127.0.0.1` पर बंधा, शून्य ट्रैकिंग)** |
| **सेटअप जटिलता** | क्लाउड कंसोल, प्रोजेक्ट निर्माण, API Key प्रबंधन | **जीरो कॉन्फ़िग: `./bob-gemini-free` चलाएँ और कोडिंग शुरू करें** |
| **कुल वित्तीय लागत** | $0 (थ्रॉटल होने तक) या महँगे पेड टोकन | **100% मुफ़्त हमेशा के लिए** |

---

### अधिकतम ऑपरेशनल सीमाएँ और अनुशंसित दिशानिर्देश

गेटवे की अधिकतम स्थिरता और शून्य थ्रॉटलिंग के लिए:

* **अधिकतम आउटपुट टोकन लंबाई**: गूगल फ़्लैश मॉडल्स प्रति प्रतिक्रिया **8,192 टोकन (~32,000 अक्षर)** तक देते हैं। थिंकिंग ट्रेस अंतिम उत्तर से पहले **20,000 अक्षर** तक का विस्तृत विश्लेषण देता है।
* **अनुकूलतम कंकरेंसी बैंडविड्थ**:
  * **Flash 3.7 / 3.6 / Flash Lite**: **3 से 5 समवर्ती स्ट्रीम्स** प्रति IP/इंस्टेंस।
  * **Deep Thinking और Pro 3.1**: **2 से 3 समवर्ती स्ट्रीम्स**।
* **ऑटोमैटिक री-ट्राई रणनीति**: इनबिल्ट एक्सपोनेंशियल बैकऑफ़ (`retry_attempts: 3`, `retry_delay_sec: 2`) क्षणिक नेटवर्क समस्याओं को स्वचालित रूप से संभालता है।
* **अधिकतम इमेज साइज़**: ऑटोमैटिक कंप्रेशन इंजन बड़ी छवियों को **1024px JPEG (क्वालिटी 75, <1MB)** में बदलकर गूगल के Scotty स्टोरेज पर अपलोड करता है।

---

## लाइव परफॉर्मेंस व स्ट्रेस बेंचमार्क

गेटवे की गति और थ्रूपुट मापने के लिए इनबिल्ट बेंचमार्क चलाएँ:

```bash
# 3 वर्कर्स और 6 प्रश्नों के साथ बेंचमार्क चलाएँ
./bob-gemini-free --bench --bench-concurrency 3 --bench-requests 6

# या स्क्रिप्ट चलाएँ
./scripts/bench.sh http://127.0.0.1:8081 3 6 your_api_key
```

---

## बैकग्राउंड डेमन और ऑटो-स्टार्ट सर्विसेज

BOB Gemini Free को अपने सिस्टम पर बैकग्राउंड में लगातार चालू रखने के लिए:

### Linux (Systemd सर्विस)

```bash
# 1. बाइनरी और सर्विस फ़ाइल कॉपी करें
sudo cp bob-gemini-free /usr/local/bin/
sudo cp scripts/bob-gemini-free.service /etc/systemd/system/

# 2. सर्विस चालू करें
sudo systemctl daemon-reload
sudo systemctl enable --now bob-gemini-free
```

### macOS (Launchd डेमन)

```bash
sudo cp bob-gemini-free /usr/local/bin/
cp scripts/com.abcsteps.bob-gemini-free.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.abcsteps.bob-gemini-free.plist
```

### Windows (बैकग्राउंड ऑटो-स्टार्ट)

```cmd
scripts\start-service.bat
```

---

## यूनिवर्सल टूल व ऍप्लिकेशन्स इंटीग्रेशन

चूँकि **BOB Gemini Free** आधिकारिक OpenAI API मानक का 100% पालन करता है, **कोई भी AI एप्लिकेशन, एजेंट फ़्रेमवर्क, या डेवलपर टूल** जो कस्टम OpenAI Base URL का समर्थन करता है, तुरंत काम करता है।

### यूनिवर्सल इंटीग्रेशन पैटर्न

किसी भी AI एप्लिकेशन में यह तीन मानक सेटिंग्स भरें:

| सेटिंग फ़ील्ड | वैल्यू | विवरण |
| :--- | :--- | :--- |
| **API Format / Provider** | `OpenAI` या `OpenAI Compatible` | मानक REST प्रोटोकॉल |
| **Base URL / API Host** | `http://127.0.0.1:8081/v1` | लोकल हाई-स्पीड गेटवे |
| **API Key** | `none` (या आपकी सेट की गई Key) | ऑथेंटिकेशन बंद होने पर वैकल्पिक |
| **Model** | `gemini-3.7-flash`, `gemini-3.7-flash-thinking`, `gemini-pro` | हाई-स्पीड या डीप रीज़निंग |

### यूनिवर्सल एनवायरनमेंट वेरिएबल्स

CLI टूल्स, बैकग्राउंड बॉट्स, Python/Node स्क्रिप्ट्स, और एजेंट फ़्रेमवर्क्स के लिए:

```bash
export OPENAI_BASE_URL=http://127.0.0.1:8081/v1
export OPENAI_API_BASE=http://127.0.0.1:8081/v1
export OPENAI_API_KEY=none
```

---

## मॉडल सूची और रीज़निंग स्तर

| मॉडल नाम | बैकएंड मोड | डिफ़ॉल्ट रीज़निंग | विवरण | आवश्यक खाता |
| :--- | :---: | :---: | :--- | :--- |
| `gemini-3.7-flash` | Mode 1 | `@think=4` | **नवीनतम फ्लैगशिप फास्ट मॉडल** (~12k अक्षर) | मुफ़्त जीमेल |
| `gemini-3.7-flash-thinking` | Mode 2 | `@think=0` | **नवीनतम फ्लैगशिप डीप थिंकिंग मॉडल** (~20k+ अक्षर) | मुफ़्त जीमेल |
| `gemini-3.6-flash` / `gemini-flash` | Mode 1 | `@think=4` | तेज़ ऑल-राउंडर मॉडल | मुफ़्त जीमेल |
| `gemini-3.5-flash-thinking` / `gemini-thinking` | Mode 2 | `@think=0` | **विस्तृत सोच (Deep Thinking)** (~20k+ अक्षर) | मुफ़्त जीमेल |
| `gemini-3.5-flash-thinking-lite` | Mode 5 | `@think=0` | अनुकूलनीय रीज़निंग (~15k अक्षर) | मुफ़्त जीमेल |
| `gemini-flash-lite` / `gemini-lite` | Mode 6 | `@think=4` | अल्ट्रा-फास्ट कम लेटेंसी मॉडल | मुफ़्त जीमेल |
| `gemini-auto` | Mode 4 | `@think=4` | ऑटोमैटिक मॉडल चयन | मुफ़्त जीमेल |
| `gemini-3.1-pro` / `gemini-pro` | Mode 3 | `@think=4` | फ्लैगशिप प्रो कोडिंग व गणित मॉडल | **Gemini Advanced कुकी** |
| `gemini-3.1-pro-enhanced` | Mode 3 | `@think=4` | प्रो एन्हांस्ड आउटपुट (प्रायोगिक) | **Gemini Advanced कुकी** |

---

## आर्किटेक्चर और डेटा प्रवाह

<p align="center">
  <img src="assets/bob-gemini-free-architecture.jpg" alt="BOB Gemini Free Architecture" width="100%" />
</p>

1. **डेवलपर क्लाइंट्स**: Cursor, Cherry Studio, ChatBox, OpenWebUI, या Python SDK से स्टैंडर्ड OpenAI/Gemini REST रिक्वेस्ट भेजें।
2. **गेटवे इंजन**: OpenAI मैसेज को गूगल BoQ RPC पेलोड में ट्रांसलेट करता है, मल्टीमॉडल विज़न अपलोड को ऑटोमैटिक कंप्रेस करता है, और थिंकिंग टोकन्स को `reasoning_content` में एक्सट्रेक्ट करता है।
3. **गूगल वेब क्लाउड**: ब्राउज़र फिंगरप्रिंट और `SAPISIDHASH` प्रमाणीकरण के साथ सीधे गूगल जेमिनी बैकएंड से जुड़ता है।

---

## Pro मॉडल व Imagen इमेज जेनरेशन के लिए कुकी सेटअप (Gemini Advanced)

मुफ़्त जीमेल अकाउंट्स में Flash 3.7, Flash 3.6, Flash Thinking, और Flash Lite बिना किसी कुकी या कॉन्फ़िगरेशन के तुरंत काम करते हैं।

यदि आपके पास **Google AI / Gemini Advanced ($20/माह)** की सदस्यता है (या Jio 18 महीने / कॉलेज ऑफ़र है) या आप **Imagen 3 इमेज जेनरेशन** को सक्रिय करना चाहते हैं, तो प्रामाणिक **Pro** मॉडल (`gemini-3.1-pro` / `gemini-pro`) के लिए अपना सेशन कुकी जोड़ें:

---

### चरण 1: 15 सेकंड में अपनी कुकी प्राप्त करें (3 आसान तरीके)

**Google Chrome**, **Arc**, **Edge**, या **Brave** में [**gemini.google.com**](https://gemini.google.com) खोलें (साइन इन रहें)।

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Chrome DevTools (F12 या Cmd+Opt+I)                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│  [Network] Tab  Filter: [ app                   ] [X] Preserve log          │
│  ─────────────────────────────────────────────────────────────────────────  │
│  Name                                  Status   Type      Size              │
│  📄 app?eom=1&awwd=1&em=2&...          200      document  22.2 kB  <── क्लिक│
│  ⚙️ batchexecute?rpcids=...            200      xhr       14.5 kB           │
│  ─────────────────────────────────────────────────────────────────────────  │
│  [Headers] [Payload] [Preview] [Response]                                   │
│  ▼ Request Headers                                                          │
│    :authority: gemini.google.com                                            │
│    Cookie: __Secure-BUCKET=...; SID=...; SAPISID=...;  <── राइट क्लिक व कॉपी│
└─────────────────────────────────────────────────────────────────────────────┘
```

#### तरीका A: 1-क्लिक तुरंत विधि (बिना कोई चैट भेजे)
1. **`F12`** (या macOS पर **`Cmd + Option + I`**) दबाकर Developer Tools खोलें।
2. ऊपर **Network** टैब पर क्लिक करें।
3. पेज को रीफ़्रेश करें (**`Cmd + R`** या **`F5`**)।
4. डॉक्यूमेंट रिक्वेस्ट **`app?eom=1...`** (या सबसे ऊपर वाली **`batchexecute`** रिक्वेस्ट) पर क्लिक करें।
5. दाएँ पैनल में **Headers** टैब के अंदर **Request Headers** तक स्क्रॉल करें।
6. **`cookie:`** (या **`Cookie:`**) लाइन पर राइट-क्लिक करके **Copy value** चुनें।

#### तरीका B: 1-शब्द चैट विधि (`StreamGenerate`)
1. Network टैब में फ़िल्टर बॉक्स में **`StreamGenerate`** लिखें।
2. जेमिनी में कोई भी 1 शब्द टाइप करके भेजें (जैसे *"hi"* )।
3. लिस्ट में तुरंत **`StreamGenerate`** आ जाएगा $\rightarrow$ उस पर क्लिक करें $\rightarrow$ **Request Headers** में से **`cookie:`** कॉपी करें।

#### तरीका C: Application स्टोरेज टैब
1. DevTools में **Application** टैब पर जाएँ $\rightarrow$ **Storage** $\rightarrow$ **Cookies** $\rightarrow$ `https://gemini.google.com` चुनें।
2. कुकीज़ को सेलेक्ट करके कॉपी करें।

---

### चरण 2: BOB Gemini Free में कुकी सेट करें (जीरो कॉपी-पेस्ट व मैनुअल विधियाँ)

#### 🌟 विधि 0: 1-क्लिक इंटरैक्टिव लॉगिन विंडो (जीरो कॉपी-पेस्ट — सबसे आसान!)

टर्मिनल में लॉगिन कमांड चलाएँ:

```bash
./bob-gemini-free --login
```

```
┌─────────────────────────────────────────────────────────────┐
│  🌐 Sign in to Google Gemini (BOB लॉगिन विंडो)             │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│                    Google                                   │
│                    Sign in with Google                      │
│                    [ your-email@gmail.com ]                 │
│                    [ पासवर्ड / Passkey दर्ज करें ]          │
│                                                             │
│                    [ Next ]                                 │
│                                                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼ (केवल 1 बार सामान्य लॉगिन करें)
┌─────────────────────────────────────────────────────────────┐
│  [✔] 19+ सेशन टोकन्स सफलतापूर्वक प्राप्त हुए!              │
│  [✔] ./cookie.txt और ~/.config/bob-gemini-free/ में सुरक्षित│
│  [✔] Gemini Pro मॉडल (gemini-3.1-pro) व Imagen 3 सक्रिय!    │
└─────────────────────────────────────────────────────────────┘
```

1. आपकी स्क्रीन पर गूगल जेमिनी की एक स्टैंडअलोन लॉगिन विंडो खुलेगी।
2. अपने गूगल खाते से सामान्य रूप से साइन इन करें (Passkey, 2FA, SMS समर्थित)।
3. जैसे ही लॉगिन पूरा होगा, BOB Gemini Free **Chrome DevTools Protocol के माध्यम से ऑटोमैटिक रूप से सभी 19+ सेशन टोकन इंटरसेप्ट कर लेगा**, `cookie.txt` सुरक्षित रूप से सेव करेगा (`mode 0600`), और विंडो बंद कर देगा।
4. **बिना DevTools, बिना कॉपी-पेस्ट, बिना किसी कीचेन पासवर्ड पॉपअप के!**

---

#### 📸 मल्टीमॉडल इमेज विश्लेषण (Vision) और सेशन अनिवार्यता की सम्पूर्ण सच्चाई

गूगल का आंतरिक वेब ढाँचा सामान्य टेक्स्ट और मल्टीमॉडल मीडिया अपलोड्स के बीच स्पष्ट भेद करता है:

| फ़ीचर | गेस्ट / बिना लॉगिन सेशन | प्रामाणिक गूगल सेशन (`--login` द्वारा `./cookie.txt`) |
| :--- | :--- | :--- |
| **टेक्स्ट चैट व कोडिंग** (`gemini-3.7-flash`, `gemini-3.6-flash`) | ✅ पूरी तरह सक्रिय | ✅ 1M+ टोकन संदर्भ व तीव्र गति |
| **डीप रीज़निंग** (`gemini-3.7-flash-thinking`) | ✅ पूरी तरह सक्रिय | ✅ पूर्ण थिंकिंग ब्लॉक्स |
| **इमेज/स्क्रीनशॉट विश्लेषण (Vision)** | ❌ **गूगल द्वारा अवरुद्ध** (`BardErrorInfo [1003]`) | ✅ **पूरी तरह सक्रिय** (मुफ़्त) |
| **Imagen 3 इमेज जेनरेशन** (`imagen-3`) | ❌ **गूगल द्वारा अवरुद्ध** | ✅ **पूरी तरह सक्रिय** |
| **Pro मॉडल रूटिंग** (`gemini-3.1-pro` / `gemini-pro`) | ❌ Flash पर डायवर्ट | ✅ **पूरी तरह सक्रिय** |

**गूगल इमेजेस के लिए सेशन क्यों अनिवार्य करता है?**
जब आप इमेज या स्क्रीनशॉट अटैच करते हैं, तो BOB Gemini Free उसे गूगल के Scotty स्टोरेज (`content-push.googleapis.com`) पर अपलोड करता है। गूगल का सर्वर यह प्रमाणित करता है कि स्टोरेज खाता किसी वैध गूगल सेशन (`SAPISIDHASH` व `__Secure-1PSIDTS`) से क्रिप्टोग्राफ़िक रूप से जुड़ा है। यदि कुकी अनुपस्थित या पुरानी हो, तो गूगल `BardErrorInfo [1003]` रिटर्न करता है।

केवल एक बार `./bob-gemini-free --login` चलाने से यह क्षमता हमेशा के लिए अनलॉक हो जाती है।

---

#### विधि A: ऑटोमैटेड कुकी सेटअप हेल्पर (पेस्ट प्रॉम्प्ट)

टर्मिनल में स्वचालित हेल्पर कमांड चलाएँ:

```bash
./bob-gemini-free --setup-cookie
```

कॉपी की गई कुकी स्ट्रिंग पेस्ट करें और **Enter** दबाएँ। यह टूल:
* सभी 19+ सेशन टोकन्स (`SID`, `HSID`, `SSID`, `APISID`, `SAPISID`, `__Secure-1PSID`, `__Secure-3PSID`, `SIDCC`) को सत्यापित करता है।
* डायनामिक क्रिप्टोग्राफिक SHA-1 **`SAPISIDHASH`** को सक्रिय करता है।
* सुरक्षित रूप से `~/.config/bob-gemini-free/cookie.txt` में सेव करता है (`chmod 600`)।
* Pro मॉडल (`gemini-3.1-pro`) और Imagen 3 इमेज जेनरेशन को तुरंत सक्रिय कर देता है।

#### विधि B: जीरो-कॉन्फ़िग लोकल `cookie.txt`

सीधे गेटवे फ़ोल्डर में `cookie.txt` बनाकर अपनी कुकी पेस्ट कर दें:

```bash
cat << 'EOF' > cookie.txt
PASTE_YOUR_COOKIE_STRING_HERE
EOF
chmod 600 cookie.txt
./bob-gemini-free
```
*(BOB Gemini Free बिना किसी फ़्लैग के `./cookie.txt` को स्टार्टअप पर ऑटोमैटिक लोड कर लेता है)।*

#### विधि C: सीधे CLI फ़्लैग या एनवायरनमेंट वेरिएबल

```bash
# सीधे फ़्लैग द्वारा
./bob-gemini-free --cookie-string "SID=...; HSID=...; SAPISID=...; __Secure-1PSID=..."

# या एनवायरनमेंट वेरिएबल द्वारा
export BOB_GEMINI_FREE_COOKIE_FILE="/path/to/cookie.txt"
./bob-gemini-free
```

---

### चरण 3: मल्टी-अकाउंट प्रोफ़ाइल हैंडलिंग (`auth_user`)

यदि आप एक ही ब्राउज़र में कई जीमेल खातों में लॉग इन हैं (उदा. खाता 0 = व्यक्तिगत, खाता 1 = व्यावसायिक):

1. अपने जेमिनी ब्राउज़र URL की जाँच करें:
   * `https://gemini.google.com/app` $\rightarrow$ खाता इंडेक्स `0` है (डिफ़ॉल्ट)।
   * `https://gemini.google.com/u/1/app` $\rightarrow$ खाता इंडेक्स `1` है।
2. `config.json` में खाता इंडेक्स सेट करें:
   ```json
   {
     "auth_user": "1",
     "cookie_file": "./cookie.txt"
   }
   ```
   यह सुनिश्चित करता है कि BOB Gemini Free सही गूगल प्रोफ़ाइल पर रिक्वेस्ट भेजे (`X-Goog-AuthUser: 1`)।

---

### चरण 4: मल्टी-अकाउंट कुकी पूल व ऑटो-रोटेशन (`cookie_pool`)

उच्च समवर्ती (High-Concurrency) टीमों या AI एजेंट पाइपलाइनों के लिए:

1. `cookies/` फ़ोल्डर बनाएँ और कई कुकी फ़ाइलें रखें:
   ```
   ./cookies/
   ├── account_primary.txt
   ├── account_secondary.txt
   └── account_team.txt
   ```
2. या `config.json` में `"cookie_pool"` कॉन्फ़िगर करें:
   ```json
   {
     "cookie_pool": [
       "./cookies/account_primary.txt",
       "./cookies/account_secondary.txt"
     ]
   }
   ```
3. **स्वचालित लोड बैलेंसिंग व फ़ेलओवर**: BOB Gemini Free स्वचालित रूप से सभी खातों में अनुरोध वितरित करता है। यदि कोई खाता अस्थायी दर सीमा (HTTP 429) पर पहुँचता है, तो BOB **उसे 60 सेकंड के लिए बैकऑफ़ करता है और अगले स्वस्थ खाते पर पारदर्शी रूप से पुनः प्रयास करता है**।

---

### चरण 5: Docker और OrbStack के साथ चलाना

Docker या **OrbStack** में कुकी के साथ कंटेनर चलाने के लिए:

```bash
# 1. डॉकर इमेज बनाएँ
docker build -t bob-gemini-free:local .

# 2. कुकी माउंट करके कंटेनर चलाएँ
docker run -d \
  --name bob-gemini-free \
  -p 8081:8081 \
  -v $(pwd)/cookie.txt:/app/cookie.txt:ro \
  -e BOB_GEMINI_FREE_COOKIE_FILE=/app/cookie.txt \
  bob-gemini-free:local
```

कंटेनर स्टेटस देखें:
```bash
docker logs bob-gemini-free
# Output:
#   Cookie: yes (/app/cookie.txt)
#   Listening: http://0.0.0.0:8081
```

---

## अक्सर पूछे जाने वाले प्रश्न (FAQ)

<details>
<summary><strong>1. BOB Gemini Free बिना किसी API Key के मुफ़्त कैसे काम करता है?</strong></summary>

गूगल हर जीमेल अकाउंट धारक को अपने वेब इंटरफ़ेस के ज़रिए जेमिनी (Flash 3.7, Flash 3.6, Flash Lite, Flash Thinking) का मुफ़्त एक्सेस देता है। **BOB Gemini Free** एक हाई-परफ़ॉर्मेंस लोकल प्रॉक्सी है जो स्टैंडर्ड OpenAI और Gemini API कॉल्स को गूगल के आंतरिक वेब RPC फ़ॉर्मेट में ट्रांसलेट करता है।
</details>

<details>
<summary><strong>2. क्या मेरा गूगल अकाउंट सुरक्षित है? रेट लिमिट्स या बैन का कोई ख़तरा?</strong></summary>

* **मुफ़्त / बेनामी (Anonymous) मोड**: BOB Gemini Free **शून्य सेशन कुकीज़ और शून्य अकाउंट टोकन** भेजता है। सभी अनुरोध आपकी गूगल पहचान से पूरी तरह से अलग रहते हैं।
* **Gemini Advanced ($20/माह) मोड**: BOB Gemini Free वास्तविक ब्राउज़र TLS फ़िंगरप्रिंट (`tls-client` Chrome 133) और सामान्य वेब ट्रैफ़िक पैटर्न का उपयोग करता है।
* **ऑपरेशनल दिशानिर्देश**: एक समय में 3 से 5 समवर्ती अनुरोध चलाएँ। एक ही IP से प्रति सेकंड सैकड़ों अनुरोध भेजने से बचें।
</details>

<details>
<summary><strong>3. यह Google AI Studio Free Tier या Paid OpenAI APIs की तुलना में कैसा है?</strong></summary>

* **Google AI Studio Free Tier**: 15 RPM और सख्त दैनिक कोटा सीमा लगाता है (दैनिक सीमा समाप्त होने पर मध्यरात्रि UTC तक सेवा बंद हो जाती है)।
* **BOB Gemini Free**: गूगल के इंटरैक्टिव वेब बैकएंड पर काम करता है जिसमें **व्यावहारिक रूप से असीमित दैनिक प्रश्न**, शून्य टोकन लागत, और **20,000+ अक्षरों की डीप रीज़निंग** मुफ़्त मिलती है।
</details>

<details>
<summary><strong>4. मैं Google Pro मॉडल (`gemini-3.1-pro` / `gemini-pro`) और Imagen 3 को कैसे अनलॉक करूँ?</strong></summary>

मुफ़्त जीमेल पर Flash और Thinking पहले से उपलब्ध हैं। यदि आपके पास Gemini Advanced ($20/माह) या Reliance Jio 18 महीने / कॉलेज ऑफ़र है:
1. `./bob-gemini-free --login` चलाएँ (सुझाया गया 1-क्लिक स्टैंडअलोन लॉगिन)।
2. या `./bob-gemini-free --setup-cookie` चलाकर अपनी सेशन कुकी पेस्ट करें।
3. टूल ऑटोमैटिक रूप से `SAPISIDHASH` जनरेट कर Pro मॉडल और Imagen 3 को सक्रिय कर देगा।
</details>

<details>
<summary><strong>5. थिंकिंग / रीज़निंग मोड कैसे काम करता है? थिंकिंग टोकन कहाँ दिखते हैं?</strong></summary>

जब आप थिंकिंग मॉडल (`gemini-3.7-flash-thinking` या `@think=0`) पर क्वेरी करते हैं, तो BOB Gemini Free ऑटोमैटिक रूप से इंटरनल रीज़निंग ट्रेस को अलग कर OpenAI के `reasoning_content` फ़ील्ड में भेजता है। Cursor, Cherry Studio, ChatBox या OpenWebUI में यह अंतिम उत्तर के साथ एक सुंदर कोलैप्सेबल "Reasoning Process" कार्ड के रूप में दिखाई देता है।
</details>

<details>
<summary><strong>6. क्या मैं इसे Linux VPS, Raspberry Pi या Docker पर चला सकता हूँ?</strong></summary>

हाँ! BOB Gemini Free एक हल्की स्टैटिक बाइनरी (<15MB RAM) और आधिकारिक Multi-Arch Docker इमेज (`alpine:3.21`) के साथ आता है।
* VPS पर पब्लिक होस्टिंग के लिए `BOB_GEMINI_FREE_HOST=0.0.0.0` और `BOB_GEMINI_FREE_API_KEYS` सेट करें।
* यदि क्लाउड डेटासेंटर IP पर गूगल WAF चुनौती आए, तो `--proxy socks5://...` या TLS ब्राउज़र इंपर्सनेशन (`--impersonate chrome_133`) का उपयोग करें।
</details>

<details>
<summary><strong>7. क्या यह Tool / Function Calling और Structured JSON आउटपुट को सपोर्ट करता है?</strong></summary>

हाँ। यह ऑटोमैटिक रूप से टूल स्कीमा को सिस्टम प्रॉम्ट में इंजेक्ट करता है और ` ```tool_call ` आउटपुट को स्टैंडर्ड OpenAI `tool_calls` JSON में बदलता है। `response_format: {"type": "json_object"}` सख्त JSON आउटपुट सुनिश्चित करता है।
</details>

<details>
<summary><strong>8. मल्टीमॉडल विज़न और इमेज अपलोड कैसे काम करता है? इमेजेस के लिए सेशन कुकी क्यों आवश्यक है?</strong></summary>

OpenAI फ़ॉर्मेट में Base64 इमेज डेटा भेजें। BOB Gemini Free बड़ी छवियों को कंप्रेस करके गूगल के Scotty Resumable Upload प्रोटोकॉल (`content-push.googleapis.com`) पर अपलोड करता है।

* **सेशन अनिवार्यता**: गूगल Scotty स्टोरेज अपलोड्स को प्रामाणिक गूगल खाते (`SAPISIDHASH` + `__Secure-1PSIDTS`) से जोड़ता है। बिना लॉगिन वाले अनुरोधों पर गूगल `BardErrorInfo [1003]` रिटर्न करता है।
* **समाधान**: केवल एक बार `./bob-gemini-free --login` चलाकर लॉगिन करें; विज़न हमेशा के लिए सक्रिय हो जाएगा।
</details>

<details>
<summary><strong>9. क्या कई ऐप्स या टीम के सदस्य एक ही गेटवे शेयर कर सकते हैं?</strong></summary>

हाँ। `config.json` में `api_keys: ["sk-team-1", "sk-team-2"]` सेट करें। सभी आने वाले अनुरोधों का प्रमाणीकरण टाइमिंग-हमलों से सुरक्षित कॉन्स्टेंट-टाइम कम्पैरिजन (`crypto/subtle`) द्वारा होता है।
</details>

<details>
<summary><strong>10. क्या कोई टेलीमेट्री, ट्रैकिंग या बाहरी डेटा भेजा जाता है?</strong></summary>

शून्य। BOB Gemini Free 100% ओपन-सोर्स (MIT लाइसेंस) है, इसमें कोई ट्रैकिंग नहीं है, और नेटवर्क अनुरोध केवल सीधे आपके कंप्यूटर और गूगल सर्वर के बीच TLS पर होते हैं।
</details>

<details>
<summary><strong>11. Python या Node.js के बजाय Go भाषा क्यों?</strong></summary>

* **सिंगल स्टैटिक बाइनरी**: कोई virtualenv या `node_modules` की झंझट नहीं।
* **तुरंत स्टार्टअप**: <5ms में शुरू और <15MB RAM की न्यूनतम खपत।
* **अभूतपूर्व कंकरेंसी**: एक साथ कई SSE स्ट्रीमिंग डेल्टा कनेक्शन बिना किसी रुकावट के।
</details>

<details>
<summary><strong>12. क्या मैं बिना LiteLLM या प्रॉक्सी के Claude Code CLI को सीधे BOB Gemini Free से जोड़ सकता हूँ?</strong></summary>

**हाँ, 100% सीधे।** BOB Gemini Free में आधिकारिक **Anthropic Messages API प्रोटोकॉल (`POST /v1/messages`)** का पूर्ण इनबिल्ट सपोर्ट है, जिसमें सभी SSE इवेंट्स (`message_start`, `content_block_start`, `content_block_delta`, `message_delta`, `message_stop`) और bash टूल कॉलिंग शामिल हैं।

बस अपने टर्मिनल में ये वैरिएबल्स सेट करें और Claude Code शुरू करें:
```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:8081
export ANTHROPIC_API_KEY=none
claude
```
</details>

<details>
<summary><strong>13. क्या यह OpenAI Codex CLI (`openai/codex`) और AI Routers (LiteLLM / OpenRouter / Portkey) के साथ काम करता है?</strong></summary>

* **OpenAI Codex CLI**: नेटिव `/v1/responses` और `/v1/chat/completions` एंडपॉइंट्स द्वारा पूर्ण समर्थित। बस `OPENAI_BASE_URL=http://127.0.0.1:8081/v1` सेट करें।
* **LiteLLM / OpenRouter / Portkey / OneAPI**: `http://127.0.0.1:8081/v1` को कस्टम OpenAI अपस्ट्रीम के रूप में जोड़ें। गेटवे बिना किसी समस्या के सभी SSE डेल्टा चंक्स, रीज़निंग टोकन्स और यूसेज मीट्रिक्स रिटर्न करता है।
</details>

---

## ABCsteps के बारे में

[**ABCsteps**](https://abcsteps.com/) जोधपुर, राजस्थान, भारत में स्थापित एक ऑनलाइन एआई इंजीनियरिंग स्कूल है, जिसकी स्थापना **दिव्यांशु सिंह चौहान** द्वारा की गई है (ABC Steps Technologies Pvt Ltd द्वारा संचालित)।

इसका उद्देश्य डेवलपर्स, छात्रों और एआई इंजीनियरों के लिए व्यावहारिक और खुली शिक्षा प्रदान करना है:
* 📚 [**20-पाठों का AI इंजीनियरिंग पाठ्यक्रम**](https://abcsteps.com/offerings/) — प्रोजेक्ट-आधारित लर्निंग जिसमें AI कोपायलट, Docker, APIs, डेटाबेस और फुल-स्टैक AI सिस्टम्स शामिल हैं।
* ✍️ [**प्रैक्टिकल इंजीनियरिंग ब्लॉग**](https://abcsteps.com/blog/) — LLM इंटरनल्स, कोडिंग एजेंट्स और सिस्टम्स डिज़ाइन पर गहन तकनीकी लेख।
* 🧭 [**रीडिंग पाथ्स और शब्दावली**](https://abcsteps.com/blog/paths/) — स्व-अध्ययन करने वाले छात्रों के लिए संरचित मार्गदर्शिकाएँ।
* 🎓 [**फाउंडर-लेड मेंटरशिप व वर्कशॉप्स**](https://abcsteps.com/enroll/compare/) — लाइव कोहोर्ट्स, 1:1 मेंटरशिप और संस्थागत वर्कशॉप्स।

अधिक जानकारी के लिए देखें: [**https://abcsteps.com/**](https://abcsteps.com/)।

## आभार व शोध आधार (Acknowledgements & Foundations)

BOB Gemini Free वैश्विक एआई और ओपन-सोर्स समुदाय के सामूहिक ज्ञान और तकनीकी सफलताओं की नींव पर खड़ा है:

1. **Google Research और DeepMind**: ट्रांसफॉर्मर आर्किटेक्चर (*"Attention Is All You Need"*, Vaswani et al., 2017) के मूलभूत शोध तथा विश्वस्तरीय Gemini 3.7 Flash, Flash Thinking, 3.1 Pro और Imagen 3 मॉडल्स की मुफ़्त सार्वजनिक वेब उपलब्धता के लिए।
2. **OpenAI और Anthropic**: ओपन एपीआई स्टैंडर्ड्स, Messages स्कीमा, रीज़निंग ब्लॉक्स और कोडिंग एजेंट CLI पैटर्न स्थापित करने के लिए जिसने पूरी दुनिया के डेवलपर टूल्स को एक सूत्र में पिरोया।
3. **The Go Language Team और Chromium Engineers**: सिस्टम-स्तरीय इंजीनियरिंग (Go मानक लाइब्रेरी, ज़ीरो-डिपेंडेंसी स्टैटिक बाइनरी, और Chrome DevTools Protocol) के लिए जिसने इसे हाई-परफ़ॉर्मेंस, सुरक्षित और लोकल-फर्स्ट बनाया।
4. **वैश्विक ओपन-सोर्स समुदाय**: Cursor, Windsurf, Aider, Continue.dev, OpenWebUI, Cherry Studio, ChatBox के रचनाकारों और स्वतंत्र हैकर्स के लिए जो सॉफ़्टवेयर इंजीनियरिंग की सीमाओं को निरंतर आगे बढ़ा रहे हैं।
5. **ABCsteps Technologies (जोधपुर, राजस्थान)**: निष्काम कर्म योग, सत्यवादी इंजीनियरिंग शिक्षा और **Break Ordinary Boundaries (BOB)** के अंतर्गत हर डेवलपर को सशक्त बनाने के संकल्प के लिए।

---

## लाइसेंस (License)

MIT License. Developed with pride by [ABCsteps.com](https://abcsteps.com/) & [दिव्यांशु सिंह चौहान](https://github.com/div197).

MIT लाइसेंस। [ABCsteps.com](https://abcsteps.com/) द्वारा विकसित।
