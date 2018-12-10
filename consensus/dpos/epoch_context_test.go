package dpos

import (
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/allsportschain/go-allsportschain/common"
	"github.com/allsportschain/go-allsportschain/core/state"
	"github.com/allsportschain/go-allsportschain/core/types"
	"github.com/allsportschain/go-allsportschain/socdb"
	"github.com/allsportschain/go-allsportschain/trie"

	"github.com/stretchr/testify/assert"
		"github.com/allsportschain/go-allsportschain/params"
)

func TestEpochContextCountVotes(t *testing.T) {
	config := &params.ChainConfig{
		MultiVoteBlock:      big.NewInt(100),
	}
	header := &types.Header{
		Number: big.NewInt(10),
	}

	voteMap := map[common.Address][]common.Address{
		common.HexToAddress("0x44d1ce0b7cb3588bca96151fe1bc05af38f91b6e"): {
			common.HexToAddress("0xb040353ec0f2c113d5639444f7253681aecda1f8"),
		},
		common.HexToAddress("0xa60a3886b552ff9992cfcd208ec1152079e046c2"): {
			common.HexToAddress("0x14432e15f21237013017fa6ee90fc99433dec82c"),
			common.HexToAddress("0x9f30d0e5c9c88cade54cd1adecf6bc2c7e0e5af6"),
		},
		common.HexToAddress("0x4e080e49f62694554871e669aeb4ebe17c4a9670"): {
			common.HexToAddress("0xd83b44a3719720ec54cdb9f54c0202de68f1ebcb"),
			common.HexToAddress("0x56cc452e450551b7b9cffe25084a069e8c1e9441"),
			common.HexToAddress("0xbcfcb3fa8250be4f2bf2b1e70e1da500c668377b"),
		},
		common.HexToAddress("0x9d9667c71bb09d6ca7c3ed12bfe5e7be24e2ffe1"): {},
	}
	balance := int64(5)
	db := socdb.NewMemDatabase()
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(db))
	dposContext, err := types.NewDposContext(db)
	assert.Nil(t, err)

	epochContext := &EpochContext{
		DposContext: dposContext,
		statedb:     stateDB,
	}
	_, err = epochContext.countVotes()
	assert.NotNil(t, err)

	for candidate, electors := range voteMap {
		assert.Nil(t, dposContext.BecomeCandidate(config,header,candidate))
		for _, elector := range electors {
			stateDB.SetBalance(elector, big.NewInt(balance))
			assert.Nil(t, dposContext.Delegate(config, header, elector, candidate))
		}
	}
	result, err := epochContext.countVotes()
	assert.Nil(t, err)
	assert.Equal(t, len(voteMap), len(result))
	for candidate, electors := range voteMap {
		voteCount, ok := result[candidate]
		assert.True(t, ok)
		assert.Equal(t, balance*int64(len(electors)), voteCount.Int64())
	}
}

func TestEpochContextCountVotesForMultiVote(t *testing.T) {
	config := &params.ChainConfig{
		MultiVoteBlock:      big.NewInt(0),
	}
	header := &types.Header{
		Number: big.NewInt(10),
	}

	voteMap := map[common.Address][]common.Address{
		common.HexToAddress("0x44d1ce0b7cb3588bca96151fe1bc05af38f91b6e"): {
			common.HexToAddress("0xb040353ec0f2c113d5639444f7253681aecda1f8"),
		},
		common.HexToAddress("0xa60a3886b552ff9992cfcd208ec1152079e046c2"): {
			common.HexToAddress("0xb040353ec0f2c113d5639444f7253681aecda1f8"),
			common.HexToAddress("0x9f30d0e5c9c88cade54cd1adecf6bc2c7e0e5af6"),
		},
		common.HexToAddress("0x4e080e49f62694554871e669aeb4ebe17c4a9670"): {
			common.HexToAddress("0xb040353ec0f2c113d5639444f7253681aecda1f8"),
			common.HexToAddress("0x56cc452e450551b7b9cffe25084a069e8c1e9441"),
			common.HexToAddress("0xbcfcb3fa8250be4f2bf2b1e70e1da500c668377b"),
		},
		common.HexToAddress("0x9d9667c71bb09d6ca7c3ed12bfe5e7be24e2ffe1"): {},
	}
	balance := int64(5)
	db := socdb.NewMemDatabase()
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(db))
	dposContext, err := types.NewDposContext(db)
	assert.Nil(t, err)

	epochContext := &EpochContext{
		DposContext: dposContext,
		statedb:     stateDB,
	}
	_, err = epochContext.countVotes()
	assert.NotNil(t, err)

	mapElector := map[common.Address]bool{}
	for candidate, electors := range voteMap {
		assert.Nil(t, dposContext.BecomeCandidate(config,header,candidate))
		for _, elector := range electors {
			if mapElector[elector] == false {
				stateDB.SetBalance(elector, big.NewInt(balance))
				mapElector[elector] = true
			}
			assert.Nil(t, dposContext.Delegate(config, header, elector, candidate))
		}
	}
	result, err := epochContext.countVotes()
	assert.Nil(t, err)
	assert.Equal(t, len(voteMap), len(result))
	for candidate, electors := range voteMap {
		voteCount, ok := result[candidate]
		assert.True(t, ok)
		assert.Equal(t, balance*int64(len(electors)), voteCount.Int64())
	}
}

