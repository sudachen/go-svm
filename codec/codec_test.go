package codec

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCodec_EncodeCallData(t *testing.T) {
	c, err := NewCodec()
	require.NoError(t, err)

	rawCallData, err := c.EncodeCallData([]string{"i32"}, []int{-10})
	require.NoError(t, err)
	require.NotNil(t, rawCallData)
}

func TestCodec_DecodeDeployTemplateReceipt(t *testing.T) {
	c, err := NewCodec()
	require.NoError(t, err)

	rawReceipt, err := hex.DecodeString("0001bc213ffe5f285adf9b2df9975a98a8f3b8106bf7a02fda0000")
	require.NoError(t, err)

	receipt, err := c.DecodeDeployTemplateReceipt(rawReceipt)
	require.NoError(t, err)
	require.NotNil(t, receipt)
	require.Equal(t, "bc213ffe5f285adf9b2df9975a98a8f3b8106bf7", receipt.TemplateAddr.String())
}
