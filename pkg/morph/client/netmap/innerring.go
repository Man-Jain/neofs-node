package netmap

import (
	"crypto/elliptic"
	"fmt"

	"github.com/nspcc-dev/neo-go/pkg/crypto/keys"
	"github.com/nspcc-dev/neo-go/pkg/vm/stackitem"
	"github.com/nspcc-dev/neofs-node/pkg/morph/client"
)

// UpdateInnerRing updates inner ring members in netmap contract.
func (c *Client) UpdateInnerRing(keys keys.PublicKeys) error {
	args := make([][]byte, len(keys))
	for i := range args {
		args[i] = keys[i].Bytes()
	}

	return c.client.Invoke(c.updateInnerRing, args)
}

// InnerRingList returns public keys of inner ring members in
// netmap contract.
func (c *Client) InnerRingList() (keys.PublicKeys, error) {
	prms, err := c.client.TestInvoke(
		c.innerRingList,
	)
	if err != nil {
		return nil, fmt.Errorf("could not perform test invocation (%s): %w", c.innerRingList, err)
	}

	return irKeysFromStackItem(prms, c.innerRingList)
}

func irKeysFromStackItem(stack []stackitem.Item, method string) (keys.PublicKeys, error) {
	if ln := len(stack); ln != 1 {
		return nil, fmt.Errorf("unexpected stack item count (%s): %d", method, ln)
	}

	irs, err := client.ArrayFromStackItem(stack[0])
	if err != nil {
		return nil, fmt.Errorf("could not get stack item array from stack item (%s): %w", method, err)
	}

	irKeys := make(keys.PublicKeys, len(irs))

	for i := range irs {
		irKeys[i], err = irKeyFromStackItem(irs[i])
		if err != nil {
			return nil, err
		}
	}

	return irKeys, nil
}

const irNodeFixedPrmNumber = 1

func irKeyFromStackItem(prm stackitem.Item) (*keys.PublicKey, error) {
	prms, err := client.ArrayFromStackItem(prm)
	if err != nil {
		return nil, fmt.Errorf("could not get stack item array (IRNode): %w", err)
	} else if ln := len(prms); ln != irNodeFixedPrmNumber {
		return nil, fmt.Errorf(
			"unexpected stack item count (IRNode): expected %d, has %d",
			irNodeFixedPrmNumber,
			ln,
		)
	}

	byteKey, err := client.BytesFromStackItem(prms[0])
	if err != nil {
		return nil, fmt.Errorf("could not parse bytes from stack item (IRNode): %w", err)
	}

	return keys.NewPublicKeyFromBytes(byteKey, elliptic.P256())
}
