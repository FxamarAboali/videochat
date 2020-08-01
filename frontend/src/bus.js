import Vue from 'vue';
export default new Vue();
export const UNAUTHORIZED  = 'unauthorized';
export const LOGGED_OUT = "loggedOut";
export const LOGGED_IN = "loggedIn";
export const CHAT_ADD = "chatAdd";
export const CHAT_EDITED = "chatEdited";
export const CHAT_DELETED = "chatDeleted";
export const OPEN_CHAT_EDIT = "openChatEdit";
export const OPEN_CHAT_DELETE = "openChatDelete";
export const CHAT_SEARCH_CHANGED = "chatSearchChanged";
export const CHANGE_TITLE = "changeTitle";
export const MESSAGE_ADD = "messageAdd";
export const MESSAGE_DELETED = "messageDeleted";
export const MESSAGE_EDITED = "messageEdited";
export const SET_EDIT_MESSAGE = "setEditMessageDto";
export const UNREAD_MESSAGES_CHANGED = "unreadMessagesChanged";
export const VIDEO_LOCAL_ESTABLISHED = "videoLocalEstablished"