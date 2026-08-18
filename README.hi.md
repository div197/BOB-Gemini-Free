<p align="center">
  <img src="assets/bob-gemini-free-banner.jpg" alt="BOB Gemini Free Banner" width="100%">
</p>

# BOB Gemini Free (बॉब जेमिनी फ्री)

<p align="center">
  <strong>Break Ordinary Boundaries — हाई-परफॉर्मेंस लोकल OpenAI और Gemini API गेटवे</strong><br>
  <em>गूगल जेमिनी (Google Gemini) वेब इंटरफ़ेस द्वारा संचालित</em>
</p>

<p align="center">
  <a href="https://abcsteps.com/"><img src="https://img.shields.io/badge/Powered%20by-ABCsteps.com-blue?style=for-the-badge" alt="ABCsteps"></a>
  <a href="https://github.com/div197/bob-gemini-free"><img src="https://img.shields.io/badge/Author-Divyanshu%20Singh%20Chouhan%20(@div197)-green?style=for-the-badge" alt="Author"></a>
  <img src="https://img.shields.io/badge/Release-v0.1.0-blueviolet?style=for-the-badge" alt="Release">
  <img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/License-MIT-orange?style=for-the-badge" alt="License">
</p>

---

[English](README.md) | [हिंदी (Hindi)](README.hi.md) | [बदलाव सूची (Changelog)](CHANGELOG.md)

