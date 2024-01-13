package handlers

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
	"nkonev.name/video/auth"
	"nkonev.name/video/client"
	"nkonev.name/video/dto"
	"nkonev.name/video/logger"
	. "nkonev.name/video/logger"
	"nkonev.name/video/producer"
	"nkonev.name/video/services"
	"nkonev.name/video/tasks"
	"nkonev.name/video/utils"
)

type InviteHandler struct {
	dialRedisRepository   *services.DialRedisRepository
	chatClient            *client.RestClient
	dialStatusPublisher   *producer.RabbitDialStatusPublisher
	notificationPublisher *producer.RabbitNotificationsPublisher
	userService           *services.UserService
	chatDialerService     *tasks.ChatDialerService
	chatInvitationService *services.ChatInvitationService
}

const EventMissedCall = "missed_call"

func NewInviteHandler(dialService *services.DialRedisRepository, chatClient *client.RestClient, dialStatusPublisher *producer.RabbitDialStatusPublisher, notificationPublisher *producer.RabbitNotificationsPublisher, userService *services.UserService, chatDialerService *tasks.ChatDialerService, chatInvitationService *services.ChatInvitationService) *InviteHandler {
	return &InviteHandler{
		dialRedisRepository:   dialService,
		chatClient:            chatClient,
		dialStatusPublisher:   dialStatusPublisher,
		notificationPublisher: notificationPublisher,
		userService:           userService,
		chatDialerService:     chatDialerService,
		chatInvitationService: chatInvitationService,
	}
}

// used by owner to add or remove from dial list
func (vh *InviteHandler) ProcessCallInvitation(c echo.Context) error {
	var userPrincipalDto, ok = c.Get(utils.USER_PRINCIPAL_DTO).(*auth.AuthResult)
	if !ok {
		Logger.Errorf("Error during getting auth context")
		return errors.New("Error during getting auth context")
	}

	chatId, err := GetPathParamAsInt64(c, "id")
	if err != nil {
		return err
	}

	callee, err := utils.ParseInt64(c.QueryParam("userId"))
	if err != nil {
		return err
	}

	addToCallCall, err := utils.ParseBoolean(c.QueryParam("call"))
	if err != nil {
		return err
	}

	// check my access to chat
	if ok, err := vh.chatClient.CheckAccess(userPrincipalDto.UserId, chatId, c.Request().Context()); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	} else if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	if (addToCallCall) {
		return c.NoContent(vh.addToCalling(c, callee, chatId, userPrincipalDto))
	} else {
		return c.NoContent(vh.removeFromCalling(c, callee, chatId, userPrincipalDto))
	}
}

func (vh *InviteHandler) checkAccessOverCall(ctx context.Context, callee int64, chatId int64, userPrincipalDto *auth.AuthResult) (bool, int) {
	// check participant's access to chat
	if ok, err := vh.chatClient.CheckAccess(callee, chatId, ctx); err != nil {
		return false, http.StatusInternalServerError
	} else if !ok {
		return false, http.StatusUnauthorized
	}

	ownerId, err := vh.dialRedisRepository.GetDialMetadata(ctx, chatId)
	if err != nil {
		logger.GetLogEntry(ctx).Errorf("Error %v", err)
		return false, http.StatusInternalServerError
	}
	if ownerId == services.NoUser {
		// ok
	} else if userPrincipalDto.UserId != ownerId {
		logger.GetLogEntry(ctx).Infof("Call already started in this chat %v by %v", chatId, ownerId)
		return false, http.StatusAccepted
	}
	return true, http.StatusOK
}

func (vh *InviteHandler) addToCalling(c echo.Context, callee int64, chatId int64, userPrincipalDto *auth.AuthResult) int {
	ok, code := vh.checkAccessOverCall(c.Request().Context(), callee, chatId, userPrincipalDto)
	if !ok {
		return code
	}

	status, err := vh.dialRedisRepository.GetUserCallStatus(c.Request().Context(), callee)
	if err != nil {
		logger.GetLogEntry(c.Request().Context()).Errorf("Error %v", err)
		return http.StatusInternalServerError
	}

	if !services.CanOverrideCallStatus(status) {
		return http.StatusConflict
	}

	// for better user experience
	vh.chatInvitationService.SendInvitations(c.Request().Context(), chatId, userPrincipalDto.UserId, []int64{callee})

	err = vh.dialRedisRepository.AddToDialList(c.Request().Context(), callee, chatId, userPrincipalDto.UserId, services.CallStatusInviting)
	if err != nil {
		logger.GetLogEntry(c.Request().Context()).Errorf("Error %v", err)
		return http.StatusInternalServerError
	}

	return http.StatusOK
}

