# Multi-Lingual i18n and Pedagogical Architecture for BOB Gemini Free

**Implementation status:** English and Hindi UI translations are present. The
Indic script selector exposes eight script codes, but it is not a complete
eight-language UI or an offline translation engine. Transliteration may call
Google Input Tools and has a small local fallback.

> **"विद्ययाऽमृतमश्नुते" (Through True Knowledge, One Attains Immortality)**  
> *Radical accessibility: Bringing state-of-the-art AI, first-principles systems engineering, and pedagogical clarity to every student in their native mother tongue.*

---

## 1. Executive Summary & Vision

AI must not be an exclusive domain restricted by linguistic barriers or financial gatekeeping. **BOB Gemini Free** was created by **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)). The browser UI is local-first and does not require a BOB account, but provider requests and optional transliteration services remain network- and session-dependent.

The current implementation establishes a foundation for a multi-lingual
internationalization (i18n) and pedagogical subsystem. It currently provides
English/Hindi UI coverage plus selectable Indic transliteration targets; it does
not establish full UI translation or native-language model capability for
every listed language.

---

## 2. Core Architectural Pillars

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           🌐 Universal i18n Engine                           │
├─────────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────┐  ┌──────────────────────┐  ┌────────────────────┐ │
│  │   Language Registry  │  │ Dynamic DOM Mapper   │  │   Phonetic Engine  │ │
│  │   (I18N_REGISTRY)    │  │ (data-i18n / t(key)) │  │ (Indic Translit)   │ │
│  └──────────┬───────────┘  └──────────┬───────────┘  └─────────┬──────────┘ │
│             │                         │                        │            │
│             ▼                         ▼                        ▼            │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                  📚 Student AI & Systems Knowledge Base                 │ │
│  │       (First-Principles Definitions • Physical Analogies • Proofs)     │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                                       │                                     │
│                                       ▼                                     │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                      💬 "Ask the Teacher" Pipeline                     │ │
│  │         (Socratic Prompts • LaTeX Math • Mermaid Visual Flows)         │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Pillar 1: Bilingual Base (`en` + `hi`)
- The current `I18N` registry supplies English and Hindi values for the
  implemented UI surfaces. New dialogs, error paths, and future features still
  require translation review; 100% synchronization is not assumed.
- Pure zero-cliché explanations: avoiding tired LEGO tropes in favor of intuitive physical models (wave optics, statistical thermodynamics, vector spaces, and graph paths).

### Pillar 2: The Interactive AI & Systems Glossary ("Ask the Teacher")
- Built directly into the client runtime (`#glossary-modal`).
- 9 Core First-Principles Knowledge Cards:
  1. **Tokens & Byte-Pair Encoding (BPE)**: Atomic subword units ($1\text{ Token} \approx 4\text{ Bytes}$).
  2. **Vector Embeddings & Semantic Geometry**: High-dimensional continuous manifolds and cosine angle proximity ($\cos\theta$).
  3. **Sampling Temperature & Softmax Entropy**: Probability sharpness and thermodynamic distribution ($P(x_i) = \frac{\exp(z_i / T)}{\sum \exp(z_j / T)}$).
  4. **Step-by-Step Reasoning (@think)**: Explicit hidden thought token emission and sequential inference chains.
  5. **Self-Attention & Transformer Architecture**: Dynamic pairwise key-query-value routing with $\sqrt{d_k}$ normalization.
  6. **Server-Sent Events (SSE)**: Unidirectional HTTP streaming protocol (`text/event-stream`) over persistent TCP sockets.
  7. **Web RPC Protocol Bridges**: Translating OpenAI/Anthropic REST into Google's internal web RPC (`streamGenerateContent`).
  8. **Multimodal Vision & Scotty Storage**: Session-signed blob uploads and token caching.
  9. **Cloud vs Free Economics**: Real-time calculation of commercial API rates saved by local gateway routing.

### Pillar 3: "Ask BOB Teacher in Chat" One-Click Socratic Bridge
- Clicking **"💬 Ask Teacher"** (or **"💬 शिक्षक से पूछें"**) on any concept automatically loads a custom-tailored Socratic inquiry prompt into the chat composer, instructing BOB to formulate intuitive physical proofs and mathematical derivations in real time.

