package ptaggr

import (
	"fmt"
	"strings"
	"testing"

	"github.com/coredns/caddy"
)

type setupTestCase struct {
	config        string
	expectedError string
}

func TestSetupAlternate(t *testing.T) {
	testCases := []setupTestCase{
		{
			config: `ptaggr . 192.168.1.1:53 {
						max_fails 5
						force_tcp
					}`,
			expectedError: `additional parameters not allowed`,
		},
		{
			config:        `ptaggr . tls://192.168.1.1:443`,
			expectedError: `only dns transport allowed`,
		},
		{
			config:        `ptaggr . abc`,
			expectedError: `not an IP address or file`,
		},
		{
			config: `ptaggr . 192.168.1.1:53`,
		},
		{
			config: `ptaggr original . 192.168.1.1:53`,
		},
		{
			config: `ptaggr original private . 192.168.1.1:53`,
		},
		{
			config: `ptaggr private . 192.168.1.1:53`,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s", tc.config), func(t *testing.T) {
			c := caddy.NewTestController("dns", tc.config)
			err := setup(c)
			if err == nil {
				if tc.expectedError != "" {
					t.Errorf("Expected error '%s', but got no error", tc.expectedError)
				}
			} else {
				if tc.expectedError == "" {
					t.Errorf("Expected no error, but got '%s'", err)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error '%s', but got '%s'", tc.expectedError, err)
				}
			}
		})
	}
}
