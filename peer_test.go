package wc

import (
	"encoding/json"
	"testing"

	"github.com/libs4go/scf4go"
	_ "github.com/libs4go/scf4go/codec" //
	"github.com/libs4go/scf4go/reader/file"
	"github.com/libs4go/slf4go"
	_ "github.com/libs4go/slf4go/backend/console"
	"github.com/stretchr/testify/require"
)

var url = "wc:da6548c7-d7d7-479d-97d6-a57c9a0ac4e7@1?bridge=https%3A%2F%2Fbridge.walletconnect.org&key=3ff7d84fa6ad834d8f8a1793294e2821200fb4e24328a4c93b0168f044a34140"

func init() {
	config := scf4go.New()

	err := config.Load(file.New(file.Yaml("./slf4go.yaml")))

	if err != nil {
		panic(err)
	}

	err = slf4go.Config(config)

	if err != nil {
		panic(err)
	}
}

func TestSubcribeTopic(t *testing.T) {

	defer slf4go.Sync()

	peer, err := NewServer(url, 1, []string{"0x2D8e5b082dFA5cD2A8EcFA5A0a93956cAD3dF91A"}, &PeerInfo{
		Description: "Test",
		Name:        "TestWallet",
	})

	require.NoError(t, err)

	err = peer.Handshake()

	require.NoError(t, err)

	printJSON(peer.session)

	err = peer.HandshakeApprove(true)

	require.NoError(t, err)

	buff, err := peer.ReadMessage()

	require.NoError(t, err)

	println(string(buff))

}

func printJSON(v interface{}) {
	buff, _ := json.MarshalIndent(v, "\t", "")

	println(buff)
}

var j = `
{"id":1612446165093546,"jsonrpc":"2.0","method":"wc_sessionRequest","params":[{"peerId":"babd2a3b-0f6d-4bea-89c7-0ea33283513c","peerMeta":{"description":"","url":"https://snapshot.page","icons":["https://snapshot.page/favicon.png"],"name":"FinNexus"},"chainId":1}]}
`

func TestUnmarshal(t *testing.T) {
	var request *jsonRPCRequest

	err := json.Unmarshal([]byte(j), &request)

	require.NoError(t, err)
}

var d = `
{"topic":"037bb6f3-d42a-4652-ba9b-99156c3603e5","type":"pub","payload":"{\"data\":\"18662d9147a44b0ed60c04eea9e011f67afdba38091af493c27d7285cb9a6b365c7e6ab9060d52f6df3c6db8a5f4832c6d8e8054d49b3e00ef0b17ae85356e576f5d321ae18c7ce5a5adbe388cb85ea8c673e3613a817e7f9c9ad0c7119f178759e80ab1f062a480c05337ea4fb583893581974d534b42f15d47343cf2662e6bfe8cf04f098b52f17fad36a29a2234a4ea4ed4ec4c69e7438442f0dd9af5415b93804247a3d0eb82453478145fd5fc9153663ded5e039e6cea297c3e9e77440cad0bc3eebc3cc25875e071630fe5b68490855dec96d9f91a689fdf2c71a9b1bfb9513aaf59ba1f6a16ad4eb1fc93191178fbe2bb657a5b7117f977edd33a3e5f8df80339e6d4c6ec8792bbf38731ac0290b344ce46ea4832f5a2ffb448ad5c911cb690bdf333ae32f54e978e5c22560408e62af67baa78cb4101ae63dd53f708da8b8c6fbfe6baf0abcf31165f556032052edaf51fc75cf3cd59fbeeebec4b31c5ac763fde54625641f8d8ecb13880a666d469b1c0cb660ac44e122aa87701603438c7e8261a11cfe72798e58e0db449\",\"hmac\":\"136d5e0e8951f5efaa3325d099a00c54829e0c56aa03c18ca8f8c7e059cc2c24\",\"iv\":\"846a00773dad95c1b425bcefaaab5912\"}","silent":false}
`

func TestDecrypt(t *testing.T) {
	s, err := NewServerSession(url)

	require.NoError(t, err)

	buff, err := s.HandleSubscribe([]byte(d))

	require.NoError(t, err)

	println(string(buff))
}
