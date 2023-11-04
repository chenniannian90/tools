package openapi2word

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenApi_GenerateDoc(t *testing.T) {
	gen := NewGenerateOpenAPIDoc("user-center", &url.URL{
		Scheme: "http",
		Path:   "srv-user-center.base.d.rktl.work/user-center",
	}, 3)
	gen.Load()
	err := gen.GenerateClientOpenAPIDoc("./user-center.docx")
	require.NoError(t, err)
}
