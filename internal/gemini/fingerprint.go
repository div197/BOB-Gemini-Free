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
	seededRand  = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type Fingerprint struct {
	Profile profiles.ClientProfile
	Headers map[string]string
}

var highTrustProfiles = []Fingerprint{
	{
		Profile: profiles.Safari_IOS_17_0,
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			"Accept-Language": "en-US,en;q=0.9",
			"Sec-Fetch-Dest": "empty",
			"Sec-Fetch-Mode": "cors",
			"Sec-Fetch-Site": "same-origin",
		},
	},
	{
		Profile: profiles.Safari_16_0,
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Safari/605.1.15",
			"Accept-Language": "en-US,en;q=0.9",
			"Sec-Fetch-Dest": "empty",
			"Sec-Fetch-Mode": "cors",
			"Sec-Fetch-Site": "same-origin",
		},
	},
}

func ResolveFingerprint(name string) Fingerprint {
	name = strings.ToLower(name)
	if name == "" || name == "random" || name == "iphone" || name == "ios" {
		return highTrustProfiles[seededRand.Intn(len(highTrustProfiles))]
	}

	switch name {
	case "chrome_120", "chrome":
		return Fingerprint{
			Profile: profiles.Chrome_120,
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
				"sec-ch-ua": "\"Not_A Brand\";v=\"8\", \"Chromium\";v=\"120\", \"Google Chrome\";v=\"120\"",
				"sec-ch-ua-mobile": "?0",
				"sec-ch-ua-platform": "\"Windows\"",
			},
		}
	case "firefox_120", "firefox":
		return Fingerprint{
			Profile: profiles.Firefox_120,
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:120.0) Gecko/20100101 Firefox/120.0",
				"Accept-Language": "en-US,en;q=0.5",
			},
		}
	}
	// Fallback to a random high trust profile
	return highTrustProfiles[0]
}

func getTLSClient(profileName string, timeoutSec int) (*tlsClientAdapter, error) {
	clientMapMu.Lock()
	defer clientMapMu.Unlock()

	cacheKey := fmt.Sprintf("%s:%d", profileName, timeoutSec)
	if adapter, ok := clientMap[cacheKey]; ok {
		return adapter, nil
	}

	fp := ResolveFingerprint(profileName)

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

	freq, err := fhttp.NewRequest(req.Method, req.URL.String(), bytes.NewReader(bodyBytes))
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
