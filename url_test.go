package wc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUrlParse(t *testing.T) {
	u, err := ParseURL("wc:8a5e5bdc-a0e4-4702-ba63-8f1a5655744f@1?bridge=https%3A%2F%2Fbridge.walletconnect.org&key=41791102999c339c844880b23950704cc43aa840f3739e365323cda4dfa89e7a")

	require.NoError(t, err)

	println("topic: ", u.Topic)
	println("version: ", u.Version)
	println("bridge: ", u.Bridge)
	println("key: ", u.Key)

	println(fmt.Sprintf("%s", u))
}