**BOB Gemini Free** (बॉब जेमिनी फ्री), [**ABCsteps.com**](https://abcsteps.com/) — **दिव्यांशु सिंह चौहान** ([@div197](https://github.com/div197)) द्वारा स्थापित ऑनलाइन एआई इंजीनियरिंग स्कूल — की **BOB सीरीज़** (*Break Ordinary Boundaries*) का एक प्रमुख उत्पाद है।

यह एक हाई-परफॉर्मेंस, सिंगल-फाइल Go गेटवे है जो गूगल जेमिनी के वेब इंटरफ़ेस को स्टैंडर्ड **OpenAI-कम्पैटिबल** (`/v1/chat/completions`, `/v1/models`, `/v1/responses`) और **Gemini-नेटिव** (`/v1beta/models`) API एंडपॉइंट्स में बदल देता है।

---

## BOB सीरीज़ क्या है? (*Break Ordinary Boundaries*)

ABCsteps की **BOB सीरीज़** का उद्देश्य डेवलपर्स, छात्रों और एआई इंजीनियरों के लिए बिना किसी पेवॉल या महँगे सब्सक्रिप्शन के शक्तिशाली टूल्स और रनटाइम्स उपलब्ध कराना है:

* 🎥 [**BOB YouTube**](https://github.com/div197/BOB-Youtube) — डेवलपर्स और एआई एजेंट्स के लिए डॉकर-फर्स्ट यूट्यूब डेटा इंजेक्शन और प्रोसेसिंग टूल।
* ⚡ [**BOB Gemini Free**](https://github.com/div197/bob-gemini-free) — गूगल जेमिनी वेब की पूरी शक्ति को OpenAI फॉर्मेट में अनलॉक करने वाला लोकल गेटवे।

---

## मुख्य विशेषताएँ

* **प्रत्येक जीमेल (Gmail) यूज़र के लिए मुफ़्त**: दुनिया के हर व्यक्ति के पास जीमेल के साथ जेमिनी का मुफ़्त एक्सेस होता है। यह गेटवे आपको बिना किसी अतिरिक्त शुल्क के Flash, Flash Lite और Deep Thinking (20,000+ अक्षरों की विस्तृत रीज़निंग) का उपयोग करने देता है।
* **Gemini Advanced ($20/माह) प्रो अनलॉक**: यदि आपके पास गूगल का पेड सब्सक्रिप्शन है, तो अपनी लोकल कुकी जोड़कर गूगल के फ्लैगशिप **Pro** मॉडल को एक्टिवेट करें।
* **OpenAI ड्रॉप-इन रिप्लेसमेंट**: Cherry Studio, ChatBox, Codex CLI, Cursor, तथा OpenAI के Python / Node.js SDKs के साथ सीधे इस्तेमाल करें।
* **मल्टीमॉडल विज़न (Vision)**: OpenAI फॉर्मेट में Base64 इमेज या इमेज लिंक्स भेजें — यह टूल ऑटोमैटिक कम्प्रेशन और गूगल के स्कॉटी रेज़्युमेबल अपलोड का उपयोग करता है।
* **रीज़निंग कंट्रोल (`@think=N`)**: मॉडल नाम के आगे `@think=N` लगाकर सोचने की गहराई को नियंत्रित करें (`@think=0` = सबसे गहरी सोच, `@think=4` = तेज़ व संक्षिप्त उत्तर)।
* **सुरक्षित व लोकल-फर्स्ट**: डिफ़ॉल्ट रूप से `127.0.0.1` पर बाइंड होता है, जिससे आपके क्रेडेंशियल्स पूरी तरह से आपके कंप्यूटर पर ही सुरक्षित रहते हैं।

---

## त्वरित शुरुआत (Quick Start)

### 1. यदि आपके कंप्यूटर पर Go इंस्टॉल नहीं है (Non-Go Users)

#### विकल्प A: डॉकर (Docker) द्वारा चलाएँ

```bash
docker build -t bob-gemini-free .
docker run -d --name bob-gemini-free -p 8081:8081 bob-gemini-free
```

#### विकल्प B: ऑटोमैटिक इंस्टॉलर स्क्रिप्ट (macOS / Linux)

```bash
chmod +x install.sh
./install.sh
```

---

### 2. सोर्स कोड से निर्माण (Build from Source - Go 1.22+)

```bash
# रिपॉजिटरी क्लोन करें
git clone https://github.com/div197/bob-gemini-free.git
cd bob-gemini-free

# सिंगल बाइनरी बनाएँ
go build -o bob-gemini-free .

# सर्वर शुरू करें
./bob-gemini-free --port 8081
```

सर्वर `http://127.0.0.1:8081/v1` पर शुरू हो जाएगा।

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

## Pro मॉडल के लिए कुकी (Gemini Advanced)

1. क्रोम ब्राउज़र में [gemini.google.com](https://gemini.google.com) खोलें और अपने सब्सक्राइब्ड अकाउंट से लॉगिन करें।
2. डेवलपर टूल्स (`F12`) → **Application** → **Cookies** में जाएँ।
3. अपनी कुकीज़ कॉपी करें: `SID`, `HSID`, `SSID`, `APISID`, `SAPISID`, `__Secure-1PSID`।
4. `~/.config/bob-gemini-free/cookie.txt` में सेव करें:
   ```text
   SID=your_sid; HSID=your_hsid; SSID=your_ssid; APISID=your_apisid; SAPISID=your_sapisid; __Secure-1PSID=your_1psid
   ```
5. फ़ाइल पर सुरक्षित परमिशन लगाएँ: `chmod 600 ~/.config/bob-gemini-free/cookie.txt`।
6. सर्वर चालू करें: `./bob-gemini-free --cookie-file ~/.config/bob-gemini-free/cookie.txt`।

---

## ABCsteps के बारे में

[**ABCsteps**](https://abcsteps.com/) जोधपुर, राजस्थान, भारत में स्थापित एक ऑनलाइन एआई इंजीनियरिंग स्कूल है जिसकी स्थापना **दिव्यांशु सिंह चौहान** द्वारा की गई है।

* [**20-पाठों का AI इंजीनियरिंग पाठ्यक्रम**](https://abcsteps.com/offerings/)
* [**प्रैक्टिकल इंजीनियरिंग ब्लॉग**](https://abcsteps.com/blog/)
* [**रीडिंग पाथ्स और शब्दावली**](https://abcsteps.com/blog/paths/)

अधिक जानकारी के लिए देखें: [https://abcsteps.com/](https://abcsteps.com/)।

---

## लाइसेंस

MIT लाइसेंस। [ABCsteps.com](https://abcsteps.com/) द्वारा विकसित।
