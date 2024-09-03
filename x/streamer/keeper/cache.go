package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// streamInfo is a cache for streams.
type streamInfo struct {
	nextID       int
	streamIDToID map[uint64]int
	IDToStream   []types.Stream
	totalDistr   sdk.Coins
}

func newStreamInfo(streams []types.Stream) *streamInfo {
	info := &streamInfo{
		nextID:       0,
		streamIDToID: make(map[uint64]int),
		IDToStream:   make([]types.Stream, 0),
	}
	for _, stream := range streams {
		info.addDistrCoins(stream, sdk.Coins{})
	}
	return info
}

func (i *streamInfo) addDistrCoins(stream types.Stream, coins sdk.Coins) types.Stream {
	id, ok := i.streamIDToID[stream.Id]
	if ok {
		i.IDToStream[id].DistributedCoins = i.IDToStream[id].DistributedCoins.Add(coins...)
	} else {
		id = i.nextID
		i.nextID++
		i.streamIDToID[stream.Id] = id
		stream.DistributedCoins = stream.DistributedCoins.Add(coins...)
		i.IDToStream = append(i.IDToStream, stream)
	}
	i.totalDistr = i.totalDistr.Add(coins...)
	return i.IDToStream[id]
}

func (i *streamInfo) getStreams() []types.Stream {
	return i.IDToStream
}

// gaugeInfo is a cache for gauges.
type gaugeInfo struct {
	nextID      int
	gaugeIDToID map[uint64]int
	IDToGauge   []incentivestypes.Gauge
}

func newGaugeInfo() *gaugeInfo {
	return &gaugeInfo{
		nextID:      0,
		gaugeIDToID: make(map[uint64]int),
		IDToGauge:   make([]incentivestypes.Gauge, 0),
	}
}

func (i *gaugeInfo) addDistrCoins(gauge incentivestypes.Gauge, coins sdk.Coins) incentivestypes.Gauge {
	id, ok := i.gaugeIDToID[gauge.Id]
	if ok {
		i.IDToGauge[id].Coins = i.IDToGauge[id].Coins.Add(coins...)
	} else {
		id = i.nextID
		i.nextID++
		i.gaugeIDToID[gauge.Id] = id
		gauge.Coins = gauge.Coins.Add(coins...)
		i.IDToGauge = append(i.IDToGauge, gauge)
	}
	return i.IDToGauge[id]
}

func (i *gaugeInfo) getGauges() []incentivestypes.Gauge {
	return i.IDToGauge
}
