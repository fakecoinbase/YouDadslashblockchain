package api

import (
	"bytes"
	"time"

	"github.com/YouDad/blockchain/core"
	"github.com/YouDad/blockchain/global"
	"github.com/YouDad/blockchain/log"
	"github.com/YouDad/blockchain/network"
	"github.com/YouDad/blockchain/utils"
)

// 同步
func Sync(group int) error {
	group = group % global.MaxGroupNum
	network.Register()
	bc := core.GetBlockchain(group)

	if bc.GetHeight() < 0 {
		genesis, err := GetGenesis(group)
		for err != nil {
			log.Warn(err)
			time.Sleep(5 * time.Second)
			log.Warn(network.GetKnownNodes())
			network.UpdateSortedNodes()
			genesis, err = GetGenesis(group)
		}
		bc.AddBlock(genesis)
		core.GetUTXOSet(group).Reindex()
	}

	genesis := bc.GetGenesis()
	lastest := bc.GetLastest()
	var height int32
	var address string
	var err error
	for {
		height, err, address = SendVersion(group, lastest.Height, genesis.Hash(), lastest.Hash())
		log.Warn(err)
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	SyncBlocks(group, height, address)
	return nil
}

// 同步group组的区块，最新的区块高度是newHeight，发送者是address
func SyncBlocks(group int, newHeight int32, address string) {
	log.Debugln("SyncBlocks", "{{{{{{{{")
	syncBlocks(group, newHeight, address)
	log.Debugln("SyncBlocks", "}}}}}}}}")
}

func syncBlocks(group int, newHeight int32, address string) {
	bc := core.GetBlockchain(group)

	lastestHeight := bc.GetHeight()
	if newHeight <= lastestHeight {
		// 认为不需要同步
		return
	}

	if lastestHeight == -1 {
		// 同步不了，没有genesis
		return
	}

	global.SyncMutex.Lock()
	defer global.SyncMutex.Unlock()
	log.Debugln("SyncBlock Start!")
	lastest := bc.GetLastest()
	originHash := lastest.Hash()

	// 二分同步找到差异点
	var l int32 = 0
	var r int32 = lastestHeight
	for r >= l {
		m := l + (r-l)/2

		block := bc.GetBlockByHeight(m)
		hash, err := CallbackGetHash(group, m, address)
		if err != nil {
			log.Warn(err)
			return
		}

		if bytes.Compare(hash, block.Hash()) == 0 {
			l = m + 1
		} else {
			r = m - 1
		}
	}

	// 将lastest移到最新的相同点
	lastest = bc.GetBlockByHeight(r)
	if lastest == nil {
		// 有可能r是-1
		log.Errln("二分nil")
	}

	lastestBytes := utils.Encode(lastest)
	bc.SetLastest(lastestBytes)

	// 获得group组的l到newHeight高度的区块
	lastestHash := lastest.Hash()
	blocks, err := CallbackGetBlocks(group, l, newHeight, lastestHash, address)
	if err != nil {
		log.Warn(err)
		lastestBytes = bc.Get(originHash)
		bc.SetLastest(lastestBytes)
		return
	}

	// 撤销原先lastestHeight到r的高度的UTXOSet集合
	set := core.GetUTXOSet(group)
	for i := lastestHeight; i > r; i-- {
		block := bc.GetBlockByHeight(i)
		set.Reverse(block)
		bc.DeleteBlock(block)
	}

	// 然后将后面所有区块都追加到lastest的后面
	for _, block := range blocks {
		if bytes.Compare(block.PrevHash, lastestHash) != 0 {
			break
		}

		bc.AddBlock(block)
		set.Update(block)
		lastestHash = block.Hash()
	}
}