func (vh *InviteHandler) removeFromCalling(c echo.Context, callee int64, chatId int64, userPrincipalDto *auth.AuthResult) int {
	ok, code := vh.checkAccessOverCall(c.Request().Context(), callee, chatId, userPrincipalDto)
	if !ok {
		return code
	}

	code = vh.removeFromCallingList(c, chatId, []int64{callee}, services.CallStatusRemoving)
	if code != http.StatusOK {
		return code
	}

	// if we remove user from call - send them EventMissedCall notification
	vh.sendMissedCallNotification(chatId, c.Request().Context(), userPrincipalDto, []int64{callee})

	return http.StatusOK
}

// user enters to call somehow, either by clicking green tube or opening .../video link
func (vh *InviteHandler) ProcessEnterToDial(c echo.Context) error {
	var userPrincipalDto, ok = c.Get(utils.USER_PRINCIPAL_DTO).(*auth.AuthResult)
	if !ok {
		Logger.Errorf("Error during getting auth context")
		return errors.New("Error during getting auth context")
	}

	chatId, err := GetPathParamAsInt64(c, "id")
	if err != nil {
		return err
	}

	// check my access to chat
	if ok, err := vh.chatClient.CheckAccess(userPrincipalDto.UserId, chatId, c.Request().Context()); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	} else if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	// during entering into dial. Returns status: true which means that frontend should (initially) draw the calling.
	// Now it used only in tet-a-tet.
	// If we are in the tet-a-tet
	basicChatInfo, err := vh.chatClient.GetBasicChatInfo(chatId, userPrincipalDto.UserId, c.Request().Context()) // tet-a-tet
	if err != nil {
		return err
	}

	usersOfChat := basicChatInfo.ParticipantIds

	// in this block we start calling in case tet-a-tet
	if basicChatInfo.TetATet && len(usersOfChat) > 0 {
		if err != nil {
			logger.GetLogEntry(c.Request().Context()).Errorf("Error %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		var oppositeUser *int64 = getOppositeUserOfTetAtTet(usersOfChat, userPrincipalDto.UserId)

		// uniq users by userId
		usersOfVideo, err := vh.userService.GetVideoParticipants(chatId, c.Request().Context())
		if err != nil {
			logger.GetLogEntry(c.Request().Context()).Errorf("Error %v", err)
			return c.NoContent(http.StatusInternalServerError)
		}

		var oppositeUserOfVideo *int64 = getOppositeUserOfTetAtTet(usersOfVideo, userPrincipalDto.UserId)

		// oppositeUserOfVideo is need for case when your counterpart enters into call (not entered until this moment) and this (oppositeUserOfVideo == nil) prevents us to start calling him back
		// and we(behalf user) doesn't have incoming call
		if oppositeUserOfVideo == nil && oppositeUser != nil {
			// we should call the counterpart (opposite user)
			vh.addToCalling(c, *oppositeUser, chatId, userPrincipalDto)
		}
	}

	// remove myself from a call
	// we call it in case opposite/owner user has incoming (this) call
	// react on "take the phone" (pressing green tube) which cancels ringing logic for opposite/owner user (or myself)
	vh.removeFromCallingList(c, chatId, []int64{userPrincipalDto.UserId}, services.CallStatusInCall)

	return c.NoContent(http.StatusOK)
}

func getOppositeUserOfTetAtTet(users []int64, me int64) *int64 {
	var oppositeUser *int64
	for _, userId := range users {
		if userId != me {
			var deUid = userId
			oppositeUser = &deUid
			break
		}
	}
	return oppositeUser
}

func (vh *InviteHandler) ProcessAcceptCall(c echo.Context) error {
	var userPrincipalDto, ok = c.Get(utils.USER_PRINCIPAL_DTO).(*auth.AuthResult)
	if !ok {
		Logger.Errorf("Error during getting auth context")
		return errors.New("Error during getting auth context")
	}

	chatId, err := GetPathParamAsInt64(c, "id")
	if err != nil {
		return err
	}

	// check my access to chat
	if ok, err := vh.chatClient.CheckAccess(userPrincipalDto.UserId, chatId, c.Request().Context()); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	} else if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	return c.NoContent(vh.removeFromCallingList(c, chatId, []int64{userPrincipalDto.UserId}, services.CallStatusInCall))
}