func TestLookupValidator(t *testing.T) {
	db := socdb.NewMemDatabase()
	dposCtx, _ := types.NewDposContext(db)
	mockEpochContext := &EpochContext{
		DposContext: dposCtx,
	}
	validators := []common.Address{
		common.BytesToAddress([]byte("addr1")),
		common.BytesToAddress([]byte("addr2")),
		common.BytesToAddress([]byte("addr3")),
	}
	mockEpochContext.DposContext.SetValidators(validators)
	for i, expected := range validators {
		got, _ := mockEpochContext.lookupValidator(int64(i) * blockInterval)
		if got != expected {
			t.Errorf("Failed to test lookup validator, %s was expected but got %s", string(expected[:]), string(got[:]))
		}
	}
	_, err := mockEpochContext.lookupValidator(blockInterval - 1)
	if err != ErrInvalidMintBlockTime {
		t.Errorf("Failed to test lookup validator. err '%v' was expected but got '%v'", ErrInvalidMintBlockTime, err)
	}
}

func TestEpochContextKickoutValidator(t *testing.T) {
	// for not multi vote
	config := &params.ChainConfig{
		MultiVoteBlock: big.NewInt(100),
	}
	header := &types.Header{
		Number: big.NewInt(99),
	}

	checkEpochContextKickoutValidator(t,config,header)

	// for multi vote
	config = &params.ChainConfig{
		MultiVoteBlock: big.NewInt(0),
	}
	header = &types.Header{
		Number: big.NewInt(100),
	}

	checkEpochContextKickoutValidator(t,config,header)
}
func checkEpochContextKickoutValidator(t *testing.T, config *params.ChainConfig, header *types.Header) {
	db := socdb.NewMemDatabase()
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(db))
	dposContext, err := types.NewDposContext(db)
	assert.Nil(t, err)
	epochContext := &EpochContext{
		TimeStamp:   epochInterval,
		DposContext: dposContext,
		statedb:     stateDB,
	}
	atLeastMintCnt := epochInterval / blockInterval / maxValidatorSize / 2
	testEpoch := int64(1)

	// no validator can be kickout, because all validators mint enough block at least
	validators := []common.Address{}
	for i := 0; i < maxValidatorSize; i++ {
		validator := common.BytesToAddress([]byte("addr" + strconv.Itoa(i)))
		validators = append(validators, validator)
		assert.Nil(t, dposContext.BecomeCandidate(config, header, validator))
		setTestMintCnt(dposContext, testEpoch, validator, atLeastMintCnt)
	}
	assert.Nil(t, dposContext.SetValidators(validators))
	assert.Nil(t, dposContext.BecomeCandidate(config, header, common.BytesToAddress([]byte("addr"))))
	assert.Nil(t, epochContext.kickoutValidator(config, header, testEpoch))
	candidateMap := getCandidates(dposContext.CandidateTrie())
	assert.Equal(t, maxValidatorSize +1, len(candidateMap))

	// atLeast a safeSize count candidate will reserve
	dposContext, err = types.NewDposContext(db)
	assert.Nil(t, err)
	epochContext = &EpochContext{
		TimeStamp:   epochInterval,
		DposContext: dposContext,
		statedb:     stateDB,
	}
	validators = []common.Address{}
	for i := 0; i < maxValidatorSize; i++ {
		validator := common.BytesToAddress([]byte("addr" + strconv.Itoa(i)))
		validators = append(validators, validator)
		assert.Nil(t, dposContext.BecomeCandidate(config, header, validator))
		setTestMintCnt(dposContext, testEpoch, validator, atLeastMintCnt-int64(i)-1)
	}
	assert.Nil(t, dposContext.SetValidators(validators))
	assert.Nil(t, epochContext.kickoutValidator(config, header, testEpoch))
	candidateMap = getCandidates(dposContext.CandidateTrie())
	assert.Equal(t, safeSize, len(candidateMap))
	for i := maxValidatorSize - 1; i >= safeSize; i-- {
		assert.False(t, candidateMap[common.BytesToAddress([]byte("addr"+strconv.Itoa(i)))])
	}

	// all validator will be kickout, because all validators didn't mint enough block at least
	dposContext, err = types.NewDposContext(db)
	assert.Nil(t, err)
	epochContext = &EpochContext{
		TimeStamp:   epochInterval,
		DposContext: dposContext,
		statedb:     stateDB,
	}
	validators = []common.Address{}
	for i := 0; i < maxValidatorSize; i++ {
		validator := common.BytesToAddress([]byte("addr" + strconv.Itoa(i)))
		validators = append(validators, validator)
		assert.Nil(t, dposContext.BecomeCandidate(config, header, validator))
		setTestMintCnt(dposContext, testEpoch, validator, atLeastMintCnt-1)
	}
	for i := maxValidatorSize; i < maxValidatorSize *2; i++ {
		candidate := common.BytesToAddress([]byte("addr" + strconv.Itoa(i)))
		assert.Nil(t, dposContext.BecomeCandidate(config, header, candidate))
	}
	assert.Nil(t, dposContext.SetValidators(validators))
	assert.Nil(t, epochContext.kickoutValidator(config, header, testEpoch))
	candidateMap = getCandidates(dposContext.CandidateTrie())
	assert.Equal(t, maxValidatorSize, len(candidateMap))

	// only one validator mint count is not enough
	dposContext, err = types.NewDposContext(db)
	assert.Nil(t, err)
	epochContext = &EpochContext{
		TimeStamp:   epochInterval,
		DposContext: dposContext,
		statedb:     stateDB,
	}
	validators = []common.Address{}
	for i := 0; i < maxValidatorSize; i++ {
		validator := common.BytesToAddress([]byte("addr" + strconv.Itoa(i)))
		validators = append(validators, validator)
		assert.Nil(t, dposContext.BecomeCandidate(config, header, validator))
		if i == 0 {
			setTestMintCnt(dposContext, testEpoch, validator, atLeastMintCnt-1)
		} else {
			setTestMintCnt(dposContext, testEpoch, validator, atLeastMintCnt)
		}
	}
	assert.Nil(t, dposContext.BecomeCandidate(config, header, common.BytesToAddress([]byte("addr"))))
	assert.Nil(t, dposContext.SetValidators(validators))
	assert.Nil(t, epochContext.kickoutValidator(config, header, testEpoch))
	candidateMap = getCandidates(dposContext.CandidateTrie())
	assert.Equal(t, maxValidatorSize, len(candidateMap))
	assert.False(t, candidateMap[common.BytesToAddress([]byte("addr"+strconv.Itoa(0)))])

	// epochTime is not complete, all validators mint enough block at least
	dposContext, err = types.NewDposContext(db)
	assert.Nil(t, err)
	epochContext = &EpochContext{
		TimeStamp:   epochInterval / 2,
		DposContext: dposContext,
		statedb:     stateDB,
	}
	validators = []common.Address{}
	for i := 0; i < maxValidatorSize; i++ {
		validator := common.BytesToAddress([]byte("addr" + strconv.Itoa(i)))
		validators = append(validators, validator)
		assert.Nil(t, dposContext.BecomeCandidate(config, header, validator))
		setTestMintCnt(dposContext, testEpoch, validator, atLeastMintCnt/2)
	}
	for i := maxValidatorSize; i < maxValidatorSize *2; i++ {
		candidate := common.BytesToAddress([]byte("addr" + strconv.Itoa(i)))
		assert.Nil(t, dposContext.BecomeCandidate(config, header, candidate))
	}
	assert.Nil(t, dposContext.SetValidators(validators))
	assert.Nil(t, epochContext.kickoutValidator(config, header, testEpoch))
	candidateMap = getCandidates(dposContext.CandidateTrie())
	assert.Equal(t, maxValidatorSize *2, len(candidateMap))

	// epochTime is not complete, all validators didn't mint enough block at least
	dposContext, err = types.NewDposContext(db)
	assert.Nil(t, err)
	epochContext = &EpochContext{
		TimeStamp:   epochInterval / 2,
		DposContext: dposContext,
		statedb:     stateDB,
	}
	validators = []common.Address{}
	for i := 0; i < maxValidatorSize; i++ {
		validator := common.BytesToAddress([]byte("addr" + strconv.Itoa(i)))
		validators = append(validators, validator)
		assert.Nil(t, dposContext.BecomeCandidate(config, header, validator))
		setTestMintCnt(dposContext, testEpoch, validator, atLeastMintCnt/2-1)
	}
	for i := maxValidatorSize; i < maxValidatorSize *2; i++ {
		candidate := common.BytesToAddress([]byte("addr" + strconv.Itoa(i)))
		assert.Nil(t, dposContext.BecomeCandidate(config, header, candidate))
	}
	assert.Nil(t, dposContext.SetValidators(validators))
	assert.Nil(t, epochContext.kickoutValidator(config, header, testEpoch))
	candidateMap = getCandidates(dposContext.CandidateTrie())
	assert.Equal(t, maxValidatorSize, len(candidateMap))

	dposContext, err = types.NewDposContext(db)
	assert.Nil(t, err)
	epochContext = &EpochContext{
		TimeStamp:   epochInterval / 2,
		DposContext: dposContext,
		statedb:     stateDB,
	}
	assert.NotNil(t, epochContext.kickoutValidator(config, header, testEpoch))
	dposContext.SetValidators([]common.Address{})
	assert.NotNil(t, epochContext.kickoutValidator(config, header, testEpoch))
}

