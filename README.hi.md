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
  <img src="https://img.shields.io/badge/Release-v0.2.0--preview.9-7c3aed?style=flat-square" alt="Release">
  <img src="https://img.shields.io/badge/Go-1.26.6-00ADD8?style=flat-square&logo=go" alt="Go Version">
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

## वर्तमान स्थिति और प्रमाण सीमा

वर्तमान सार्वजनिक macOS previews में immutable `v0.2.0-preview.1` migration
bridge, historical Preview 3–8 और current controlled preview
`v0.2.0-preview.9` उपलब्ध हैं। Public `main` में Preview 4 के बाद का
artifact-preview, responsive-header, multiline-SSE, credential-boundary,
route-control, diagnostics, native external-link और deterministic
artifact-focus सुधार आ चुके हैं। Preview 9 के public bytes और signature को
फिर से download करके verify किया गया है और पाँचों assets local signed inputs से
byte-for-byte मिलते हैं। Public Preview 7 app के Help → Check for Updates path ने
live test में Preview 9 का consent dialog दिखाया; install action cancel किया गया,
इसलिए replacement, restart और rollback अभी device gates हैं।
Preview 1 → Preview 5 installed migration का प्रमाण उपलब्ध है; बाद के audit
Mac checks ने Preview 9 को खोज लिया था, लेकिन install action cancel किया गया,
इसलिए current replacement, rollback, clean-device, pilot और platform-trust
gates अभी खुले हैं। Stable `v0.2.0` अभी प्रकाशित नहीं है। Historical और local
version transition matrix [`RELEASE-TRANSITION-AUDIT-2026-09-01.md`](docs/engineering/RELEASE-TRANSITION-AUDIT-2026-09-01.md)
में दी गई है।
यह दस्तावेज़ सुविधा-सूची है, production या
unlimited-access प्रमाणपत्र नहीं। मॉडल पहचान, quota, context, vision, Imagen,
Google session और live compatibility provider पर निर्भर हैं; tool calling और
token counts के कुछ हिस्से emulated/estimated हैं। अंग्रेज़ी
[`VERIFICATION-MATRIX.md`](docs/engineering/VERIFICATION-MATRIX.md) में हर
दावे का source/test/device boundary दिया गया है। वर्तमान release और installed-
base gate [`RELEASE-AUDIT-2026-08-31.md`](docs/engineering/RELEASE-AUDIT-2026-08-31.md)
में दर्ज है।

---

## 🌟 दर्शन: "Break Ordinary Boundaries" क्यों?

आज के दौर में एआई सीखने वाले छात्रों, स्वतंत्र रचनाकारों और डेवलपर्स के सामने **तीन साधारण सीमाएँ (Ordinary Boundaries)** आती हैं:

1. 💸 **आर्थिक सीमा (महँगी API कीमतें)**: उन्नत एआई मॉडल्स के लिए महँगे क्रेडिट कार्ड्स और $20-$100 प्रति माह के बिल चुकाने पड़ते हैं।
2. 🔒 **इकोसिस्टम सीमा (प्लेटफ़ॉर्म बंधन)**: एंथ्रोपिक टूल्स केवल एंथ्रोपिक से बात करते हैं, ओपनएआई केवल अपने से, और जेमिनी का मुफ़्त वेब एक्सेस ब्राउज़र टैब तक सीमित रहता है।
3. ⚙️ **जटिलता की सीमा (इंस्टॉलेशन का झंझट)**: प्रॉक्सी सेटअप करने के लिए पायथन, गो, नोड या जटिल टूल्स इंस्टॉल करने पड़ते हैं।

**BOB इन तीनों सीमाओं को एक साथ तोड़ता है**:
- ✨ **Session-Dependent Access**: मॉडल, quota, reasoning, vision और image access वर्तमान Google web session तथा provider rules पर निर्भर हैं; कोई fixed free/unlimited guarantee नहीं है।
- 🔓 **"API-Less AI" आर्किटेक्चर**: gateway billing account की ज़रूरत नहीं, लेकिन Google session/provider access फिर भी लागू है। आपका local gateway configured credentials को boundary के भीतर रखता है।
- 🌉 **यूनिवर्सल 4-इन-1 प्रोटोकॉल**: एक ही सिंगल गेटवे से **OpenAI** (`/v1/chat/completions`, `/v1/responses`, `/v1/tokens/count`), **Anthropic Messages (Claude Code CLI)**, **Google Gemini v1beta** (`:countTokens`), और **Embedded Go लाइब्रेरी** (`pkg/gateway`) का सीधा अनुवाद।
- ⚡ **Packaged runtime simplicity**: packaged binary में अलग Go, Python, Node,
  SQLite या memory service की runtime ज़रूरत नहीं; source build के लिए
  documented toolchain चाहिए। **1-क्लिक लॉगिन (`--login`)** browser और Google
  पर निर्भर है।

---

## 🚀 "API-Less AI" क्रांति: डेवलपर्स के लिए सच्ची आज़ादी

पारंपरिक क्लाउड कंपनियाँ डेवलपर्स को क्रेडिट कार्ड और प्रति-टोकन बिलिंग के चक्रव्यूह में उलझा देती हैं। **BOB Gemini Free अपने gateway के लिए paid billing account अनिवार्य नहीं करता**; Google session और provider policy फिर भी लागू हैं:

| पारंपरिक क्लाउड API मॉडल | BOB का API-Less आर्किटेक्चर |
| :--- | :--- |
| 💳 क्रेडिट कार्ड और क्लाउड बिलिंग खाता अनिवार्य | **₹0.00 / शून्य क्रेडिट कार्ड की ज़रूरत** |
| 💸 हर मिलियन टोकन और रीज़निंग स्टेप का महँगा बिल | **Gateway billing नहीं; upstream quotas और session rules लागू** |
| 🔑 API Key लीक होने पर लाखों का नुकसान | **वैकल्पिक student-owned key केवल Studio memory में; default route को Developer API key नहीं चाहिए** |
| 🔒 किसी एक कंपनी के CLI या टूल में क़ैद | **यूनिवर्सल 4-इन-1 ट्रांसलेशन (OpenAI, Claude, Google, Go)** |
| 📊 महीने के अंत में चौंकाने वाला इनवॉइस | **लोकल अनुमानित टोकन व बचत काउंटर (`GET /`); billing record नहीं** |

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

जब **OpenAI Codex CLI**, **Claude Code CLI**, **Cursor**, या कोई अन्य समर्थित क्लाइंट जवाब माँगता है, तो BOB चुने हुए adapter route के अनुसार अनुरोध को जेमिनी वेब फ़ॉर्मेट में बदलने का प्रयास करता है। उत्तर, reasoning, vision और latency Google session तथा upstream protocol पर निर्भर हैं; कोई universal/full compatibility या मिलीसेकंड guarantee नहीं है।

---

## BOB सीरीज़ क्या है? (*Break Ordinary Boundaries*)

ABCsteps की **BOB सीरीज़** का उद्देश्य डेवलपर्स, छात्रों और एआई इंजीनियरों के लिए बिना किसी पेवॉल या महँगे सब्सक्रिप्शन के शक्तिशाली टूल्स और रनटाइम्स उपलब्ध कराना है:

