package devicecommands

import (
	"net/http"

	"github.com/gluster/glusterd2/glusterd2/gdctx"
	"github.com/gluster/glusterd2/glusterd2/peer"
	restutils "github.com/gluster/glusterd2/glusterd2/servers/rest/utils"
	"github.com/gluster/glusterd2/glusterd2/transaction"
	"github.com/gluster/glusterd2/pkg/api"

	"github.com/gorilla/mux"
	"github.com/pborman/uuid"
)

func deviceAddHandler(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	logger := gdctx.GetReqLogger(ctx)

	req := new(api.AddDeviceReq)
	if err := restutils.UnmarshalRequest(r, req); err != nil {
		logger.WithError(err).Error("Failed to Unmarshal request")
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "Unable to marshal request", api.ErrCodeDefault)
		return
	}
	peerID := mux.Vars(r)["peerid"]
	if peerID == "" {
		restutils.SendHTTPError(ctx, w, http.StatusBadRequest, "peerid not present in request", api.ErrCodeDefault)
		return
	}
	peerInfo, err := peer.GetPeer(peerID)
	if err != nil {
		logger.WithError(err).WithField("peerid", peerID).Error("Peer ID not found in store")
		restutils.SendHTTPError(ctx, w, http.StatusNotFound, "Peer Id not found in store", api.ErrCodeDefault)
		return
	}
	/*
		var deviceList []api.DeviceInfo
		for _, name := range req.Devices {
			tempDevice := api.DeviceInfo{
				Name: name,
			}
			deviceList = append(deviceList, tempDevice)
		}*/
	txn := transaction.NewTxn(ctx)
	defer txn.Cleanup()
	lock, unlock, err := transaction.CreateLockSteps(string(peerInfo.ID))
	txn.Nodes = []uuid.UUID{peerInfo.ID}
	txn.Steps = []*transaction.Step{
		lock,
		{
			DoFunc: "prepare-device",
			Nodes:  txn.Nodes,
		},
		unlock,
	}
	err = txn.Ctx.Set("peerid", peerID)
	if err != nil {
		logger.WithError(err).Error("Failed to set data for transaction")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	err = txn.Ctx.Set("device-details", req)
	if err != nil {
		logger.WithError(err).Error("Failed to set data for transaction")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, err.Error(), api.ErrCodeDefault)
		return
	}
	err = txn.Do()
	if err != nil {
		logger.WithError(err).Error("Transaction to prepare device failed")
		restutils.SendHTTPError(ctx, w, http.StatusInternalServerError, "Transaction to prepare device failed", api.ErrCodeDefault)
		return
	}
	peerInfo, err = peer.GetPeer(peerID)
	if err != nil {
		logger.WithError(err).Error("Failed to get peer from store")
	}
	restutils.SendHTTPResponse(ctx, w, http.StatusOK, peerInfo)
}
