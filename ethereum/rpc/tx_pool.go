package rpc

import (
	"errors"

	rpcclient "github.com/cometbft/cometbft/rpc/client"
)

func (b *BackendImpl) PendingTransactionsCount() (int, error) {
	client, ok := b.clientCtx.Client.(rpcclient.MempoolClient)
	if !ok {
		return 0, errors.New("failed to assert MempoolClient")
	}

	res, err := client.UnconfirmedTxs(b.ctx, nil)
	if err != nil {
		return 0, err
	}
	return res.Count, nil
}
