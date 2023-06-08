package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hibiken/asynq"
)

const AllAddresses = "/address"

type AddressResp struct {
	Msg    []Msg  `json:"msg"`
	Status string `json:"status"`
}

type Msg struct {
	Hash  common.Address `json:"Hash"`
	Chain int64          `json:"Chain"`
	Users interface{}    `json:"Users"`
}

func OnlineUsersHandler(ctx context.Context, task *asynq.Task) error {
	resp, err := http.Get(conf.Config.UserAppUrl.JoinPath(AllAddresses).String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	addResp := AddressResp{}
	if err := json.Unmarshal(body, &addResp); err != nil {
		return err
	}
	for _, msg := range addResp.Msg {
		conf.OnlineUsers.NewAdd(msg.Hash)
	}
	_, _ = task, ctx
	return nil
}
