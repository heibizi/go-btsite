package btsite

import (
	"fmt"
	"net/url"
)

// JoinURL 将基准 URL 和相对 URL 拼接在一起，返回完整的 URL
func JoinURL(baseURL, relativeURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing base URL: %v", err)
	}

	rel, err := url.Parse(relativeURL)
	if err != nil {
		return "", fmt.Errorf("error parsing relative URL: %v", err)
	}

	// 组合 URL
	resolvedURL := base.ResolveReference(rel)
	return resolvedURL.String(), nil
}
