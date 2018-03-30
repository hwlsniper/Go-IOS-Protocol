package protocol

import (
	"IOS/src/iosbase"
)

type ReqType int

const (
	ReqPrePrepare   ReqType = iota
	ReqPrepare
	ReqCommit
	ReqSubmitTxPack
	ReqPublishTx
	ReqNewBlock
)

type ResState int

const (
	Accepted ResState = iota
	Reject
	Error
)

type NetworkFilter struct {
	base iosbase.Network
	*RuntimeData

	reqChan chan iosbase.Request
	resChan chan iosbase.Response
}

func (n *NetworkFilter) Send(request iosbase.Request) chan iosbase.Response {
	return n.base.Send(request)
}

func (n *NetworkFilter) Init(rd *RuntimeData, nw iosbase.Network, port uint16) error {
	n.RuntimeData = rd
	n.base = nw
	var err error
	n.reqChan, n.resChan, err = n.base.Listen(port)
	return err
}

func (n *NetworkFilter) Router(replica Replica, recorder Recorder, holder DataHolder) {
	for n.isRunning { // TODO bug：lock here!
		req := <-n.reqChan
		switch req.ReqType {
		case int(ReqPrePrepare):
			fallthrough
		case int(ReqPrepare):
			fallthrough
		case int(ReqSubmitTxPack):
			fallthrough
		case int(ReqCommit):
			n.replicaFilter(replica, n.resChan, req)
		case int(ReqPublishTx):
			n.recorderFilter(recorder, n.resChan, req)
		case int(ReqNewBlock):
			n.dataholderFilter(holder, n.resChan, req)
		}
	}
}

func (n *NetworkFilter) replicaFilter(replica Replica, res chan iosbase.Response, req iosbase.Request) {
	// 1. if req comes from right member
	if !n.view.isPrimary(req.From) && !n.view.isBackup(req.From) {
		res <- authorityError(req)
		return
	}

	switch req.ReqType {
	case int(ReqSubmitTxPack):
		var txpool iosbase.TxPool
		txpool.Decode(req.Body)
		replica.OnTxPack(txpool)
		return
	}

	// 2. if req in right phase
	switch n.phase {
	case StartPhase:
		res <- invalidPhase(req)
	case PrePreparePhase:
		if req.ReqType == int(ReqPrePrepare) {
			res <- accept(req)
			replica.OnRequest(req)
		} else {
			res <- invalidPhase(req)
		}
	case PreparePhase:
		if req.ReqType == int(ReqPrepare) {
			res <- accept(req)
			replica.OnRequest(req)
		} else {
			res <- invalidPhase(req)
		}
	case CommitPhase:
		if req.ReqType == int(ReqCommit) {
			res <- accept(req)
			replica.OnRequest(req)
		} else {
			res <- invalidPhase(req)
		}
	case PanicPhase:
		res <- internalError(req)
	case EndPhase:
		res <- invalidPhase(req)
	default:
		res <- internalError(req)
	}

}

func (n *NetworkFilter) recorderFilter(recorder Recorder, resChan chan iosbase.Response, req iosbase.Request) {
	switch req.ReqType {
	case int(ReqPublishTx):
		var tx iosbase.Tx
		err := tx.Decode(req.Body)
		if err != nil {
			resChan <- illegalTx(req)
		}
		recorder.PublishTx(tx)
	default:
		resChan <- internalError(req)
	}
}

func (n *NetworkFilter) dataholderFilter(holder DataHolder, resChan chan iosbase.Response, req iosbase.Request) {
	switch req.ReqType {
	case int(ReqNewBlock):
		var blk iosbase.Block
		err := blk.Decode(req.Body)
		if err != nil {
			resChan <- illegalTx(req)
		}
		holder.OnNewBlock(&blk)
	default:
		resChan <- internalError(req)
	}

}

func invalidPhase(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		From:        req.To,
		To:          req.From,
		Code:        int(Reject),
		Description: "Error: Invalid phase",
	}
}

func accept(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		From:        req.To,
		To:          req.From,
		Code:        int(Accepted),
		Description: "",
	}
}

func internalError(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		From:        req.To,
		To:          req.From,
		Code:        int(Reject),
		Description: "Error: Internal error",
	}
}

func authorityError(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		From:        req.To,
		To:          req.From,
		Code:        int(Reject),
		Description: "Error: Authority error",
	}
}

func illegalTx(req iosbase.Request) iosbase.Response {
	return iosbase.Response{
		From:        req.To,
		To:          req.From,
		Code:        int(Error),
		Description: "ERROR: Illegal Transaction",
	}
}

func (n *NetworkFilter) BroadcastToMembers(req iosbase.Request) {

}