func setTestMintCnt(dposContext *types.DposContext, epoch int64, validator common.Address, count int64) {
	for i := int64(0); i < count; i++ {
		dposContext.UpdateMintCnt(epoch*epochInterval, epoch*epochInterval+blockInterval, validator, epochInterval)
	}
}

func getCandidates(candidateTrie *trie.Trie) map[common.Address]bool {
	candidateMap := map[common.Address]bool{}
	iter := trie.NewIterator(candidateTrie.NodeIterator(nil))
	for iter.Next() {
		candidateMap[common.BytesToAddress(iter.Value)] = true
	}
	return candidateMap
}

func TestEpochContextTryElect(t *testing.T) {

	// for not multi vote
	config := &params.ChainConfig{
		MultiVoteBlock: big.NewInt(100),
	}
	header := &types.Header{
		Number: big.NewInt(99),
	}

	checkEpochContextTryElect(t,config,header)

	// for multi vote
	config = &params.ChainConfig{
		MultiVoteBlock: big.NewInt(0),
	}
	header = &types.Header{
		Number: big.NewInt(100),
	}

	checkEpochContextTryElect(t,config,header)
}

func checkEpochContextTryElect(t *testing.T, config *params.ChainConfig, header *types.Header) {

	db := socdb.NewMemDatabase()
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(db))
	dposContext, err := types.NewDposContext(db)
	assert.Nil(t, err)
	epochContext := &EpochContext{
		TimeStamp:   epochInterval,
		DposContext: dposContext,
		statedb:     stateDB,
	}
	atLeastMintCnt := epochInterval / blockInterval / maxValidatorSize / 2
	testEpoch := int64(1)
	validators := []common.Address{}
	for i := 0; i < maxValidatorSize; i++ {
		validator := common.BytesToAddress([]byte("addr" + strconv.Itoa(i)))
		validators = append(validators, validator)
		assert.Nil(t, dposContext.BecomeCandidate(config, header, validator))
		assert.Nil(t, dposContext.Delegate(config, header, validator, validator))
		stateDB.SetBalance(validator, big.NewInt(1))
		setTestMintCnt(dposContext, testEpoch, validator, atLeastMintCnt-1)
	}
	dposContext.BecomeCandidate(config, header, common.BytesToAddress([]byte("more")))
	assert.Nil(t, dposContext.SetValidators(validators))

	// genesisEpoch == parentEpoch do not kickout
	genesis := &types.Header{
		Time: big.NewInt(0),
		Number: big.NewInt(0) ,
	}
	parent := &types.Header{
		Time: big.NewInt(epochInterval - blockInterval),
	}
	oldHash := dposContext.EpochTrie().Hash()
	assert.Nil(t, epochContext.tryElect(config, header, genesis, parent))
	result, err := dposContext.GetValidators()
	assert.Nil(t, err)
	assert.Equal(t, maxValidatorSize, len(result))
	for _, validator := range result {
		assert.True(t, strings.Contains(string(validator[:]), "addr"))
	}
	assert.NotEqual(t, oldHash, dposContext.EpochTrie().Hash())

	// genesisEpoch != parentEpoch  and have none mintCnt do not kickout
	genesis = &types.Header{
		Time: big.NewInt(-epochInterval),
		Number: big.NewInt(0) ,
	}
	parent = &types.Header{
		Difficulty: big.NewInt(1),
		Time:       big.NewInt(epochInterval - blockInterval),
	}
	epochContext.TimeStamp = epochInterval
	oldHash = dposContext.EpochTrie().Hash()
	assert.Nil(t, epochContext.tryElect(config, header, genesis, parent))
	result, err = dposContext.GetValidators()
	assert.Nil(t, err)
	assert.Equal(t, maxValidatorSize, len(result))
	for _, validator := range result {
		assert.True(t, strings.Contains(string(validator[:]), "addr"))
	}

	assert.NotEqual(t, oldHash, dposContext.EpochTrie().Hash())

	// genesisEpoch != parentEpoch kickout
	genesis = &types.Header{
		Time: big.NewInt(0),
		Number: big.NewInt(0) ,
	}
	parent = &types.Header{
		Time: big.NewInt(epochInterval*2 - blockInterval),
	}
	epochContext.TimeStamp = epochInterval * 2
	oldHash = dposContext.EpochTrie().Hash()
	assert.Nil(t, epochContext.tryElect(config, header, genesis, parent))
	result, err = dposContext.GetValidators()
	assert.Nil(t, err)
	assert.Equal(t, safeSize, len(result))
	moreCnt := 0
	for _, validator := range result {
		if strings.Contains(string(validator[:]), "more") {
			moreCnt++
		}
	}
	assert.Equal(t, 1, moreCnt)
	assert.NotEqual(t, oldHash, dposContext.EpochTrie().Hash())

	// parentEpoch == currentEpoch do not elect
	genesis = &types.Header{
		Time: big.NewInt(0),
		Number: big.NewInt(0) ,
	}
	parent = &types.Header{
		Time: big.NewInt(epochInterval),
	}
	epochContext.TimeStamp = epochInterval + blockInterval
	oldHash = dposContext.EpochTrie().Hash()
	assert.Nil(t, epochContext.tryElect(config, header, genesis, parent))
	result, err = dposContext.GetValidators()
	assert.Nil(t, err)
	assert.Equal(t, safeSize, len(result))
	assert.Equal(t, oldHash, dposContext.EpochTrie().Hash())
}


