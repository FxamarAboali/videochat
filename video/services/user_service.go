package services

import (
	"context"
	"encoding/json"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"nkonev.name/video/dto"
	. "nkonev.name/video/logger"
	"nkonev.name/video/utils"
)

type UserService struct {
	livekitRoomClient *lksdk.RoomServiceClient
}

func NewUserService(livekitRoomClient *lksdk.RoomServiceClient) *UserService {
	return &UserService{
		livekitRoomClient: livekitRoomClient,
	}
}

func (h *UserService) CountUsers(ctx context.Context, roomName string) (int64, error) {
	var req *livekit.ListParticipantsRequest = &livekit.ListParticipantsRequest{Room: roomName}
	participants, err := h.livekitRoomClient.ListParticipants(ctx, req)
	if err != nil {
		return 0, err
	}

	var usersCount = int64(len(participants.Participants))
	return usersCount, nil
}

func (vh *UserService) GetVideoParticipants(chatId int64, ctx context.Context) ([]int64, error) {
	roomName := utils.GetRoomNameFromId(chatId)

	var ret = []int64{}
	var set = make(map[int64]bool)

	lpr := &livekit.ListParticipantsRequest{Room: roomName}
	participants, err := vh.livekitRoomClient.ListParticipants(ctx, lpr)
	if err != nil {
		Logger.Errorf("Unable to get participants %v", err)
		return ret, err
	}

	for _, participant := range participants.Participants {
		md := &dto.MetadataDto{}
		err = json.Unmarshal([]byte(participant.Metadata), md)
		if err != nil {
			Logger.Errorf("got error during parsing metadata from chatId=%v, %v", chatId, err)
			continue
		}
		set[md.UserId] = true
	}

	for key, value := range set {
		if value {
			ret = append(ret, key)
		}
	}

	return ret, nil
}
