package blocks

import (
	"bytes"
	"context"
	"fmt"

	"github.com/pkg/errors"
	ethpb "github.com/prysmaticlabs/ethereumapis/eth/v1alpha1"
	"github.com/prysmaticlabs/prysm/beacon-chain/core/helpers"
	iface "github.com/prysmaticlabs/prysm/beacon-chain/state/interface"
	"github.com/prysmaticlabs/prysm/shared/params"
)

// ProcessBlockHeader validates a block by its header.
//
// Spec pseudocode definition:
//
//  def process_block_header(state: BeaconState, block: BeaconBlock) -> None:
//    # Verify that the slots match
//    assert block.slot == state.slot
//    # Verify that the block is newer than latest block header
//    assert block.slot > state.latest_block_header.slot
//    # Verify that proposer index is the correct index
//    assert block.proposer_index == get_beacon_proposer_index(state)
//    # Verify that the parent matches
//    assert block.parent_root == hash_tree_root(state.latest_block_header)
//    # Cache current block as the new latest block
//    state.latest_block_header = BeaconBlockHeader(
//        slot=block.slot,
//        proposer_index=block.proposer_index,
//        parent_root=block.parent_root,
//        state_root=Bytes32(),  # Overwritten in the next process_slot call
//        body_root=hash_tree_root(block.body),
//    )
//
//    # Verify proposer is not slashed
//    proposer = state.validators[block.proposer_index]
//    assert not proposer.slashed
func ProcessBlockHeader(
	_ context.Context,
	beaconState iface.BeaconState,
	block *ethpb.SignedBeaconBlock,
) (iface.BeaconState, error) {
	beaconState, err := ProcessBlockHeaderNoVerify(beaconState, block.Block)
	if err != nil {
		return nil, err
	}

	// Verify proposer signature.
	if err := VerifyBlockSignature(beaconState, block); err != nil {
		return nil, err
	}

	return beaconState, nil
}

// ProcessBlockHeaderNoVerify validates a block by its header but skips proposer
// signature verification.
//
// WARNING: This method does not verify proposer signature. This is used for proposer to compute state root
// using a unsigned block.
//
// Spec pseudocode definition:
//  def process_block_header(state: BeaconState, block: BeaconBlock) -> None:
//    # Verify that the slots match
//    assert block.slot == state.slot
//    # Verify that the block is newer than latest block header
//    assert block.slot > state.latest_block_header.slot
//    # Verify that proposer index is the correct index
//    assert block.proposer_index == get_beacon_proposer_index(state)
//    # Verify that the parent matches
//    assert block.parent_root == hash_tree_root(state.latest_block_header)
//    # Cache current block as the new latest block
//    state.latest_block_header = BeaconBlockHeader(
//        slot=block.slot,
//        proposer_index=block.proposer_index,
//        parent_root=block.parent_root,
//        state_root=Bytes32(),  # Overwritten in the next process_slot call
//        body_root=hash_tree_root(block.body),
//    )
//
//    # Verify proposer is not slashed
//    proposer = state.validators[block.proposer_index]
//    assert not proposer.slashed
func ProcessBlockHeaderNoVerify(
	beaconState iface.BeaconState,
	block *ethpb.BeaconBlock,
) (iface.BeaconState, error) {
	if block == nil {
		return nil, errors.New("nil block")
	}
	if beaconState.Slot() != block.Slot {
		return nil, fmt.Errorf("state slot: %d is different than block slot: %d", beaconState.Slot(), block.Slot)
	}
	idx, err := helpers.BeaconProposerIndex(beaconState)
	if err != nil {
		return nil, err
	}
	if block.ProposerIndex != idx {
		return nil, fmt.Errorf("proposer index: %d is different than calculated: %d", block.ProposerIndex, idx)
	}
	parentHeader := beaconState.LatestBlockHeader()
	if parentHeader.Slot >= block.Slot {
		return nil, fmt.Errorf("block.Slot %d must be greater than state.LatestBlockHeader.Slot %d", block.Slot, parentHeader.Slot)
	}
	parentRoot, err := parentHeader.HashTreeRoot()
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(block.ParentRoot, parentRoot[:]) {
		return nil, fmt.Errorf(
			"parent root %#x does not match the latest block header signing root in state %#x",
			block.ParentRoot, parentRoot)
	}

	proposer, err := beaconState.ValidatorAtIndexReadOnly(idx)
	if err != nil {
		return nil, err
	}
	if proposer.Slashed() {
		return nil, fmt.Errorf("proposer at index %d was previously slashed", idx)
	}

	bodyRoot, err := block.Body.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	if err := beaconState.SetLatestBlockHeader(&ethpb.BeaconBlockHeader{
		Slot:          block.Slot,
		ProposerIndex: block.ProposerIndex,
		ParentRoot:    block.ParentRoot,
		StateRoot:     params.BeaconConfig().ZeroHash[:],
		BodyRoot:      bodyRoot[:],
	}); err != nil {
		return nil, err
	}
	return beaconState, nil
}
