package gemini

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	http_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

type tlsClientAdapter struct {
	client http_client.HttpClient
}

var (
	clientMap   = make(map[string]*tlsClientAdapter)
	clientMapMu sync.Mutex
	// randMu protects seededRand: math/rand.Rand is NOT safe for concurrent use.
	randMu     sync.Mutex
	seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// safeRandIntn returns a random int in [0,n) in a concurrency-safe manner.
func safeRandIntn(n int) int {
	randMu.Lock()
	v := seededRand.Intn(n)
	randMu.Unlock()
	return v
}

type Fingerprint struct {
	Profile profiles.ClientProfile
	Headers map[string]string
}

var highTrustProfiles = []Fingerprint{
	{
		Profile: profiles.Safari_IOS_17_0,
		Headers: map[string]string{
			"User-Agent":      "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			"Accept-Language": "en-US,en;q=0.9",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	},
	{
		Profile: profiles.Safari_16_0,
		Headers: map[string]string{
			"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Safari/605.1.15",
			"Accept-Language": "en-US,en;q=0.9",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	},
}

func ResolveFingerprint(name string) (Fingerprint, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" || name == "random" {
		return highTrustProfiles[safeRandIntn(len(highTrustProfiles))], nil
	}
	if name == "iphone" || name == "ios" || name == "safari_ios_17_0" {
		return safariFingerprint(profiles.Safari_IOS_17_0, "iPhone; CPU iPhone OS 17_0 like Mac OS X", "17.0"), nil
	}

	switch name {
	case "chrome", "chrome_120":
		return chromeFingerprint(profiles.Chrome_120, "120"), nil
	case "chrome_124":
		return chromeFingerprint(profiles.Chrome_124, "124"), nil
	case "chrome_131":
		return chromeFingerprint(profiles.Chrome_131, "131"), nil
	case "chrome_133":
		return chromeFingerprint(profiles.Chrome_133, "133"), nil
	case "chrome_144":
		return chromeFingerprint(profiles.Chrome_144, "144"), nil
	case "chrome_146":
		return chromeFingerprint(profiles.Chrome_146, "146"), nil
	case "firefox", "firefox_120":
		return firefoxFingerprint(profiles.Firefox_120, "120"), nil
	case "firefox_123":
		return firefoxFingerprint(profiles.Firefox_123, "123"), nil
	case "firefox_147":
		return firefoxFingerprint(profiles.Firefox_147, "147"), nil
	case "safari", "safari_16_0":
		return safariFingerprint(profiles.Safari_16_0, "Macintosh", "16.0"), nil
	default:
		return Fingerprint{}, fmt.Errorf("unknown TLS fingerprint profile %q", name)
	}
}

func chromeFingerprint(profile profiles.ClientProfile, version string) Fingerprint {
	return Fingerprint{
		Profile: profile,
		Headers: map[string]string{
			"User-Agent":         fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s.0.0.0 Safari/537.36", version),
			"sec-ch-ua":          fmt.Sprintf("\"Not_A Brand\";v=\"8\", \"Chromium\";v=\"%s\", \"Google Chrome\";v=\"%s\"", version, version),
			"sec-ch-ua-mobile":   "?0",
			"sec-ch-ua-platform": "\"Windows\"",
		},
	}
}

func firefoxFingerprint(profile profiles.ClientProfile, version string) Fingerprint {
	return Fingerprint{
		Profile: profile,
		Headers: map[string]string{
			"User-Agent":      fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:%s) Gecko/20100101 Firefox/%s", version, version),
			"Accept-Language": "en-US,en;q=0.5",
		},
	}
}

func safariFingerprint(profile profiles.ClientProfile, platform, version string) Fingerprint {
	uaPlatform := platform
	if platform == "Macintosh" {
		uaPlatform = "Macintosh; Intel Mac OS X 10_15_7"
	}
	return Fingerprint{
		Profile: profile,
		Headers: map[string]string{
			"User-Agent":      fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/%s Safari/605.1.15", uaPlatform, version),
			"Accept-Language": "en-US,en;q=0.9",
			"Sec-Fetch-Dest":  "empty",
			"Sec-Fetch-Mode":  "cors",
			"Sec-Fetch-Site":  "same-origin",
		},
	}
}

func getTLSClient(profileName string, timeoutSec int) (*tlsClientAdapter, error) {
	clientMapMu.Lock()
	defer clientMapMu.Unlock()

	cacheKey := fmt.Sprintf("%s:%d", profileName, timeoutSec)
	if adapter, ok := clientMap[cacheKey]; ok {
		return adapter, nil
	}

	fp, err := ResolveFingerprint(profileName)
	if err != nil {
		return nil, err
	}

	options := []http_client.HttpClientOption{
		http_client.WithClientProfile(fp.Profile),
		http_client.WithTimeoutSeconds(timeoutSec),
		http_client.WithNotFollowRedirects(),
	}

	client, err := http_client.NewHttpClient(http_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, err
	}

	adapter := &tlsClientAdapter{client: client}
	clientMap[cacheKey] = adapter
	return adapter, nil
}

func (a *tlsClientAdapter) Do(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		_ = req.Body.Close()
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	freq, err := fhttp.NewRequestWithContext(req.Context(), req.Method, req.URL.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	for k, vv := range req.Header {
		for _, v := range vv {
			freq.Header.Add(k, v)
		}
	}

	fresp, err := a.client.Do(freq)
	if err != nil {
		return nil, err
	}

	resp := &http.Response{
		Status:        fresp.Status,
		StatusCode:    fresp.StatusCode,
		Proto:         fresp.Proto,
		ProtoMajor:    fresp.ProtoMajor,
		ProtoMinor:    fresp.ProtoMinor,
		Header:        make(http.Header),
		Body:          fresp.Body,
		ContentLength: fresp.ContentLength,
		Request:       req,
	}

	for k, vv := range fresp.Header {
		for _, v := range vv {
			resp.Header.Add(k, v)
		}
	}

	return resp, nil
}
