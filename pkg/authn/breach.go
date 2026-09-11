package authn

import (
	"bufio"
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Screening a password against the Have I Been Pwned corpus.
//
// This is the check that a length rule cannot make. "Sommer2024!!" satisfies any
// plausible minimum and has been seen in breaches thousands of times; a rule
// counting characters will never tell the two apart.
//
// The password is not sent anywhere. Its SHA-1 is computed locally and only the
// first five hex characters are transmitted; the service answers with every
// suffix sharing that prefix — several hundred of them — and the match is made
// here. The service therefore learns a bucket that contains roughly one in a
// million passwords, and never which one.

const pwnedRangeURL = "https://api.pwnedpasswords.com/range/"

// BreachChecker asks how often a password appears in known breaches.
type BreachChecker struct {
	// BaseURL is the range endpoint. Empty means the public service.
	BaseURL string
	Client  *http.Client
}

// Count reports how many times the password appears in the corpus. Zero means it
// does not appear, which is not a promise that the password is good — only that
// it is not one of the ones already known.
func (c *BreachChecker) Count(ctx context.Context, password string) (int, error) {
	sum := sha1.Sum([]byte(password))
	hash := strings.ToUpper(fmt.Sprintf("%x", sum))
	prefix, suffix := hash[:5], hash[5:]

	base := c.BaseURL
	if base == "" {
		base = pwnedRangeURL
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+prefix, nil)
	if err != nil {
		return 0, err
	}
	// Padding makes every response a similar size, so an observer cannot infer
	// anything from how much came back.
	req.Header.Set("Add-Padding", "true")

	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("checking the password against known breaches: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("checking the password against known breaches: %s", resp.Status)
	}
	return countInRange(resp.Body, suffix)
}

// countInRange finds our suffix among the ones returned.
func countInRange(body io.Reader, suffix string) (int, error) {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line, count, ok := strings.Cut(strings.TrimSpace(scanner.Text()), ":")
		if !ok {
			continue
		}
		if !strings.EqualFold(line, suffix) {
			continue
		}
		n, err := strconv.Atoi(count)
		if err != nil {
			return 0, fmt.Errorf("unreadable count for a breached password: %w", err)
		}
		// Padding entries are returned with a count of zero and mean nothing.
		return n, nil
	}
	return 0, scanner.Err()
}
