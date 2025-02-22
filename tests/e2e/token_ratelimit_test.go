//go:build test_e2e

package e2e

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/openai/openai-go"
	"github.com/stretchr/testify/require"
)

func Test_Examples_TokenRateLimit(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	const manifest = "../../examples/token_ratelimit/token_ratelimit.yaml"
	require.NoError(t, kubectlApplyManifest(ctx, manifest))

	const egSelector = "gateway.envoyproxy.io/owning-gateway-name=envoy-ai-gateway-token-ratelimit"
	requireWaitForPodReady(t, egNamespace, egSelector)

	makeRequest := func(usedID string, input, output, total int, expStatus int) {
		fwd := requireNewHTTPPortForwarder(t, egNamespace, egSelector, egDefaultPort)
		defer fwd.kill()

		requestBody := fmt.Sprintf(`{"messages":[{"role":"user","content":"Say this is a test"}],"model":"gpt-4o-mini"}`)

		const fakeResponseBodyTemplate = `{"choices":[{"message":{"content":"This is a test.","role":"assistant"}}],"stopReason":null,"usage":{"prompt_tokens":%d,"completion_tokens":%d,"total_tokens":%d}}`
		fakeResponseBody := fmt.Sprintf(fakeResponseBodyTemplate, input, output, total)

		req, err := http.NewRequest(http.MethodPut, fwd.address()+"/v1/chat/completions", strings.NewReader(requestBody))
		require.NoError(t, err)
		req.Header.Set("x-response-body", base64.StdEncoding.EncodeToString([]byte(fakeResponseBody)))
		req.Header.Set("x-expected-path", base64.StdEncoding.EncodeToString([]byte("/v1/chat/completions")))
		req.Header.Set("x-user-id", usedID)
		req.Header.Set("Host", "openai.com")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		for key, values := range resp.Header {
			t.Logf("key: %s, values: %v\n", key, values)
		}

		body, err := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusOK {
			var oaiBody openai.ChatCompletion
			require.NoError(t, json.Unmarshal(body, &oaiBody))
			// Sanity check the fake response is correctly parsed.
			require.Equal(t, "This is a test.", oaiBody.Choices[0].Message.Content)
			require.Equal(t, int64(input), oaiBody.Usage.PromptTokens)
			require.Equal(t, int64(output), oaiBody.Usage.CompletionTokens)
			require.Equal(t, int64(total), oaiBody.Usage.TotalTokens)
		}
		require.Equal(t, expStatus, resp.StatusCode)
	}

	// Test the input token limit.
	baseID := int(time.Now().UnixNano()) // To avoid collision with previous runs.
	usedID := strconv.Itoa(baseID)
	makeRequest(usedID, 10000, 0, 0, 200)
	makeRequest(usedID, 1, 0, 0, 429)

	// Test the output token limit.
	usedID = strconv.Itoa(baseID + 1)
	makeRequest(usedID, 0, 20, 0, 200) // This output number exceeds the input limit, but should still be allowed.
	makeRequest(usedID, 0, 10000, 0, 200)
	makeRequest(usedID, 0, 1, 0, 429)

	// Test the total token limit.
	usedID = strconv.Itoa(baseID + 2)
	makeRequest(usedID, 0, 0, 20, 200)  // This total number exceeds the input limit, but should still be allowed.
	makeRequest(usedID, 0, 0, 200, 200) // This total number exceeds the output limit, but should still be allowed.
	makeRequest(usedID, 0, 0, 1000000, 200)
	makeRequest(usedID, 0, 0, 1, 429)
}
