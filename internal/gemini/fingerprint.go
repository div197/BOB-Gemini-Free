package gemini

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

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
)

func resolveProfile(name string) profiles.ClientProfile {
	switch strings.ToLower(name) {
	case "chrome_120":
		return profiles.Chrome_120
	case "chrome_124":
		return profiles.Chrome_124
	case "chrome_131":
		return profiles.Chrome_131
	case "chrome_133":
		return profiles.Chrome_133
	case "chrome_144":
		return profiles.Chrome_144
	case "chrome_146":
		return profiles.Chrome_146
	case "firefox_120":
		return profiles.Firefox_120
	case "firefox_123":
		return profiles.Firefox_123
	case "firefox_147":
		return profiles.Firefox_147
	case "safari_16_0":
		return profiles.Safari_16_0
	case "safari_ios_17_0":
		return profiles.Safari_IOS_17_0
	default:
		return profiles.Chrome_146
	}
}

func getTLSClient(profileName string, timeoutSec int) (*tlsClientAdapter, error) {
	clientMapMu.Lock()
	defer clientMapMu.Unlock()

	cacheKey := fmt.Sprintf("%s:%d", profileName, timeoutSec)
	if adapter, ok := clientMap[cacheKey]; ok {
		return adapter, nil
	}

	options := []http_client.HttpClientOption{
		http_client.WithClientProfile(resolveProfile(profileName)),
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
