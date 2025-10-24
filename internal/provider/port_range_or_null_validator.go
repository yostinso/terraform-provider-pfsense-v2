package provider

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// PortRangeOrNullValidator validates that a string is either "null" or a number in the range 1-65535.

type PortRangeOrNullValidator struct{}

func (v PortRangeOrNullValidator) Description(ctx context.Context) string {
	return "Set to `null` to allow any destination port.<br>Other valid options are: a TCP/UDP port number, a TCP/UDP port range separated by `:`."
}

func (v PortRangeOrNullValidator) MarkdownDescription(ctx context.Context) string {
	return "Set to `null` to allow any destination port.<br>Other valid options are: a TCP/UDP port number, a TCP/UDP port range separated by `:`."
}

func (v PortRangeOrNullValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	val := req.ConfigValue.ValueString()
	if val == "null" {
		return
	}

	portRangeRe := regexp.MustCompile(`^(\d+):(\d+)$`)
	portRangeMatch := portRangeRe.FindStringSubmatch(val)
	// Single port number
	if portRangeMatch != nil {
		_, err := PortNumber(portRangeMatch[1])
		if err != nil {
			resp.Diagnostics.Append(
				diag.NewAttributeErrorDiagnostic(
					req.Path,
					fmt.Sprintf("Invalid port number in range: %s", portRangeMatch[1]),
					"Value must be a number between 1 and 65535.",
				),
			)
			return
		}
		_, err = PortNumber(portRangeMatch[2])
		if err != nil {
			resp.Diagnostics.Append(
				diag.NewAttributeErrorDiagnostic(
					req.Path,
					fmt.Sprintf("Invalid port number in range: %s", portRangeMatch[2]),
					"Value must be a number between 1 and 65535.",
				),
			)
			return
		}
		return
	}

	// Single port number
	_, err := strconv.Atoi(val)
	if err != nil {
		resp.Diagnostics.Append(
			diag.NewAttributeErrorDiagnostic(
				req.Path,
				"Invalid port value",
				err.Error(),
			),
		)
	}
}

func PortNumber(val string) (int, error) {
	n, err := strconv.Atoi(val)
	if err != nil || n < 1 || n > 65535 {
		return 0, fmt.Errorf("value must `null` or a number between 1 and 65535")
	}
	return n, nil
}
