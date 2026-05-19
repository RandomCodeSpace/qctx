package gitlab_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/RandomCodeSpace/qctx/internal/gitlab"
)

func TestParseMRURL(t *testing.T) {
	cases := []struct {
		in    string
		host  string
		path  string
		iid   int
		errOk bool
	}{
		{
			in:   "https://gitlab.example.com/team/my-svc/-/merge_requests/42",
			host: "https://gitlab.example.com", path: "team/my-svc", iid: 42,
		},
		{
			in:   "https://gitlab.example.com/group/sub/proj/-/merge_requests/7?diff=true",
			host: "https://gitlab.example.com", path: "group/sub/proj", iid: 7,
		},
		{
			in:   "https://gitlab.example.com:8443/x/y/-/merge_requests/1",
			host: "https://gitlab.example.com:8443", path: "x/y", iid: 1,
		},
		{in: "not a url", errOk: true},
		{in: "https://gitlab.example.com/x/y/issues/1", errOk: true},
		{in: "https://gitlab.example.com/x/y/-/merge_requests/abc", errOk: true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := gitlab.ParseMRURL(tc.in)
			if tc.errOk {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.host, got.Host)
			require.Equal(t, tc.path, got.ProjectPath)
			require.Equal(t, tc.iid, got.IID)
		})
	}
}