---

## 3. Universal Language Registry & Expansion Roadmap

### Indic Language Tier (Planned Rollout)
| Code | Language | Script | Phonetic Transliteration Engine | Status |
| :--- | :--- | :--- | :--- | :--- |
| `en` | English | Latin | N/A | **Active (Tier 1)** |
| `hi` | हिन्दी (Hindi) | Devanagari | `hi-t-i0-und` | UI active; transliteration target |
| `sa` | संस्कृतम् (Sanskrit) | Devanagari | `sa-t-i0-und` | Selector target; partial local fallback |
| `mr` | मराठी (Marathi) | Devanagari | `mr-t-i0-und` | Selector target; partial local fallback |
| `bn` | বাংলা (Bengali) | Eastern Nagari | `bn-t-i0-und` | Selector target; provider-dependent |
| `gu` | ગુજરાતી (Gujarati) | Gujarati | `gu-t-i0-und` | Selector target; provider-dependent |
| `ta` | தமிழ் (Tamil) | Tamil | `ta-t-i0-und` | Selector target; provider-dependent |
| `te` | తెలుగు (Telugu) | Telugu | `te-t-i0-und` | Selector target; provider-dependent |
| `pa` | ਪੰਜਾਬੀ (Punjabi) | Gurmukhi | `pa-t-i0-und` | Selector target; provider-dependent |

### Global UI Language Tier
| Code | Language | Native Title | Status |
| :--- | :--- | :--- | :--- |
| `es` | Spanish | Español | Not implemented in current `I18N` registry |
| `ja` | Japanese | 日本語 | Not implemented in current `I18N` registry |
| `de` | German | Deutsch | Not implemented in current `I18N` registry |
| `fr` | French | Français | Not implemented in current `I18N` registry |
| `ru` | Russian | Русский | Not implemented in current `I18N` registry |
| `ar` | Arabic | العربية | Not implemented in current `I18N` registry |
| `zh` | Chinese (Simplified) | 简体中文 | Not implemented in current `I18N` registry |

---

## 4. Technical Implementation Pattern for New Languages

To add a new language pack to `I18N_REGISTRY`:

```javascript
// Registering a new language pack (e.g. Tamil 'ta')
I18N.ta = {
  brandTagline: "இயல்பான எல்லைகளை உடைகங்கள் • ABCSTEPS",
  statUptime: "இயங்கும் நேரம்:",
  statRequests: "கோரிக்கைகள்:",
  statTokens: "டோக்கன்கள்:",
  statSavings: "சேமிப்பு:",
  ui: {
    navConfig: "கட்டமைப்பு",
    navCode: "குறியீடு",
    navMenu: "மெனு",
    glossaryNav: "சொற்களஞ்சியம்",
    glossary: {
      title: "மாணவர் AI & அமைப்புகள் சொற்களஞ்சியம்",
      badge: "ஆசிரியரிடம் கேளுங்கள் • முதல் கொள்கைகள்",
      subtitle: "முதல் கொள்கை விளக்கங்கள் • இயற்பியல் ஒப்புமைகள்",
      searchPlaceholder: "சொல்லைத் தேடுங்கள் (எ.கா. டோக்கன், வெப்பநிலை)...",
      filterAll: "அனைத்தும்",
      filterLLM: "முக்கிய LLM",
      filterSystems: "அமைப்புகள் & RPC",
      filterEconomics: "பொருளாதாரம்",
      footerHint: "💡 உதவிக்குறிப்பு: விவாதத்தை தொடங்க <b>\"💬 ஆசிரியரிடம் கேளுங்கள்\"</b> என்பதை அழுத்தவும்!",
      close: "மூடு"
    }
  }
};
```

---

## 5. Verification & Testing

The current evidence is narrower:
1. The source registry and selector entries are inspected in
   `internal/server/playground.html`.
2. The full Go test suite checks the Go packages, not every browser language
   string or viewport.
3. A release-quality language expansion still needs browser tests at target
   viewports, native-script typography review, fallback/error tests, and a
   privacy review of any external transliteration request.

---

*Authored with Nishkaam Karma by ABCsteps & Divyanshu Singh Chouhan for global learners.*