* ⚡ [**BOB Gemini Free**](https://github.com/div197/bob-gemini-free) — OpenAI-shaped, Anthropic-shaped, Google और Go interfaces वाला multi-protocol लोकल गेटवे; compatibility endpoint-specific है।
* 🎥 [**BOB YouTube**](https://github.com/div197/BOB-Youtube) — डेवलपर्स और एआई एजेंट्स के लिए डॉकर-फर्स्ट यूट्यूब डेटा इंजेक्शन और प्रोसेसिंग टूल।

---

## मुख्य विशेषताएँ

* **Google वेब सेशन पर निर्भर रूटिंग**: मॉडल, कोटा, context, reasoning और image access वर्तमान Google session पर निर्भर हैं; कोई fixed free/unlimited guarantee नहीं है।
* **Multi-Protocol Adapter**: OpenAI, Anthropic, Google और Go interfaces implement और fixture-tested हैं; हर client feature के लिए native/drop-in guarantee नहीं है।
* **1-क्लिक नेटिव लॉगिन (`--login`)**: बिना DevTools या चाबी (Keychain) पॉपअप के एक क्लिक में गूगल अकाउंट जोड़ें।
* **मल्टी-अकाउंट कुकी पूल (`cookie_pool`)**: स्पष्ट रूप से दिए गए Google web
  sessions के बीच स्थानीय चयन और failure cooldown; यह quota बढ़ाने, 429 को
  bypass करने या साझा cookie बाँटने का तरीका नहीं है।
* **Authenticated Pro/Image path**: सही session और upstream support होने पर Pro/Image aliases उपलब्ध हो सकते हैं; gateway इसे guarantee नहीं करता।
* **Anthropic-shaped thinking**: Claude Code-shaped requests के लिए reasoning blocks मिलते हैं; यह native Claude inference नहीं है।
* **मल्टीमॉडल विज़न (Vision)**: OpenAI फॉर्मेट में Base64 इमेज या इमेज लिंक्स भेजें — ऑटोमैटिक कम्प्रेशन के साथ।
* **लोकल privacy boundary**: Go gateway automatic telemetry नहीं भेजता; binary memory और upstream performance target build पर मापनी होगी।

BOB कोई अलग account/signup नहीं बनाता और chat prompts को BOB cloud service पर
नहीं भेजता। जरूरत होने पर Google session उसी उपयोगकर्ता का होना चाहिए और
उसकी expiry, policy तथा entitlements लागू रहेंगी।
## समर्थित टूल्स और इकोसिस्टम (Supported Tools & Ecosystem)

BOB Gemini Free कई AI coding tools के लिए adapter endpoints देता है; हर client और feature को target endpoint पर verify करें:

| श्रेणी (Category) | समर्थित टूल्स व क्लाइंट्स | कनेक्शन एंडपॉइंट |
| :--- | :--- | :--- |
| **टर्मिनल कोडिंग एजेंट्स** | OpenAI Codex CLI (`codex`), Claude Code CLI (`claude`), Gemini CLI (`gemini`) | custom base URL के साथ adapter routes |
| **एजेंटिक IDEs व एक्सटेंशन्स** | Cursor (Agent Mode), Windsurf, VS Code (Continue, Roo Code, Cline) | `http://127.0.0.1:9610/v1` |
| **स्वायत्त एजेंट फ़्रेमवर्क्स** | Grok Build, LangChain, CrewAI, AutoGen, OpenAI Agents SDK | `http://127.0.0.1:9610/v1` |
| **रूटर्स व लोकल प्रॉक्सीज़** | LiteLLM, OneAPI, NewAPI, Portkey, OpenRouter | `http://127.0.0.1:9610/v1` |
| **ऑफिशियल SDKs** | OpenAI (Python/JS/Go/.NET/Java), Anthropic (Python/TypeScript), Google GenAI | लोकल बेस यूआरएल |

---

## त्वरित शुरुआत (Quick Start)

### सबसे आसान तरीका (Super Simple - बिना किसी कॉन्फ़िग के)

```bash
# 1. गेटवे शुरू करें
./bob-gemini-free

# 2. किसी भी टूल में सेट करें:
# Base URL: http://127.0.0.1:9610/v1
# API Key:  none
```

---

### ऑटोमैटिक 15-पॉइंट टेस्ट किट (Diagnostic Test Kit)

सभी 13 मॉडल्स, स्ट्रीमिंग, टोकन काउंटिंग, और API फ़ॉर्मेट्स की लाइव जाँच के लिए:

```bash
# डिफ़ॉल्ट लोकल गेटवे का परीक्षण
./bob-gemini-free --test

# या कस्टम पोर्ट / API Key के साथ
./bob-gemini-free --test --test-url http://127.0.0.1:9610 --test-key your_api_key

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

ध्यान दें: वर्तमान stable `v0.1.5` CLI release में signed `SHA256SUMS` और
`SHA256SUMS.sig` manifest नहीं है। इसलिए installer सुरक्षा के लिए binary
इंस्टॉल करने से पहले रुक जाएगा। वर्तमान native student preview के लिए केवल
official [public Preview 9 release page](https://github.com/div197/BOB-Gemini-Free/releases/tag/v0.2.0-preview.9)
से सूचीबद्ध macOS package लें।

```bash
curl --fail --location --proto '=https' --proto-redir '=https' \
  --output install.sh \
  https://raw.githubusercontent.com/div197/BOB-Gemini-Free/main/install.sh
less install.sh
chmod +x install.sh
./install.sh
```

स्क्रिप्ट को पहले स्थानीय फ़ाइल के रूप में डाउनलोड करके पढ़ें। डिफ़ॉल्ट
इंस्टॉलर signed manifest और SHA-256 की जाँच के बिना binary इंस्टॉल नहीं करता;
यह अनजाने current directory को source से compile भी नहीं करता।

---

### विकल्प C: नेटिव बैकग्राउंड सर्विस डेमन (`service`)

macOS (`launchd`), Linux (`systemd`), और Windows (Startup folder) पर सिस्टम रीबूट के बाद भी BOB को 24/7 बैकग्राउंड में स्वचालित रूप से चलाएँ:

```bash
# बूट पर ऑटो-स्टार्ट सक्षम करें (शून्य टर्मिनल विंडो की आवश्यकता)
./bob-gemini-free service install

# बैकग्राउंड डेमन स्थिति व पिंग जांचें
./bob-gemini-free service status

# डेमन प्रारंभ / बंद / अनइंस्टॉल करें
./bob-gemini-free service start
./bob-gemini-free service stop
./bob-gemini-free service uninstall
```

---

### विकल्प D: signed CLI updater (`--update`)

GitHub की signed release से CLI को एक explicit command में अपडेट करें:

```bash
./bob-gemini-free --update
```

Public `v0.2.0-preview.9` preview में project update key embedded है। यह
current same-key controlled preview है और explicit consent मांगता है; stable
`v0.2.0` अभी pilot testing के बाद ही प्रकाशित होगी। एक writable Preview 1
installation से Preview 5 तक का replacement इस audit Mac पर प्रमाणित है;
Preview 7 fleet, deliberate rollback और clean-device प्रमाण अलग gates हैं।

---

### विकल्प E: इंटरैक्टिव लोकल-फ़र्स्ट वेब स्टूडियो (`/playground` व `bob-gemini-free.abcsteps.com`)

अपने किसी भी वेब ब्राउज़र में इनबिल्ट विज़ुअल इंटरफ़ेस का उपयोग करें:
* 🌐 **ऑनलाइन Cloudflare Pages वेब स्टूडियो**: **[bob-gemini-free.abcsteps.com](https://bob-gemini-free.abcsteps.com/)** *(स्थिर UI; गेटवे, सेशन और upstream सीमाएँ लागू)*
* 🏠 **लोकल सर्वर एड्रेस**: `http://127.0.0.1:9610/playground` (या `/ui`)

#### 🌟 क्लाइंट क्षमताएँ (ब्राउज़र, गेटवे और upstream सीमाओं के अधीन):
* 🔒 **लोकल-फ़र्स्ट प्राइवेसी**: hosted origin को `allowed_origins` में स्पष्ट रूप से अनुमति देनी होती है; PNA authentication नहीं है। Go gateway prompts को किसी BOB मध्यवर्ती सर्वर पर नहीं भेजता, लेकिन browser/CDN dependencies अलग network calls कर सकती हैं।
* 🐍 **संस्थागत स्तर पर इन-ब्राउज़र Pyodide WASM Python सैंडबॉक्स**: CPython 3.11 को बिना किसी सर्वर-साइड जोखिम, बिना Python इंस्टालेशन और शून्य बिलिंग के सीधे ब्राउज़र में चलाएँ। इंटरैक्टिव `input()` और साइंटिफिक पैकेजेस (`numpy`, `pandas`, `matplotlib`, `scipy`, `sympy`) का ऑटो-लोडर।
* 🧭 **कोई server database या memory service नहीं**: gateway stateless request handling, explicit session pool और safe aggregate counters तक सीमित है; studio SQLite या server-side database provision नहीं करता।
* ⚡ **नेटिव क्लाइंट-साइड इंटरैक्टिव आर्टिफ़ैक्ट्स इंजन (Claude-Style Live Canvas)**: HTML5 वेब ऍप्लिकेशन्स, CSS3 एनिमेशन्स, Canvas 2D/WebGL गेम्स, SVG वेक्टर ग्राफिक्स और Mermaid डायग्राम्स को 1-क्लिक **`Launch ⚡`** चिप से सीधे सैंडबॉक्स्ड स्टूडियो मोडल (`[ ▶ Preview | ⟨/⟩ Code ]`) में चलाएँ।
* 🪄 **Prompt assistant (`🪄`)**: local gateway उपलब्ध होने पर prompt सुधारने का
  प्रयास करता है; unavailable होने पर स्पष्ट error देता है और मूल prompt को
  जस का तस रखता है। यह नकली provider response या latency guarantee नहीं है।
* 🔍 **नॉन-ब्रेकिंग रीडिंग टेक्स्ट ज़ूम कंट्रोलर (`🔍 100%`)**: सब-बार चिप, `⌘+`/`⌘-`/`⌘0` और कमांड पैलेट द्वारा पूरे लेआउट को बिना तोड़े पढ़ने के फॉन्ट साइज़ को बढ़ाएँ/घटाएँ।
* 🏛️ **यूनिफ़ाइड सेक्रेड ज्योमेट्री स्टूडियो इनपुट कैप्सूल**: मोबाइल व डेस्कटॉप दोनों पर टेक्स्टएरिया, विज़न अटैचमेंट्स, टूल्स (`📎`, `अ`, `🎙️`, `🪄`) और स्वर्णिम `SEND ➤` बटन को समाहित करने वाला एकल कैप्सूल।
* 🎙️ **नेचुरल HD स्पीच स्टूडियो व फ्लोटिंग ऑडियो बार**: NotebookLM-श्रेणी का न्यूरल वॉइस सिंथेसाइज़र (Play/Pause, `0.8x`–`1.5x` स्पीड, और 4-बार लाइव इक्वलाइज़र)।
* ✏️ **इन-प्लेस मैसेज एडिटिंग व रीवाइंड**: पुराने संदेशों को सीधे इनलाइन एडिट (`✏️`) करने और चैट को वहीं से फिर से शुरू करने की सुविधा।
* 🌐 **द्विभाषी व 8 क्षेत्रीय लिपियों में इंटरफ़ेस**: 1-क्लिक हेडर स्विचर व ⌘K शॉर्टकट (`L1` English, `L2` हिन्दी), साथ में 8 भारतीय लिपियों का समर्थन (हिन्दी, संस्कृतम्, मराठी, বাংলা, ગુજરાતી, தமிழ், తెలుగు, ਪੰਜਾਬੀ)।
* ✍️ **रीयल-टाइम फोनेटिक लिप्यंतरण (Backspace Undo के साथ)**: अंग्रेज़ी अक्षरों में टाइप करने पर (`"namaste"`) स्पेस दबाते ही शुद्ध देवनागरी (`"नमस्ते"`) में परिवर्तन। `Backspace` दबाते ही शब्द वापस अंग्रेज़ी अक्षरों में बदल जाता है।
* 🏛️ **स्थानीय classroom deployment option**: LAN hosting जानबूझकर सक्षम करने
  पर ही `--host 0.0.0.0` उपयोग करें; default gateway loopback-only है और
  classroom scale/provider quota को अलग से test करना होगा।
* 🧠 **लाइव रीज़निंग विज़ुअलाइज़र**: Gemini 3.7 Flash Thinking के रीज़निंग टोकन्स को रीयल-टाइम में प्रवाहित होते हुए देखें।
* ⚡ **मॉडल व थिंकिंग स्विचर**: तेज़ मॉडल्स, डीप रीज़निंग, या Imagen 3 इमेज सिंथेसिस को एक क्लिक में टेस्ट करें।
* 📊 **लोकल aggregate status**: uptime, requests, token estimates, latency और
  अनुमानित savings दिख सकते हैं; यह external analytics या provider billing
  record नहीं है।
* 📋 **मल्टी-प्रोटोकॉल स्निपेट जनरेटर**: Python, Claude Code CLI, और cURL कोड स्निपेट्स को तुरंत कॉपी करें।

#### वैकल्पिक: छात्र की अपनी Gemini Developer API key

सामान्य BOB web-session route के लिए Google AI Studio key की आवश्यकता नहीं है।
यदि कोई छात्र अपने project का उपयोग करना चाहता है, तो **Config → Google Gemini
Developer API** खोलें और [Google AI Studio key page](https://aistudio.google.com/app/apikey)
पर अपनी key बनाएँ या प्रबंधित करें। [वर्तमान model list](https://ai.google.dev/gemini-api/docs/models) और
[rate limits](https://ai.google.dev/gemini-api/docs/rate-limits) देखें, फिर Key को
BOB में केवल current session के लिए paste करें और **Use for this session** चालू करें।

यह quota bypass नहीं है। Key page memory में रहती है, `localStorage`, config
file, logs, metrics या release में save नहीं होती, और BOB keys की list rotate
नहीं करता। हर छात्र केवल अपने authorized project/key का उपयोग करे। Google के
current model, free-tier, rate-limit, billing और data-use नियम लागू रहते हैं।
पूरी route और limit सीमा [`GEMINI-API-ROUTING.md`](docs/engineering/GEMINI-API-ROUTING.md)
में है। Hosted public Studio में key भेजने से पहले local या explicitly trusted
gateway endpoint save करना अनिवार्य है; public demo में अज्ञात gateway को key
न दें। Direct route provider के `gemini-*` model IDs को forward करता है, इसलिए
भविष्य के model IDs को BOB alias समझकर silently बदलता नहीं है।

#### Credential map: कौन-सी key कहाँ है

इन values के owner और काम अलग-अलग हैं:

| Value | Owner | काम | BOB में स्थिति |
|---|---|---|---|
| **BOB Gateway Access Key** (`api_keys`) | Gateway operator | API-key protection चालू होने पर BOB endpoint की access सुरक्षा; यह Google credential नहीं है | Config में केवल current page session के लिए; BOB authorization के रूप में भेजी जाती है |
| **Gemini Developer API key** | Student / Google AI Studio project owner | स्पष्ट Google Developer API route चुनती है और उसी project के model, quota, billing और data-use नियम लागू करती है | अलग field में paste करें, route को स्पष्ट रूप से चालू करें; key केवल page memory में रहती है |
| **Web-session cookies** (`cookie_file` / cookie pool) | Local engine चलाने वाला व्यक्ति | BOB के default reverse-engineered Google web-session route को उपलब्ध होने पर authenticate करती हैं | Engine इन्हें disk पर manage करता है; किसी भी Config key field में cookie paste न करें |
| **Gateway Endpoint URL** | Studio से जुड़ने वाला व्यक्ति | तय करता है कि request किस BOB process को मिलेगी; remote endpoint अलग trust decision है | UI preference के रूप में save हो सकती है; local के लिए loopback या trusted remote के लिए HTTPS उपयोग करें |

Local unauthenticated install में BOB access key आवश्यक नहीं है। यदि gateway
में यह configured है, तो पहले BOB access check होगा, फिर request default web-session
या स्पष्ट रूप से चुने गए Developer API route पर जाएगी। Developer API key BOB
gateway को authenticate नहीं करती और BOB access key Google Developer API access
नहीं देती।

<p align="center">
  <img src="assets/bob-gemini-free-playground.png" alt="BOB Gemini Free वेब प्लेग्राउंड व टेलीमेट्री डैशबोर्ड — BOB Builder Theme Default" width="100%">
</p>

| थीम | डिज़ाइन सौंदर्यशास्त्र | शॉर्टकट | डायरेक्ट URL प्रिव्यू |
| :--- | :--- | :---: | :--- |
| 🏗️ **BOB Builder** *(डिफ़ॉल्ट)* | औद्योगिक हाई-कॉन्ट्रास्ट डार्क स्लेट व ऊर्जावान अंबर | `T1` | [स्क्रीनशॉट देखें](assets/theme-bob-builder.png) • `/playground?theme=bob-builder` |
| 🍏 **Apple Design** | SF Pro टाइपोग्राफी, फ्रॉस्टेड ग्लास, पार्चमेंट कार्ड्स व एक्शन ब्लू | `T5` | [स्क्रीनशॉट देखें](assets/theme-apple.png) • `/playground?theme=apple` |
| 📰 **Vodafone Editorial** | स्वच्छ लाइट एडिटोरियल paper, serif typography व crimson | `T2` | [स्क्रीनशॉट देखें](assets/theme-vodafone.png) • `/playground?theme=vodafone` |
| 🎧 **Spotify Dark** | शुद्ध AMOLED obsidian deep dark व emerald हरा | `T3` | [स्क्रीनशॉट देखें](assets/theme-spotify.png) • `/playground?theme=spotify` |
| ⚛️ **Gemini Quantum** | साइबर deep indigo canvas व luminescent cyan glow | `T4` | [स्क्रीनशॉट देखें](assets/theme-quantum.png) • `/playground?theme=quantum` |

---

### विकल्प D: सोर्स कोड से निर्माण (Build from Source - Go 1.26.6 snapshot)

```bash
make build
./bob-gemini-free --port 9610
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
    base_url="http://127.0.0.1:9610/v1",
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

client = OpenAI(base_url="http://127.0.0.1:9610/v1", api_key="none")

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

### Claude Code CLI और Anthropic SDK (Anthropic-shaped adapter)

BOB Gemini Free Anthropic Messages API (`POST /v1/messages`) का tested subset
implement करता है। यह Google web protocol पर आधारित adapter/emulation layer है,
native Claude inference नहीं; client और tool compatibility endpoint-specific है।

#### Claude Code CLI सेटअप

```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:9610
export ANTHROPIC_API_KEY=none
claude
```

#### Anthropic Python SDK

```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://127.0.0.1:9610",
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
export OPENAI_BASE_URL=http://127.0.0.1:9610/v1
export OPENAI_API_KEY=none
codex
```

### OpenAI इमेज जेनरेशन (`POST /v1/images/generations`)

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://127.0.0.1:9610/v1",
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

	http.ListenAndServe("127.0.0.1:9610", handler)
}
```

---

## विस्तृत आर्किटेक्चर तुलना: Google AI Studio बनाम BOB Gemini Free

### आधिकारिक सीमाएँ और फ़ीचर मैट्रिक्स (मुफ़्त टियर बिना पेड बिलिंग)

| पैमाना / विशेषता | Google AI Studio (Free Tier) | **BOB Gemini Free (लोकल गेटवे)** |
| :--- | :--- | :--- |
| **Requests और token quota** | Model, project, tier, region और provider policy के अनुसार बदलते हैं; current [AI Studio limits](https://ai.google.dev/gemini-api/docs/rate-limits) देखें | **स्थापित नहीं; Google web session पर निर्भर** |
| **Free-tier model availability** | Current [pricing page](https://ai.google.dev/gemini-api/docs/pricing) पर selected models और शर्तें बदल सकती हैं | **स्थापित नहीं; Google web session पर निर्भर** |
| **Paid billing** | Paid project/key से charges हो सकते हैं; [billing](https://ai.google.dev/gemini-api/docs/billing) जाँचें | **BOB gateway billing नहीं; provider/session rules लागू** |
| **रीज़निंग / थिंकिंग गहराई** | मुफ़्त कुंजियों पर सीमित | **Upstream-dependent; fixed character guarantee नहीं** |
| **OpenAI प्रोटोकॉल सपोर्ट** | ❌ शून्य (कस्टम SDK कोड चाहिए) | **✅ Adapter routes; fixture-tested, universal/full नहीं** |
| **Developer रोल सपोर्ट** | ❌ नहीं | **✅ Prompt transformation; native role semantics नहीं** |
| **रीज़निंग टोकन एक्सपोर्ट** | ❌ प्रोप्राइटरी फ़ॉर्मेट | **✅ स्टैंडर्ड `reasoning_content` (Cursor/OpenWebUI में कार्ड)** |
| **डेटा प्राइवेसी व ट्रेनिंग** | ⚠️ Provider policy लागू | **🛡️ लोकल gateway `127.0.0.1` पर; Go process automatic telemetry नहीं भेजता** |
| **सेटअप जटिलता** | क्लाउड कंसोल, प्रोजेक्ट निर्माण, API Key प्रबंधन | **जीरो कॉन्फ़िग: `./bob-gemini-free` चलाएँ और कोडिंग शुरू करें** |
| **कुल वित्तीय लागत** | $0 (थ्रॉटल होने तक) या महँगे पेड टोकन | **Gateway billing नहीं; Google account/session rules लागू** |

---

### अधिकतम ऑपरेशनल सीमाएँ और अनुशंसित दिशानिर्देश

नीचे के पुराने आँकड़े release guarantee नहीं हैं। Quota, context, concurrency और output depth को target environment और current Google session में मापें:

* **Output/context limits**: gateway द्वारा कोई fixed सीमा स्थापित नहीं है; Google web behavior upstream-dependent है।
* **Concurrency**: repository के local benchmark numbers Google limits या safe live classroom capacity नहीं बताते। Burst से बचें और छोटे rollout waves में test करें।
* **Retry strategy**: configurable attempts और fixed retry delay हैं; upstream errors/rate limits फिर भी आ सकते हैं।
* **Images**: compression path oversized input को validate करके dimensions सीमित करता है और requested byte budget तक JPEG quality/resolution घटाता है; budget असंभव होने पर over-budget image upload करने के बजाय request fail होती है।

---

## लाइव परफॉर्मेंस व स्ट्रेस बेंचमार्क

गेटवे की गति और थ्रूपुट मापने के लिए इनबिल्ट बेंचमार्क चलाएँ:

```bash
# 3 वर्कर्स और 6 प्रश्नों के साथ बेंचमार्क चलाएँ
./bob-gemini-free --bench --bench-concurrency 3 --bench-requests 6

# या स्क्रिप्ट चलाएँ
./scripts/bench.sh http://127.0.0.1:9610 3 6 your_api_key
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

**BOB Gemini Free** OpenAI-shaped adapter routes का tested subset देता है।
हर client, tool mode और endpoint के लिए compatibility अलग से verify करनी
होगी; इसे official 100% OpenAI compatibility न समझें।

### यूनिवर्सल इंटीग्रेशन पैटर्न

किसी भी AI एप्लिकेशन में यह तीन मानक सेटिंग्स भरें:

| सेटिंग फ़ील्ड | वैल्यू | विवरण |
| :--- | :--- | :--- |
| **API Format / Provider** | `OpenAI` या `OpenAI Compatible` | मानक REST प्रोटोकॉल |
| **Base URL / API Host** | `http://127.0.0.1:9610/v1` | लोकल हाई-स्पीड गेटवे |
| **API Key** | `none` (या आपकी सेट की गई Key) | ऑथेंटिकेशन बंद होने पर वैकल्पिक |
| **Model** | `gemini-3.7-flash`, `gemini-3.7-flash-thinking`, `gemini-pro` | हाई-स्पीड या डीप रीज़निंग |

### यूनिवर्सल एनवायरनमेंट वेरिएबल्स

CLI टूल्स, बैकग्राउंड बॉट्स, Python/Node स्क्रिप्ट्स, और एजेंट फ़्रेमवर्क्स के लिए:

```bash
export OPENAI_BASE_URL=http://127.0.0.1:9610/v1
export OPENAI_API_BASE=http://127.0.0.1:9610/v1
export OPENAI_API_KEY=none
```

---

## मॉडल सूची और रीज़निंग स्तर

| मॉडल नाम | बैकएंड मोड | डिफ़ॉल्ट रीज़निंग | विवरण | आवश्यक खाता |
| :--- | :---: | :---: | :--- | :--- |
| `gemini-3.7-flash` | Mode 1 | `@think=4` | fast-mode alias; व्यवहार upstream-dependent | session/provider-dependent |
| `gemini-3.7-flash-thinking` | Mode 2 | `@think=0` | thinking-mode alias; depth upstream-dependent | session/provider-dependent |
| `gemini-3.6-flash` / `gemini-flash` | Mode 1 | `@think=4` | Flash-mode alias | session/provider-dependent |
| `gemini-3.5-flash-thinking` / `gemini-thinking` | Mode 2 | `@think=0` | thinking-mode alias | session/provider-dependent |
| `gemini-3.5-flash-thinking-lite` | Mode 5 | `@think=0` | lite thinking-mode alias | session/provider-dependent |
| `gemini-flash-lite` / `gemini-lite` | Mode 6 | `@think=4` | lite-mode alias | session/provider-dependent |
| `gemini-auto` | Mode 4 | `@think=4` | auto-routing alias | session/provider-dependent |
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

Flash, thinking और lite aliases को local routing contracts समझें। वास्तविक
access, quota, model identity और session requirements target account पर live
conformance runbook से जाँचें।

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
| **टेक्स्ट चैट व कोडिंग** (`gemini-3.7-flash`, `gemini-3.6-flash`) | session/provider-dependent | session/provider-dependent |
| **डीप रीज़निंग** (`gemini-3.7-flash-thinking`) | session/provider-dependent | session/provider-dependent |
| **इमेज/स्क्रीनशॉट विश्लेषण (Vision)** | session/provider-dependent | authenticated session may permit it; live proof required |
| **Imagen 3 इमेज जेनरेशन** (`imagen-3`) | session/provider-dependent | session/provider-dependent |
| **Pro मॉडल रूटिंग** (`gemini-3.1-pro` / `gemini-pro`) | session/provider-dependent | session/provider-dependent |

**गूगल इमेजेस के लिए सेशन क्यों अनिवार्य करता है?**
जब आप इमेज या स्क्रीनशॉट अटैच करते हैं, तो BOB Gemini Free उसे गूगल के Scotty स्टोरेज (`content-push.googleapis.com`) पर अपलोड करता है। गूगल का सर्वर यह प्रमाणित करता है कि स्टोरेज खाता किसी वैध गूगल सेशन (`SAPISIDHASH` व `__Secure-1PSIDTS`) से क्रिप्टोग्राफ़िक रूप से जुड़ा है। यदि कुकी अनुपस्थित या पुरानी हो, तो गूगल `BardErrorInfo [1003]` रिटर्न करता है।

`./bob-gemini-free --login` वर्तमान उपयोगकर्ता के लिए session capture का प्रयास करता है। Cookies expire/revoke हो सकती हैं और provider capability की कोई permanent guarantee नहीं है।

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
  -p 9610:9610 \
  -v $(pwd)/cookie.txt:/app/cookie.txt:ro \
  -e BOB_GEMINI_FREE_COOKIE_FILE=/app/cookie.txt \
  bob-gemini-free:local
```

कंटेनर स्टेटस देखें:
```bash
docker logs bob-gemini-free
# Output:
#   Cookie: yes (/app/cookie.txt)
#   Listening: http://0.0.0.0:9610
```

---

## अक्सर पूछे जाने वाले प्रश्न (FAQ)

<details>
<summary><strong>1. BOB Gemini Free अपने gateway API key और billing account के बिना कैसे काम करता है?</strong></summary>

**BOB Gemini Free** selected OpenAI-shaped और Google-shaped requests को Google के internal web RPC में translate करता है। Default web route में model access, quota, authentication और cost current Google web session और provider policy पर निर्भर हैं। Optional Developer API route में छात्र के अपने project की key और quota लागू होती है।
</details>

<details>
<summary><strong>2. क्या मेरा गूगल अकाउंट सुरक्षित है? रेट लिमिट्स या बैन का कोई ख़तरा?</strong></summary>

* **बिना configured cookie**: BOB user-provided Google session file load नहीं करता; upstream access और identity behavior provider-controlled हैं, इसलिए इसे anonymous guarantee न समझें।
* **Google session safety**: Cookie उपयोगकर्ता की अपनी local session होनी चाहिए; expiry और provider policy लागू रहती है।
* **ऑपरेशनल दिशानिर्देश**: छोटे classroom waves रखें, burst traffic से बचें, और 401/403/429 को provider conditions मानें। Proxy rotation, fingerprint changes या shared cookies से limits bypass न करें।
</details>

<details>
<summary><strong>3. यह Google AI Studio Free Tier या Paid OpenAI APIs की तुलना में कैसा है?</strong></summary>

* **Google AI Studio Free Tier**: exact quota model और usage tier के अनुसार बदलता है। Current [pricing](https://ai.google.dev/gemini-api/docs/pricing) और [rate-limit](https://ai.google.dev/gemini-api/docs/rate-limits) pages तथा AI Studio project view देखें। पुराने fixed RPM/RPD numbers पर निर्भर न करें।
* **BOB Gemini Free web route**: Google interactive web backend का adapter है; daily limit, token cost, model access और reasoning depth current session/provider rules पर निर्भर हैं।
* **BOB Gemini Free Developer API route**: छात्र द्वारा चुने गए project/key की current quota और billing settings लागू होती हैं।
</details>

<details>
<summary><strong>4. मैं Google Pro मॉडल (`gemini-3.1-pro` / `gemini-pro`) और Imagen 3 को कैसे अनलॉक करूँ?</strong></summary>

Flash और Thinking aliases की वास्तविक उपलब्धता target Google session पर जाँचें। यदि आपके पास Gemini Advanced ($20/माह) या कोई अन्य provider entitlement है:
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

हाँ! BOB Gemini Free एक single static binary और multi-arch Docker path (`alpine:3.21`) के साथ आता है। RAM target build पर मापें; कोई fixed <15MB claim नहीं है।
* VPS पर पब्लिक होस्टिंग के लिए `BOB_GEMINI_FREE_HOST=0.0.0.0` और `BOB_GEMINI_FREE_API_KEYS` सेट करें।
* यदि क्लाउड डेटासेंटर IP पर गूगल WAF चुनौती आए, तो `--proxy socks5://...` या TLS ब्राउज़र इंपर्सनेशन (`--impersonate chrome_133`) का उपयोग करें।
</details>

<details>
<summary><strong>7. क्या यह Tool / Function Calling और Structured JSON आउटपुट को सपोर्ट करता है?</strong></summary>

आंशिक रूप से। यह tool schema को prompt में inject करता है और ` ```tool_call ` output को parse करता है; यह native Google function calling नहीं है। `response_format: {"type": "json_object"}` केवल instruction है, strict provider guarantee नहीं।
</details>

<details>
<summary><strong>8. मल्टीमॉडल विज़न और इमेज अपलोड कैसे काम करता है? इमेजेस के लिए सेशन कुकी क्यों आवश्यक है?</strong></summary>

OpenAI फ़ॉर्मेट में Base64 इमेज डेटा भेजें। BOB Gemini Free बड़ी छवियों को कंप्रेस करके गूगल के Scotty Resumable Upload प्रोटोकॉल (`content-push.googleapis.com`) पर अपलोड करता है।

* **सेशन अनिवार्यता**: गूगल Scotty स्टोरेज अपलोड्स को प्रामाणिक गूगल खाते (`SAPISIDHASH` + `__Secure-1PSIDTS`) से जोड़ता है। बिना लॉगिन वाले अनुरोधों पर गूगल `BardErrorInfo [1003]` रिटर्न करता है।
* **समाधान**: `./bob-gemini-free --login` से session capture का प्रयास करें; vision availability और validity Google session पर निर्भर हैं।
</details>

<details>
<summary><strong>9. क्या कई ऐप्स या टीम के सदस्य एक ही गेटवे शेयर कर सकते हैं?</strong></summary>

हाँ। `config.json` में `api_keys: ["sk-team-1", "sk-team-2"]` सेट करें। सभी आने वाले अनुरोधों का प्रमाणीकरण टाइमिंग-हमलों से सुरक्षित कॉन्स्टेंट-टाइम कम्पैरिजन (`crypto/subtle`) द्वारा होता है। Credential को `Authorization`, `x-api-key` या `x-goog-api-key` header में भेजें। URL में `?key=` प्रमाणीकरण default रूप से बंद है क्योंकि URL history, logs, referrer और proxy में leak हो सकता है। पुराने client के लिए केवल स्पष्ट opt-in के साथ `allow_query_api_key: true` या `BOB_GEMINI_FREE_ALLOW_QUERY_API_KEY=true` सेट करें और gateway को local रखें।
</details>

<details>
<summary><strong>10. क्या कोई टेलीमेट्री, ट्रैकिंग या बाहरी डेटा भेजा जाता है?</strong></summary>

Go gateway automatic telemetry नहीं भेजता। Browser studio CDN assets और input-tools endpoint जैसे अलग network dependencies लोड कर सकता है; इसलिए हर UI request केवल Google तक सीमित होने का दावा नहीं किया जा सकता।
</details>

<details>
<summary><strong>11. Python या Node.js के बजाय Go भाषा क्यों?</strong></summary>

* **सिंगल स्टैटिक बाइनरी**: कोई virtualenv या `node_modules` की झंझट नहीं।
* **Startup और memory**: target build पर मापें; fixed cold-boot/RAM claim नहीं है।
* **Concurrency**: Go concurrent HTTP/SSE primitives का उपयोग करता है; capacity और latency target build तथा upstream session पर निर्भर हैं।
</details>

<details>
<summary><strong>12. क्या मैं बिना LiteLLM या प्रॉक्सी के Claude Code CLI को सीधे BOB Gemini Free से जोड़ सकता हूँ?</strong></summary>

Anthropic-shaped adapter में `POST /v1/messages` और documented SSE event names implement हैं, लेकिन यह native Claude inference नहीं है; tool execution और full client compatibility endpoint-specific हैं।

बस अपने टर्मिनल में ये वैरिएबल्स सेट करें और Claude Code शुरू करें:
```bash
export ANTHROPIC_BASE_URL=http://127.0.0.1:9610
export ANTHROPIC_API_KEY=none
claude
```
</details>

<details>
<summary><strong>13. क्या यह OpenAI Codex CLI (`openai/codex`) और AI Routers (LiteLLM / OpenRouter / Portkey) के साथ काम करता है?</strong></summary>

* **OpenAI Codex CLI**: selected `/v1/responses` और `/v1/chat/completions` adapter routes के लिए custom base URL सेट करें; broad native/full compatibility certified नहीं है।
* **LiteLLM / OpenRouter / Portkey / OneAPI**: `http://127.0.0.1:9610/v1` को कस्टम OpenAI अपस्ट्रीम के रूप में जोड़ें। गेटवे बिना किसी समस्या के सभी SSE डेल्टा चंक्स, रीज़निंग टोकन्स और यूसेज मीट्रिक्स रिटर्न करता है।
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
