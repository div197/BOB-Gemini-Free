package multimodal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/div197/bob-gemini-free/internal/gemini"
)

const MaxRemoteImageFetchBytes = 20 * 1024 * 1024

func UploadImage(client gemini.Requester, tokens PageTokens, imgBytes []byte, mime string, cookieCache *gemini.CookieCache, authUser string) (string, error) {
	if mime == "" {
		mime = "image/png"
	}

	// Optimize and compress large images if needed
	if compressed, newMime, err := CompressImageBytesIfNeeded(imgBytes, mime, MaxImageByteSize); err == nil && len(compressed) > 0 {
		imgBytes = compressed
		mime = newMime
	}

	pushID := tokens.PushID
	if pushID == "" {
		pushID = DefaultPushID
	}
	pctx := tokens.Pctx
	if pctx == "" {
		pctx = DefaultPctx
	}

	var cookieInfo gemini.CookieInfo
	if cookieCache != nil {
		cookieInfo, _ = cookieCache.Load()
	}

	// Step 1: Initiate resumable upload
	startHeaders := make(http.Header)
	startHeaders.Set("Push-ID", pushID)
	startHeaders.Set("X-Tenant-Id", "bard-storage")
	startHeaders.Set("X-Client-Pctx", pctx)
	startHeaders.Set("X-Goog-Upload-Header-Content-Length", fmt.Sprintf("%d", len(imgBytes)))
	startHeaders.Set("X-Goog-Upload-Header-Content-Type", mime)
	startHeaders.Set("X-Goog-Upload-Protocol", "resumable")
	startHeaders.Set("X-Goog-Upload-Command", "start")
	startHeaders.Set("Origin", "https://gemini.google.com")
	startHeaders.Set("Referer", fmt.Sprintf("https://gemini.google.com%s/app", gemini.AccountPrefix(authUser)))
	startHeaders.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	if cookieInfo.Cookie != "" {
		startHeaders.Set("Cookie", cookieInfo.Cookie)
	}
	if cookieInfo.SAPISID != "" {
		startHeaders.Set("Authorization", gemini.SAPISIDHash(cookieInfo.SAPISID))
	}
	if authUser != "" {
		startHeaders.Set("X-Goog-AuthUser", authUser)
	}

	startURL := "https://content-push.googleapis.com/upload/"
	req1, err := http.NewRequest("POST", startURL, bytes.NewReader(nil))
	if err != nil {
		return "", err
	}
	req1.Header = startHeaders

	resp1, err := client.Do(req1)
	if err != nil {
		return "", fmt.Errorf("Upload step 1 failed: %w", err)
	}
	defer resp1.Body.Close()

	uploadURL := resp1.Header.Get("X-Goog-Upload-URL")
	if uploadURL == "" {
		uploadURL = resp1.Header.Get("x-goog-upload-url")
	}
	if uploadURL == "" {
		return "", fmt.Errorf("No upload URL in response headers")
	}

	// Step 2: Upload file data + finalize
	uploadHeaders := make(http.Header)
	uploadHeaders.Set("X-Goog-Upload-Command", "upload, finalize")
	uploadHeaders.Set("X-Goog-Upload-Offset", "0")
	uploadHeaders.Set("Content-Type", "application/octet-stream")
	uploadHeaders.Set("Origin", "https://gemini.google.com")
	uploadHeaders.Set("Referer", fmt.Sprintf("https://gemini.google.com%s/app", gemini.AccountPrefix(authUser)))
	uploadHeaders.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	if cookieInfo.Cookie != "" {
		uploadHeaders.Set("Cookie", cookieInfo.Cookie)
	}
	if cookieInfo.SAPISID != "" {
		uploadHeaders.Set("Authorization", gemini.SAPISIDHash(cookieInfo.SAPISID))
	}
	if authUser != "" {
		uploadHeaders.Set("X-Goog-AuthUser", authUser)
	}

	req2, err := http.NewRequest("POST", uploadURL, bytes.NewReader(imgBytes))
	if err != nil {
		return "", err
	}
	req2.Header = uploadHeaders

	resp2, err := client.Do(req2)
	if err != nil {
		return "", fmt.Errorf("Upload step 2 failed: %w", err)
	}
	defer resp2.Body.Close()

	bodyBytes, err := io.ReadAll(resp2.Body)
	if err != nil {
		return "", err
	}

	fileRef := strings.TrimSpace(string(bodyBytes))
	if fileRef == "" || !strings.HasPrefix(fileRef, "/") {
		return "", fmt.Errorf("Invalid file reference: %s", fileRef)
	}

	return fileRef, nil
}

func FetchImageBytes(client gemini.Requester, imageURL string) ([]byte, error) {
	return fetchImageBytes(client, imageURL, net.LookupIP)
}

func fetchImageBytes(client gemini.Requester, imageURL string, lookupIP func(string) ([]net.IP, error)) ([]byte, error) {
	parsed, err := url.Parse(imageURL)
	if err != nil {
		return nil, fmt.Errorf("unsupported image URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported image URL scheme")
	}
	if parsed.Hostname() == "" || parsed.User != nil {
		return nil, fmt.Errorf("unsupported image URL")
	}
	if parsed.Port() != "" && parsed.Port() != "80" && parsed.Port() != "443" {
		return nil, fmt.Errorf("image URL port is not allowed")
	}
	if err := validatePublicImageHost(parsed.Hostname(), lookupIP); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := doImageRequest(client, req, parsed.Host)
	if err != nil {
		log.Printf("Image fetch failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("image fetch returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxRemoteImageFetchBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > MaxRemoteImageFetchBytes {
		return nil, fmt.Errorf("image fetch exceeded %d bytes", MaxRemoteImageFetchBytes)
	}

	detectedType := http.DetectContentType(body)
	if !strings.HasPrefix(detectedType, "image/") {
		return nil, fmt.Errorf("image fetch returned non-image content type %q", detectedType)
	}

	return body, nil
}

func validatePublicImageHost(host string, lookupIP func(string) ([]net.IP, error)) error {
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicImageIP(ip) {
			return fmt.Errorf("image URL resolves to a private or local address")
		}
		return nil
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return fmt.Errorf("image URL resolves to a local hostname")
	}
	ips, err := lookupIP(host)
	if err != nil {
		return fmt.Errorf("image URL host lookup failed: %w", err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("image URL host has no address")
	}
	for _, ip := range ips {
		if !isPublicImageIP(ip) {
			return fmt.Errorf("image URL resolves to a private or local address")
		}
	}
	return nil
}

func isPublicImageIP(ip net.IP) bool {
	return ip != nil && ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsUnspecified()
}

func doImageRequest(client gemini.Requester, req *http.Request, originalHost string) (*http.Response, error) {
	if httpClient, ok := client.(*http.Client); ok {
		clone := *httpClient
		clone.CheckRedirect = func(next *http.Request, _ []*http.Request) error {
			if !strings.EqualFold(next.URL.Host, originalHost) {
				return fmt.Errorf("image redirect changed host")
			}
			return http.ErrUseLastResponse
		}
		return clone.Do(req)
	}
	return client.Do(req)
}
