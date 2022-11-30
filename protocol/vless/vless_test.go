package vless

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/v2fly/v2ray-core/v4/common"
	"github.com/v2fly/v2ray-core/v4/common/buf"
	"github.com/v2fly/v2ray-core/v4/common/net"
	"github.com/v2fly/v2ray-core/v4/common/protocol"
	"github.com/v2fly/v2ray-core/v4/proxy/vless"
	"github.com/v2fly/v2ray-core/v4/proxy/vless/encoding"
)

func toAccount(a *vless.Account) protocol.Account {
	account, err := a.AsAccount()
	common.Must(err)
	return account
}

func TestVLESSRequest2(t *testing.T) {
	user := &protocol.MemoryUser{
		Level: 0,
		Email: "test@v2fly.org",
		Account: toAccount(&vless.Account{
			Id: uuid.New().String(),
		}),
	}
	account := user.Account.(*vless.MemoryAccount)

	expReq := &protocol.RequestHeader{
		Version: Version,
		User:    user,
		Command: protocol.RequestCommandTCP,
		Address: net.IPAddress([]byte{10, 0, 5, 14}),
		Port:    net.Port(443),
	}
	expectedAddons := &encoding.Addons{}

	buffer := buf.StackNew()
	common.Must(encoding.EncodeRequestHeader(&buffer, expReq, expectedAddons))

	aclReq := requestPool.Get().(*Request)
	err := aclReq.FromReader(&buffer)
	require.NoError(t, err)
	require.Equal(t,
		fmt.Sprintf("%s:%s", expReq.Address.String(), expReq.Port),
		aclReq.DestAddr())
	require.Equal(t, byte(expReq.Command), aclReq.Command())
	require.True(t, account.ID.UUID() == aclReq.UUID())
}
