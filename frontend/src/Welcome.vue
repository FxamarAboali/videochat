<template>
    <v-container fill-height fluid>
        <v-row align="center" justify="center">
            <v-card>
                <v-card-title class="d-flex justify-space-around">{{$vuetify.lang.t('$vuetify.welcome_participant', currentUser.login)}}</v-card-title>
                <v-card-actions  class="d-flex justify-space-around">
                    <v-btn color="primary" @click="createChat()" text>
                        <v-icon>mdi-plus</v-icon>
                        {{ $vuetify.lang.t('$vuetify.new_chat') }}
                    </v-btn>
                    <v-btn @click="findUser()" text>
                        <v-icon>mdi-magnify</v-icon>
                        {{ $vuetify.lang.t('$vuetify.find_user') }}
                    </v-btn>
                    <v-btn @click="availableChats()" text>
                        <v-icon>mdi-forum</v-icon>
                        {{ $vuetify.lang.t('$vuetify.public_chats') }}
                    </v-btn>
                </v-card-actions>
            </v-card>
        </v-row>
    </v-container>
</template>

<script>
import {GET_USER, SET_SEARCH_STRING} from "@/store";
    import {mapGetters} from "vuex";
    import bus, {OPEN_CHAT_EDIT, OPEN_FIND_USER} from "@/bus";

    export default {
        computed: {
            ...mapGetters({currentUser: GET_USER}),
        },
        methods: {
            createChat() {
                bus.$emit(OPEN_CHAT_EDIT, null);
            },
            findUser() {
                bus.$emit(OPEN_FIND_USER)
            },
            availableChats() {
                this.$store.commit(SET_SEARCH_STRING, '_');
            },
        }
    }
</script>