func (vh *InviteHandler) ProcessCancelCall(c echo.Context) error {
	var userPrincipalDto, ok = c.Get(utils.USER_PRINCIPAL_DTO).(*auth.AuthResult)
	if !ok {
		Logger.Errorf("Error during getting auth context")
		return errors.New("Error during getting auth context")
	}

	chatId, err := GetPathParamAsInt64(c, "id")
	if err != nil {
		return err
	}

	// check my access to chat
	if ok, err := vh.chatClient.CheckAccess(userPrincipalDto.UserId, chatId, c.Request().Context()); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	} else if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	return c.NoContent(vh.removeFromCallingList(c, chatId, []int64{userPrincipalDto.UserId}, services.CallStatusCancelling))
}

// question: how not to overwhelm the system by iterating over all the users and all the chats ?
// answer: using opened rooms and rooms are going to be closed - see livekit's room.empty_timeout

// TODO not here but in chat_dialer.go :: makeDial()
//  also, it's subscription on chat events;
//  a) periodically send dto.VideoCallInvitation[true|false] (call particular user to video call)
//  b) periodically send dto.VideoDialChanges (update progressbar in ChatParticipants.vue)
//  implement the algorithm:
//  run over all rooms (see livekit's room.empty_timeout), then get room's chats, then get chat's participants
//  if (we have chat participant but no their counterpart in the room) {
//    if (EXISTS user_call_state:<userId>) {
//      send VideoCallInvitation(true)
//    } else {
//      send VideoCallInvitation(false)
//    }
//    send dto.VideoIsInvitingDto -> dto.VideoDialChanges(false)
//  } else if (we have chat participant and we have their counterpart in the room) {
//    send dto.VideoIsInvitingDto -> dto.VideoDialChanges(true)
//  }

// TODO consider reworking dto.VideoCallInvitation in manner to remove App.vue's timer
//  add status "inviting", "closing"
//  when we have "closing" - send "false" all the time empty room exists



func (vh *InviteHandler) removeFromCallingList(c echo.Context, chatId int64, usersOfDial []int64, callStatus string) int {
	ownerId, err := vh.dialRedisRepository.GetDialMetadata(c.Request().Context(), chatId)
	if err != nil {
		logger.GetLogEntry(c.Request().Context()).Errorf("Error %v", err)
		return http.StatusInternalServerError
	}
	if ownerId == services.NoUser {
		return http.StatusOK
	}

	// we remove callee by setting status
	for _, userId := range usersOfDial {
		err = vh.setUserStatus(c.Request().Context(), userId, chatId, callStatus)
		if err != nil {
			logger.GetLogEntry(c.Request().Context()).Errorf("Error %v", err)
		}
	}
	if err != nil {
		logger.GetLogEntry(c.Request().Context()).Errorf("Error %v", err)
		return http.StatusInternalServerError
	}

	// we send "stop-inviting-for-userPrincipalDto.UserId-signal" to the ownerId (call's owner)
	err = vh.dialStatusPublisher.Publish(chatId, usersOfDial, false, ownerId)

	if err != nil {
		logger.GetLogEntry(c.Request().Context()).Errorf("Error %v", err)
		return http.StatusInternalServerError
	}

	// send the new status immediately
	vh.chatInvitationService.SendInvitations(c.Request().Context(), chatId, ownerId, usersOfDial)

	return http.StatusOK
}

