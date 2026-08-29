package multimodal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/div197/bob-gemini-free/internal/gemini"
)

const (
	MaxRemoteImageFetchBytes = 20 * 1024 * 1024
	MaxUploadResponseBytes   = 64 * 1024
	MaxUploadFileRefBytes    = 4 * 1024
)

func UploadImage(client gemini.Requester, tokens PageTokens, imgBytes []byte, mime string, cookieCache *gemini.CookieCache, authUser string) (string, error) {
	return UploadImageContext(context.Background(), client, tokens, imgBytes, mime, cookieCache, authUser)
}

// UploadImageContext binds both upload requests to the caller's lifecycle so
// a disconnected gateway request does not continue sending image bytes.
func UploadImageContext(ctx context.Context, client gemini.Requester, tokens PageTokens, imgBytes []byte, mime string, cookieCache *gemini.CookieCache, authUser string) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if client == nil {
		return "", fmt.Errorf("image upload requires an HTTP client")
	}
	if err := validateImageData(imgBytes); err != nil {
		return "", fmt.Errorf("image upload rejected: %w", err)
	}
	if mime == "" {
		mime = "image/png"
	}

	// Optimize and compress large images if needed
	if compressed, newMime, err := CompressImageBytesIfNeeded(imgBytes, mime, MaxImageByteSize); err == nil && len(compressed) > 0 {
		imgBytes = compressed
		mime = newMime
	} else if err != nil {
		return "", fmt.Errorf("image upload preparation failed: %w", err)
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
	req1, err := http.NewRequestWithContext(ctx, "POST", startURL, bytes.NewReader(nil))
	if err != nil {
		return "", err
	}
	req1.Header = startHeaders

	resp1, err := client.Do(req1)
	if err != nil {
		return "", fmt.Errorf("Upload step 1 failed: %w", err)
	}
	if resp1 == nil {
		return "", fmt.Errorf("Upload step 1 returned an empty response")
	}
	if resp1.Body != nil {
		defer resp1.Body.Close()
	}
	if resp1.StatusCode < http.StatusOK || resp1.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("Upload step 1 returned HTTP %d", resp1.StatusCode)
	}

	uploadURL := resp1.Header.Get("X-Goog-Upload-URL")
	if uploadURL == "" {
		uploadURL = resp1.Header.Get("x-goog-upload-url")
	}
	if uploadURL == "" {
		return "", fmt.Errorf("No upload URL in response headers")
	}
	if err := validateScottyUploadURL(uploadURL); err != nil {
		return "", err
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

	req2, err := http.NewRequestWithContext(ctx, "POST", uploadURL, bytes.NewReader(imgBytes))
	if err != nil {
		return "", err
	}
	req2.Header = uploadHeaders

	resp2, err := doScottyUpload(client, req2)
	if err != nil {
		return "", fmt.Errorf("Upload step 2 failed: %w", err)
	}
	if resp2 == nil {
		return "", fmt.Errorf("Upload step 2 returned an empty response")
	}
	if resp2.Body == nil {
		return "", fmt.Errorf("Upload step 2 returned an empty body")
	}
	defer resp2.Body.Close()
	if resp2.StatusCode < http.StatusOK || resp2.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("Upload step 2 returned HTTP %d", resp2.StatusCode)
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp2.Body, MaxUploadResponseBytes+1))
	if err != nil {
		return "", err
	}
	if len(bodyBytes) > MaxUploadResponseBytes {
		return "", fmt.Errorf("Upload step 2 response exceeded %d bytes", MaxUploadResponseBytes)
	}

	fileRef := strings.TrimSpace(string(bodyBytes))
	if fileRef == "" || len(fileRef) > MaxUploadFileRefBytes || !strings.HasPrefix(fileRef, "/") || strings.ContainsAny(fileRef, "\r\n\x00") {
		return "", fmt.Errorf("invalid file reference returned by upload service")
	}

	return fileRef, nil
}

func validateScottyUploadURL(rawURL string) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Scheme != "https" || parsed.User != nil || parsed.Fragment != "" {
		return fmt.Errorf("upload service returned an invalid upload URL")
	}
	if !strings.EqualFold(parsed.Hostname(), "content-push.googleapis.com") {
		return fmt.Errorf("upload service returned an untrusted upload host")
	}
	if parsed.Port() != "" && parsed.Port() != "443" {
		return fmt.Errorf("upload service returned an invalid upload port")
	}
	if !strings.HasPrefix(parsed.Path, "/upload/") {
		return fmt.Errorf("upload service returned an invalid upload path")
	}
	return nil
}

func doScottyUpload(client gemini.Requester, req *http.Request) (*http.Response, error) {
	if httpClient, ok := client.(*http.Client); ok {
		clone := *httpClient
		clone.CheckRedirect = func(next *http.Request, _ []*http.Request) error {
			return fmt.Errorf("upload service redirect refused for %s", next.URL.Host)
		}
		return clone.Do(req)
	}
	return client.Do(req)
}

