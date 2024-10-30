package logger

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/gurleensethi/yurl/pkg/models"
	"github.com/gurleensethi/yurl/pkg/styles"
)

// LogRequest logs the request to the console.
func LogHttpRequest(ctx context.Context, request *models.HttpRequest) {
	// Print variables
	if len(request.Variables) > 0 {
		// Sort variables
		keys := make([]string, 0, len(request.Variables))
		for key := range request.Variables {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		fmt.Println(styles.SectionHeader.Render("Variables"))
		for _, key := range keys {
			value := request.Variables[key]
			name := styles.SecondaryText.Copy().Bold(true).Render(key)
			varValue := styles.PrimaryText.Render(fmt.Sprintf("%v", value.Value))
			source := value.Source
			fmt.Printf("%s: (%s) %s\n", name, source, varValue)
		}
	}

	fmt.Println(styles.SectionHeader.Render("Request"))

	protocol := styles.Url.Render(request.RawRequest.Proto)
	method := styles.Url.Render(request.RawRequest.Method)
	completeUrl := styles.Url.Render(request.RawRequest.URL.String())
	fmt.Printf("%s %s %s\n", method, completeUrl, protocol)

	for headerName, headerValue := range request.RawRequest.Header {
		fmt.Printf("%s: %s\n", styles.HeaderName.Render(headerName), strings.Join(headerValue, ";"))
	}

	fmt.Println(request.Template.JsonBody)
}

// LogResponse logs the response to the console.
func LogHttpResponse(ctx context.Context, httpResponse *models.HttpResponse) {
	fmt.Println(styles.SectionHeader.Render("Response"))

	protocol := styles.Url.Render(httpResponse.RawResponse.Proto)
	status := styles.Url.Render(httpResponse.RawResponse.Status)
	fmt.Println(protocol, status)

	// Headers
	if len(httpResponse.RawResponse.Header) > 0 {
		fmt.Println()
	}
	for key, value := range httpResponse.RawResponse.Header {
		fmt.Printf("%s: %s\n", styles.HeaderName.Render(key), strings.Join(value, "; "))
	}

	// Body
	fmt.Println("\n" + string(httpResponse.RawBody))

	fmt.Println(styles.SectionHeader.Render("Exports"))

	if len(httpResponse.Exports) == 0 {
		fmt.Println("  No exports")
	}
	for key, value := range httpResponse.Exports {
		key = styles.SecondaryText.Copy().Bold(true).Render(key)
		value = styles.PrimaryText.Render(fmt.Sprintf("%v", value))

		fmt.Printf("%s: %s\n", key, value)
	}
}
