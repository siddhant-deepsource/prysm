package registration

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	types "github.com/prysmaticlabs/eth2-types"
	"github.com/prysmaticlabs/prysm/cmd/beacon-chain/flags"
	"github.com/urfave/cli/v2"
)

// BlockchainPreregistration prepares data for blockchain.Service's registration.
func BlockchainPreregistration(cliCtx *cli.Context) (bRoot []byte, epoch types.Epoch, err error) {
	wsp := cliCtx.String(flags.WeakSubjectivityCheckpt.Name)
	bRoot, epoch, err = convertWspInput(wsp)
	if err != nil {
		return nil, 0, err
	}

	return
}

// Given input string `block_root:epoch_number`, this verifies the input string is valid, and
// returns the block root as bytes and epoch number as unsigned integers.
func convertWspInput(wsp string) ([]byte, types.Epoch, error) {
	if wsp == "" {
		return nil, 0, nil
	}

	// Weak subjectivity input string must contain ":" to separate epoch and block root.
	if !strings.Contains(wsp, ":") {
		return nil, 0, fmt.Errorf("%s did not contain column", wsp)
	}

	// Strip prefix "0x" if it's part of the input string.
	wsp = strings.TrimPrefix(wsp, "0x")

	// Get the hexadecimal block root from input string.
	s := strings.Split(wsp, ":")
	if len(s) != 2 {
		return nil, 0, errors.New("weak subjectivity checkpoint input should be in `block_root:epoch_number` format")
	}

	bRoot, err := hex.DecodeString(s[0])
	if err != nil {
		return nil, 0, err
	}
	if len(bRoot) != 32 {
		return nil, 0, errors.New("block root is not length of 32")
	}

	// Get the epoch number from input string.
	epoch, err := strconv.ParseUint(s[1], 10, 64)
	if err != nil {
		return nil, 0, err
	}

	return bRoot, types.Epoch(epoch), nil
}