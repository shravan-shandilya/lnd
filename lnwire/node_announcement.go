package lnwire

import (
	"bytes"
	"io"
	"net"

	"github.com/go-errors/errors"
	"github.com/roasbeef/btcd/btcec"
)

var (
	startPort    uint16 = 1024
	endPort      uint16 = 49151
	aliasSpecLen        = 21
)

// RGB is used to represent the color.
type RGB struct {
	red   uint8
	green uint8
	blue  uint8
}

// Alias a hex encoded UTF-8 string that may be displayed as an alternative to
// the node's ID. Notice that aliases are not unique and may be freely chosen
// by the node operators.
type Alias struct {
	data     [32]byte
	aliasLen int
}

// NewAlias create the alias from string and also checks spec requirements.
func NewAlias(s string) (Alias, error) {
	data := []byte(s)
	return newAlias(data)
}

func newAlias(data []byte) (Alias, error) {
	var a [32]byte

	aliasEnd := len(data)
	for data[aliasEnd-1] == 0 && aliasEnd > 0 {
		aliasEnd--
	}

	copy(a[:aliasEnd], data[:aliasEnd])

	return Alias{
		data:     a,
		aliasLen: aliasEnd,
	}, nil
}

func (a *Alias) String() string {
	return string(a.data[:a.aliasLen])
}

// Validate check that alias data length is lower than spec size.
func (a *Alias) Validate() error {
	nonzero := len(a.data)
	for a.data[nonzero-1] == 0 && nonzero > 0 {
		nonzero--
	}

	if nonzero > aliasSpecLen {
		return errors.New("alias should be less then 21 bytes")
	}

	return nil
}

// NodeAnnouncement message is used to announce the presence of a lightning node
// and signal that the node is accepting incoming connections.
type NodeAnnouncement struct {
	// Signature is used to prove the ownership of node id.
	Signature *btcec.Signature

	// Timestamp allows ordering in the case of multiple announcements.
	Timestamp uint32

	// Address includes two specification fields: 'ipv6' and 'port' on which
	// the node is accepting incoming connections.
	Address *net.TCPAddr

	// NodeID is a public key which is used as node identification.
	NodeID *btcec.PublicKey

	// RGBColor is used to customize their node's appearance in maps and graphs
	RGBColor RGB

	// pad is used to reserve to additional bytes for future usage.
	pad uint16

	// Alias is used to customize their node's appearance in maps and graphs
	Alias Alias
}

// A compile time check to ensure NodeAnnouncement implements the
// lnwire.Message interface.
var _ Message = (*NodeAnnouncement)(nil)

// Validate performs any necessary sanity checks to ensure all fields present
// on the NodeAnnouncement are valid.
//
// This is part of the lnwire.Message interface.
func (a *NodeAnnouncement) Validate() error {
	// TODO(roasbeef): move validation to discovery service
	return nil
}

// Decode deserializes a serialized NodeAnnouncement stored in the
// passed io.Reader observing the specified protocol version.
//
// This is part of the lnwire.Message interface.
func (c *NodeAnnouncement) Decode(r io.Reader, pver uint32) error {
	err := readElements(r,
		&c.Signature,
		&c.Timestamp,
		&c.Address,
		&c.NodeID,
		&c.RGBColor,
		&c.pad,
		&c.Alias,
	)
	if err != nil {
		return err
	}

	return nil
}

// Encode serializes the target NodeAnnouncement into the passed
// io.Writer observing the protocol version specified.
//
// This is part of the lnwire.Message interface.
func (c *NodeAnnouncement) Encode(w io.Writer, pver uint32) error {
	err := writeElements(w,
		c.Signature,
		c.Timestamp,
		c.Address,
		c.NodeID,
		c.RGBColor,
		c.pad,
		c.Alias,
	)
	if err != nil {
		return err
	}

	return nil
}

// Command returns the integer uniquely identifying this message type on the
// wire.
//
// This is part of the lnwire.Message interface.
func (c *NodeAnnouncement) Command() uint32 {
	return CmdNodeAnnoucmentMessage
}

// MaxPayloadLength returns the maximum allowed payload size for this message
// observing the specified protocol version.
//
// This is part of the lnwire.Message interface.
func (c *NodeAnnouncement) MaxPayloadLength(pver uint32) uint32 {
	var length uint32

	// Signature - 64 bytes
	length += 64

	// Timestamp - 4 bytes
	length += 4

	// Ipv6 - 16 bytes
	length += 16

	// Port - 4 bytes
	length += 4

	// NodeID - 33 bytes
	length += 33

	// RGBColor - 3 bytes
	length += 3

	// pad - 2 bytes
	length += 2

	// Alias - 32 bytes
	length += 32

	// 158
	return length
}

// dataToSign...
func (c *NodeAnnouncement) DataToSign() ([]byte, error) {

	// We should not include the signatures itself.
	var w bytes.Buffer
	err := writeElements(&w,
		c.Timestamp,
		c.Address,
		c.NodeID,
		c.RGBColor,
		c.pad,
		c.Alias,
	)
	if err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}
