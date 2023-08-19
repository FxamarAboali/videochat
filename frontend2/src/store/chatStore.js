// Utilities
import { defineStore } from 'pinia'
import axios from "axios";
import {isMobileBrowser, setIcon} from "@/utils";
import {SEARCH_MODE_CHATS, SEARCH_MODE_MESSAGES} from "@/mixins/searchString";

export const useChatStore = defineStore('chat', {
  state: () => {
    return {
        currentUser: null,
        notifications: [],
        notificationsSettings: {},
        showCallButton: false,
        showHangButton: false,
        shouldPhoneBlink: false,
        availableOAuth2Providers: [],
        showAlert: false,
        lastError: "",
        errorColor: "",
        showDrawer: isMobileBrowser(),
        isShowSearch: true,
        searchType: SEARCH_MODE_CHATS,
        showScrollDown: false,
        title: "",
        avatar: null,
        chatUsersCount: 0,
        showChatEditButton: false,
        canBroadcastTextMessage: false,
        tetATet: false,
        showGoToBlogButton: null,
        videoChatUsersCount: 0,
        canMakeRecord: false,
        showRecordStartButton: false,
        showRecordStopButton: false,
    }
  },
  actions: {
    unsetUser() {
      this.currentUser = null
    },
    fetchUserProfile() {
        axios.get(`/api/profile`).then(( {data} ) => {
            console.debug("fetched profile =", data);
            this.currentUser = data;
        });
    },
    fetchAvailableOauth2Providers() {
          return axios.get(`/api/oauth2/providers`).then(( {data} ) => {
              console.debug("fetched oauth2 providers =", data);
              this.availableOAuth2Providers = data;
          });
    },
    fetchNotifications() {
      axios.get(`/api/notification/notification`).then(( {data} ) => {
        console.debug("fetched notifications =", data);
        this.notifications = data;
        setIcon(data != null && data.length > 0);
      });
      axios.get(`/api/notification/settings`).then(( {data} ) => {
        console.debug("fetched notifications settings =", data);
        this.notificationsSettings = data;
      });
    },
    unsetNotifications() {
      this.notifications = [];
      setIcon(false);
    },
    switchSearchType() {
      if (this.searchType == SEARCH_MODE_CHATS) {
        this.searchType = SEARCH_MODE_MESSAGES
      } else if (this.searchType == SEARCH_MODE_MESSAGES) {
        this.searchType = SEARCH_MODE_CHATS
      }
    }
  },

})