func(vh *InviteHandler) setUserStatus(ctx context.Context, callee, chatId int64, callStatus string) error {
	err := vh.dialRedisRepository.SetUserStatus(ctx, callee, callStatus)
	if err != nil {
		return err
	}
	if services.ShouldProlong(callStatus) {
		err = vh.dialRedisRepository.ResetExpiration(ctx, callee) // TODO provide mechanism for removing orphans - we run over all the rooms, and if user does not belong to any - we remove him
		if err != nil {
			return err
		}
	}
	if services.ShouldRemoveAutomaticallyAfterTimeout(callStatus) {
		err = vh.dialRedisRepository.SetCurrentTimeForCancellation(ctx, callee)
		if err != nil {
			return err
		}
	}
	return err
}

// owner stops call by exiting
func (vh *InviteHandler) ProcessLeave(c echo.Context) error {
	var userPrincipalDto, ok = c.Get(utils.USER_PRINCIPAL_DTO).(*auth.AuthResult)
	if !ok {
		Logger.Errorf("Error during getting auth context")
		return errors.New("Error during getting auth context")
	}

	chatId, err := GetPathParamAsInt64(c, "id")
	if err != nil {
		return err
	}

	// check my access to chat
	if ok, err := vh.chatClient.CheckAccess(userPrincipalDto.UserId, chatId, c.Request().Context()); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	} else if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	ownerId, err := vh.dialRedisRepository.GetDialMetadata(c.Request().Context(), chatId)
	if err != nil {
		logger.GetLogEntry(c.Request().Context()).Errorf("Error %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if ownerId == services.NoUser {
		return c.NoContent(http.StatusOK)
	}

	if ownerId != userPrincipalDto.UserId {
		return c.NoContent(http.StatusOK)
	}

	usersToDial, err := vh.dialRedisRepository.GetUsersToDial(c.Request().Context(), chatId)
	if err != nil {
		logger.GetLogEntry(c.Request().Context()).Errorf("Error %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// the owner removes all the dials by setting status
	vh.removeFromCallingList(c, chatId, usersToDial, services.CallStatusRemoving)

	// for all participants to dial - send EventMissedCall notification
	vh.sendMissedCallNotification(chatId, c.Request().Context(), userPrincipalDto, usersToDial)

	return c.NoContent(http.StatusOK)
}

func (vh *InviteHandler) sendMissedCallNotification(chatId int64, ctx context.Context, userPrincipalDto *auth.AuthResult, usersToDial []int64) {
	if len(usersToDial) > 0 {
		if chatNames, err := vh.chatClient.GetChatNameForInvite(chatId, userPrincipalDto.UserId, usersToDial, ctx); err != nil {
			logger.GetLogEntry(ctx).Errorf("Error %v", err)
		} else {
			for _, chatName := range chatNames {
				// here send missed call notification
				var missedCall = dto.NotificationEvent{
					EventType:              EventMissedCall,
					ChatId:                 chatId,
					UserId:                 chatName.UserId,
					MissedCallNotification: &dto.MissedCallNotification{chatName.Name},
					ByUserId:               userPrincipalDto.UserId,
					ByLogin:                userPrincipalDto.UserLogin,
				}
				err = vh.notificationPublisher.Publish(missedCall)
				if err != nil {
					logger.GetLogEntry(ctx).Errorf("Error %v", err)
				}
			}
		}
	}
}

// send current dial statuses to WebSocket
func (vh *InviteHandler) SendDialStatusChangedToCallOwner(c echo.Context) error {
	var userPrincipalDto, ok = c.Get(utils.USER_PRINCIPAL_DTO).(*auth.AuthResult)
	if !ok {
		Logger.Errorf("Error during getting auth context")
		return errors.New("Error during getting auth context")
	}

	chatId, err := GetPathParamAsInt64(c, "id")
	if err != nil {
		return err
	}

	// check my access to chat
	if ok, err := vh.chatClient.CheckAccess(userPrincipalDto.UserId, chatId, c.Request().Context()); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	} else if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	userIdsToDial, err := vh.dialRedisRepository.GetUsersToDial(c.Request().Context(), chatId)
	if err != nil {
		Logger.Warnf("Error %v", err)
		return c.NoContent(http.StatusOK)
	}

	err = vh.dialStatusPublisher.Publish(chatId, userIdsToDial, true, userPrincipalDto.UserId)
	if err != nil {
		Logger.Error(err, "Failed during marshal VideoIsInvitingDto")
		return c.NoContent(http.StatusOK)
	}

	return c.NoContent(http.StatusOK)
}
