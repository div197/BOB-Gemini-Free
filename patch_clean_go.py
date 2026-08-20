with open('internal/gemini/parse.go', 'r') as f:
    code = f.read()

bad_clean = """var (
	reCodeExec    = regexp.MustCompile(`(?s)\x60\x60\x60(?:python|javascript|text)\?code_(?:reference|stdout)&code_event_index=\d+\n.*?\x60\x60\x60\n?`)
	reCardContent = regexp.MustCompile(`http://googleusercontent\.com/card_content/\d+\n?`)
	reBardError   = regexp.MustCompile(`BardErrorInfo(?:\s*|",\s*)\[(\d+)\]`)
)

func CleanText(text string, strip bool) string {
	text = reCodeExec.ReplaceAllString(text, "")
	text = reCardContent.ReplaceAllString(text, "")
	if strip {
		return strings.TrimSpace(text)
	}
	return text
}"""

good_clean = """var (
	reBardError = regexp.MustCompile(`BardErrorInfo(?:\s*|",\s*)\[(\d+)\]`)
)

func CleanText(raw string, strip bool) string {
	var out strings.Builder
	for len(raw) > 0 {
		idx1 := strings.Index(raw, "```")
		idx2 := strings.Index(raw, "http://googleusercontent.com/card_content/")

		minIdx := -1
		var tag, close string

		if idx2 >= 0 {
			minIdx = idx2
			tag = "http://googleusercontent.com/card_content/"
			close = "\n"
		}
		if idx1 >= 0 && (minIdx == -1 || idx1 < minIdx) {
			minIdx = idx1
			tag = "```"
			close = "```\n"
		}

		if minIdx >= 0 {
			out.WriteString(raw[:minIdx])
			raw = raw[minIdx:]

			if tag == "```" {
				nlIdx := strings.Index(raw, "\n")
				if nlIdx >= 0 {
					line := raw[:nlIdx+1]
					if strings.Contains(line, "?code_") {
						closeIdx := strings.Index(raw, close)
						if closeIdx >= 0 {
							raw = raw[closeIdx+len(close):]
						} else {
							raw = ""
						}
					} else {
						out.WriteString(raw[:3])
						raw = raw[3:]
					}
				} else {
					raw = ""
				}
			} else {
				closeIdx := strings.Index(raw, close)
				if closeIdx >= 0 {
					raw = raw[closeIdx+len(close):]
				} else {
					raw = ""
				}
			}
		} else {
			out.WriteString(raw)
			break
		}
	}
	
	res := out.String()
	if strip {
		return strings.TrimSpace(res)
	}
	return res
}"""

code = code.replace(bad_clean, good_clean)

with open('internal/gemini/parse.go', 'w') as f:
    f.write(code)