func TestEpochContextTryElectWhenMultiVoteBlock(t *testing.T) {
	config := &params.ChainConfig{
		MultiVoteBlock:      big.NewInt(100),
	}
	header := &types.Header{
		Number: big.NewInt(99),
	}

	voteMap := map[common.Address]common.Address{
		common.HexToAddress("0x44d1ce0b7cb3588bca96151fe1bc05af38f91b6e"): common.HexToAddress("0xb040353ec0f2c113d5639444f7253681aecda1f8"),
		common.HexToAddress("0xa60a3886b552ff9992cfcd208ec1152079e046c2"): common.HexToAddress("0xb040353ec0f2c113d5639444f7253681aecda1f8"),
		common.HexToAddress("0x4e080e49f62694554871e669aeb4ebe17c4a9670"): common.HexToAddress("0xb040353ec0f2c113d5639444f7253681aecda1f8"),
	}
	balance := int64(5)
	db := socdb.NewMemDatabase()
	stateDB, _ := state.New(common.Hash{}, state.NewDatabase(db))
	dposContext, err := types.NewDposContext(db)
	assert.Nil(t, err)

	epochContext := &EpochContext{
		DposContext: dposContext,
		statedb:     stateDB,
	}

	mapElector := map[common.Address]bool{}
	for candidate, elector := range voteMap {
		assert.Nil(t, dposContext.BecomeCandidate(config,header,candidate))
			if mapElector[elector] == false {
				stateDB.SetBalance(elector, big.NewInt(balance))
				mapElector[elector] = true
			}
			assert.Nil(t, dposContext.Delegate(config, header, elector, candidate))
	}

	voteIterator := trie.NewIterator(dposContext.VoteTrie().NodeIterator(nil))
	existVote := voteIterator.Next()
	for existVote {
		assert.Equal(t,voteMap[common.BytesToAddress(voteIterator.Value)],common.BytesToAddress(voteIterator.Key))
		existVote = voteIterator.Next()
	}
	// genesisEpoch == parentEpoch do not kickout
	genesis := &types.Header{
		Time: big.NewInt(0),
		Number: big.NewInt(0) ,
	}
	parent := &types.Header{
		Time: big.NewInt(epochInterval - blockInterval),
	}
    // when header.Number = config.MultiVoteBlock
	header.Number = config.MultiVoteBlock

	assert.Nil(t, epochContext.tryElect(config, header, genesis, parent))

	voteIterator = trie.NewIterator(dposContext.VoteTrie().NodeIterator(nil))
	existVote = voteIterator.Next()
	for existVote {
		assert.Equal(t,append(voteMap[common.BytesToAddress(voteIterator.Value)].Bytes(),voteIterator.Value...),voteIterator.Key)
		existVote = voteIterator.Next()
	}
}
