package nexproto

import (
	"fmt"

	nex "github.com/PretendoNetwork/nex-go"
)

const (
	// ShopBadgeArcadeProtocolID is the Protocol ID for the Shop (Badge Arcade) protocol
	ShopBadgeArcadeProtocolID = 0x7F

	// ShopBadgeArcadeCustomID is the Custom ID for the Shop (Badge Arcade) protocol
	ShopBadgeArcadeCustomID = 0xC8

	// ShopBadgeArcadeMethodGetRivToken is the method ID for GetRivToken
	ShopBadgeArcadeMethodGetRivToken = 0x1

	// ShopBadgeArcadeMethodPostPlayLog is the method ID for PostPlayLog
	ShopBadgeArcadeMethodPostPlayLog = 0x2
)

// ShopBadgeArcadeProtocol handles the Shop (Badge Arcade) nex protocol
type ShopBadgeArcadeProtocol struct {
	server *nex.Server
	PostPlayLogHandler func(err error, client *nex.Client, callID uint32, unknown1 uint16, param *ShopPostPlayLogParam)
}

type ShopPostPlayLogParam struct {
	nex.Structure
	Unknown1  []uint32
	Timestamp *nex.DateTime
	Unknown2  string
}

// ExtractFromStream extracts a ShopPostPlayLogParam structure from a stream
func (shopPostPlayLogParam *ShopPostPlayLogParam) ExtractFromStream(stream *nex.StreamIn) error {
	shopPostPlayLogParam.Unknown1 = stream.ReadListUInt32LE()
	shopPostPlayLogParam.Timestamp = stream.ReadDateTime()
	shopPostPlayLogParam.Unknown2, _ = stream.ReadString()
	
	return nil
}

// Bytes encodes the ShopPostPlayLogParam and returns a byte array
func (shopPostPlayLogParam *ShopPostPlayLogParam) Bytes(stream *nex.StreamOut) []byte {
	stream.WriteListUInt32LE(shopPostPlayLogParam.Unknown1)
	stream.WriteUInt64LE(shopPostPlayLogParam.Timestamp.Value())
	stream.WriteString(shopPostPlayLogParam.Unknown2)

	return stream.Bytes()
}

// NewShopPostPlayLogParam returns a new ShopPostPlayLogParam
func NewShopPostPlayLogParam() *ShopPostPlayLogParam {
	return &ShopPostPlayLogParam{}
}

// Setup initializes the protocol
func (shopBadgeArcadeProtocol *ShopBadgeArcadeProtocol) Setup() {
	nexServer := shopBadgeArcadeProtocol.server

	nexServer.On("Data", func(packet nex.PacketInterface) {
		request := packet.RMCRequest()

		if ShopBadgeArcadeProtocolID == request.ProtocolID() && ShopBadgeArcadeCustomID == request.CustomID() {
			switch request.MethodID() {
			case ShopBadgeArcadeMethodPostPlayLog:
				go shopBadgeArcadeProtocol.handlePostPlayLog(packet)
			default:
				go respondNotImplementedCustom(packet, ShopBadgeArcadeCustomID)
				fmt.Printf("Unsupported ShopBadgeArcade method ID: %#v\n", request.MethodID())
			}
		}
	})
}

// PostPlayLog sets the PostPlayLog function
func (shopBadgeArcadeProtocol *ShopBadgeArcadeProtocol) PostPlayLog(handler func(err error, client *nex.Client, callID uint32, unknown1 uint16, param *ShopPostPlayLogParam)) {
	shopBadgeArcadeProtocol.PostPlayLogHandler = handler
}

func (shopBadgeArcadeProtocol *ShopBadgeArcadeProtocol) handlePostPlayLog(packet nex.PacketInterface) {
	if shopBadgeArcadeProtocol.PostPlayLogHandler == nil {
		logger.Warning("ShopBadgeArcadeProtocol::PostPlayLog not implemented")
		go respondNotImplementedCustom(packet, ShopBadgeArcadeCustomID)
		return
	}

	client := packet.Sender()
	request := packet.RMCRequest()

	callID := request.CallID()
	parameters := request.Parameters()

	parametersStream := nex.NewStreamIn(parameters, shopBadgeArcadeProtocol.server)

	unknown1 := parametersStream.ReadUInt16LE()

	param, err := parametersStream.ReadStructure(NewShopPostPlayLogParam())
	if err != nil {
		go shopBadgeArcadeProtocol.PostPlayLogHandler(err, client, callID, 0, nil)
		return
	}

	go shopBadgeArcadeProtocol.PostPlayLogHandler(nil, client, callID, unknown1, param.(*ShopPostPlayLogParam))
}

// NewShopBadgeArcadeProtocol returns a new ShopBadgeArcadeProtocol
func NewShopBadgeArcadeProtocol(server *nex.Server) *ShopBadgeArcadeProtocol {
	shopBadgeArcadeProtocol := &ShopBadgeArcadeProtocol{server: server}

	shopBadgeArcadeProtocol.Setup()

	return shopBadgeArcadeProtocol
}