func FetchImageBytes(client gemini.Requester, imageURL string) ([]byte, error) {
	return FetchImageBytesContext(context.Background(), client, imageURL)
}

func FetchImageBytesContext(ctx context.Context, client gemini.Requester, imageURL string) ([]byte, error) {
	if client == nil {
		return nil, fmt.Errorf("image fetch requires an HTTP client")
	}
	if _, ok := client.(*http.Client); !ok {
		return nil, fmt.Errorf("image fetch requires a guardable HTTP client")
	}
	return fetchImageBytesContext(ctx, client, imageURL, net.LookupIP)
}

func fetchImageBytes(client gemini.Requester, imageURL string, lookupIP func(string) ([]net.IP, error)) ([]byte, error) {
	return fetchImageBytesContext(context.Background(), client, imageURL, lookupIP)
}

func fetchImageBytesContext(ctx context.Context, client gemini.Requester, imageURL string, lookupIP func(string) ([]net.IP, error)) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if client == nil {
		return nil, fmt.Errorf("image fetch requires an HTTP client")
	}
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

	req, err := http.NewRequestWithContext(ctx, "GET", parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := doImageRequest(client, req, parsed.Host, lookupIP)
	if err != nil {
		// Do not log the source URL or transport error: image URLs can contain
		// signed query parameters, and net/http errors may echo the full URL.
		log.Printf("Image fetch failed")
		return nil, err
	}
	if resp == nil {
		return nil, fmt.Errorf("image fetch returned an empty response")
	}
	if resp.Body == nil {
		return nil, fmt.Errorf("image fetch returned an empty body")
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
	_, err := publicImageIPs(host, lookupIP)
	return err
}

func publicImageIPs(host string, lookupIP func(string) ([]net.IP, error)) ([]net.IP, error) {
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicImageIP(ip) {
			return nil, fmt.Errorf("image URL resolves to a private or local address")
		}
		return []net.IP{ip}, nil
	}
	if strings.EqualFold(host, "localhost") || strings.HasSuffix(strings.ToLower(host), ".localhost") {
		return nil, fmt.Errorf("image URL resolves to a local hostname")
	}
	if lookupIP == nil {
		lookupIP = net.LookupIP
	}
	ips, err := lookupIP(host)
	if err != nil {
		return nil, fmt.Errorf("image URL host lookup failed: %w", err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("image URL host has no address")
	}
	for _, ip := range ips {
		if !isPublicImageIP(ip) {
			return nil, fmt.Errorf("image URL resolves to a private or local address")
		}
	}
	return ips, nil
}

func isPublicImageIP(ip net.IP) bool {
	return ip != nil && ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() && !ip.IsUnspecified()
}

func doImageRequest(client gemini.Requester, req *http.Request, originalHost string, lookupIP func(string) ([]net.IP, error)) (*http.Response, error) {
	if httpClient, ok := client.(*http.Client); ok {
		clone := *httpClient
		transport := clone.Transport
		if transport == nil {
			transport = http.DefaultTransport
		}
		baseTransport, ok := transport.(*http.Transport)
		if !ok {
			return nil, fmt.Errorf("image fetch requires a guardable HTTP transport")
		}
		guardedTransport := baseTransport.Clone()
		// A proxy can resolve and fetch the destination independently of the
		// gateway's SSRF check. Remote images therefore use a direct, guarded
		// connection; the configured proxy remains available for provider RPCs.
		guardedTransport.Proxy = nil
		guardedTransport.DialTLSContext = nil
		guardedTransport.DialTLS = nil
		host := req.URL.Hostname()
		port := req.URL.Port()
		if port == "" {
			if strings.EqualFold(req.URL.Scheme, "https") {
				port = "443"
			} else {
				port = "80"
			}
		}
		guardedTransport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialPublicImageHost(ctx, network, host, port, lookupIP)
		}
		clone.Transport = guardedTransport
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

// dialPublicImageHost performs the final DNS lookup immediately before the
// socket connection and dials the approved literal address. This removes the
// validation-to-connect DNS TOCTOU window: a later private answer cannot be
// handed back to net/http for a second, unconstrained resolution.
func dialPublicImageHost(ctx context.Context, network, host, port string, lookupIP func(string) ([]net.IP, error)) (net.Conn, error) {
	ips, err := publicImageIPs(host, lookupIP)
	if err != nil {
		return nil, err
	}
	dialer := net.Dialer{}
	var lastErr error
	for _, ip := range ips {
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	if lastErr != nil {
		return nil, fmt.Errorf("image host connection failed: %w", lastErr)
	}
	return nil, fmt.Errorf("image URL host has no usable address")
}
