package types

import (
	"errors"
	"strconv"
	"strings"
)

// Commitment should contain serialized cryptographic commitment to Blob value.
type Commitment = []byte

// Blob is the data submitted/received from DA interface.
type Blob = []byte

// Client defines all the possible da clients
type Client string

// DAMetaData contains meta data about a batch on the Data Availability Layer.
type DASubmitMetaData struct {
	// Height is the height of the block in the da layer
	height uint64
	// Namespace ID
	namespace []byte
	// Client is the client to use to fetch data from the da layer
	client Client
	//Share commitment, for each blob, used to obtain blobs and proofs
	commitment Commitment
	//Initial position for each blob in the NMT
	index int
	//Number of shares of each blob
	length int
	//any NMT root for the specific height, necessary for non-inclusion proof
	root []byte
}

func NewDAMetaData(DAPath string) (*DASubmitMetaData, error) {
	pathParts := strings.FieldsFunc(DAPath, func(r rune) bool { return r == '.' })

	if len(pathParts) != 7 {
		return nil, errors.New("unable to decode da path")
	}
	height, err := strconv.ParseUint(pathParts[1], 10, 64)
	if err != nil {
		return nil, err
	}

	index, err := strconv.Atoi(pathParts[2])
	if err != nil {
		return nil, err
	}
	length, err := strconv.Atoi(pathParts[3])
	if err != nil {
		return nil, err
	}
	return &DASubmitMetaData{
		client:     Client(pathParts[0]),
		height:     height,
		namespace:  []byte(pathParts[5]),
		index:      index,
		length:     length,
		commitment: []byte(pathParts[4]),
		root:       []byte(pathParts[6]),
	}, nil

}

func (d *DASubmitMetaData) GetClient() Client {
	return d.client
}

func (d *DASubmitMetaData) GetNameSpace() []byte {
	return d.namespace
}

func (d *DASubmitMetaData) GetCommitment() []byte {
	return d.commitment
}

func (d *DASubmitMetaData) GetDataRoot() []byte {
	return d.root
}

func (d *DASubmitMetaData) GetIndex() int {
	return d.index
}
func (d *DASubmitMetaData) GetLength() int {
	return d.length
}
func (d *DASubmitMetaData) GetDAHeight() uint64 {
	return d.height
}
