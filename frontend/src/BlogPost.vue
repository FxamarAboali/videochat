<template>
    <div>
        <h1>Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. </h1>

        <div class="pr-1 mr-1 pl-1 mt-0 message-item-root" >
            <div class="message-item-with-buttons-wrapper">
                <v-list-item class="grow">
                    <v-list-item-avatar color="grey darken-3">
                        <v-img
                            class="elevation-6"
                            alt=""
                            src="https://avataaars.io/?avatarStyle=Transparent&topType=ShortHairShortCurly&accessoriesType=Prescription02&hairColor=Black&facialHairType=Blank&clotheType=Hoodie&clotheColor=White&eyeType=Default&eyebrowType=DefaultNatural&mouthType=Default&skinColor=Light"
                        ></v-img>
                    </v-list-item-avatar>

                    <v-list-item-content>
                        <v-list-item-title>Nikita Konev</v-list-item-title><v-list-item-title>at 2022-12-30</v-list-item-title>
                    </v-list-item-content>
                </v-list-item>


                <div class="pa-0 ma-0 mt-1 message-item-wrapper my">
                    <v-container v-html="text" class="message-item-text ml-0"></v-container>
                </div>
            </div>
        </div>

        <v-list>
            <!-- TODO MessageList.vue -->
            <template v-for="(item, index) in items">
                <MessageItem
                    :key="item.id"
                    :item="item"
                    :chatId="chatId"
                    :my="item.my"
                ></MessageItem>
            </template>
        </v-list>
    </div>
</template>

<script>
    import MessageItem from "@/MessageItem";

    export default {
        data() {
            return {
                chatId: 34,
                text: "<p>Lorem upsum</p><p>dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>" +
                      "<p>Lorem upsum</p><p>dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.</p>",
                items: [
                    {
                        "id": 2822,
                        "text": "<p>Testw</p>",
                        "chatId": 34,
                        "ownerId": 1,
                        "createDateTime": "2023-04-13T16:06:21.569617Z",
                        "editDateTime": "2023-05-06T01:02:18.329121Z",
                        "owner": {
                            "id": 1,
                            "login": "admin",
                            "avatar": "/api/storage/public/user/avatar/1_AVATAR_200x200.jpg?time=1676930657",
                            "shortInfo": "Admin account"
                        },
                        "canEdit": true,
                        "canDelete": true,
                        "fileItemUuid": "17b13878-2a0d-4860-a6d6-2ad8545901b0",
                        "embedMessage": null,
                        "pinned": false,
                        "my": false,
                    },
                    {
                        "id": 2823,
                        "text": "<p>Lorem</p>",
                        "chatId": 34,
                        "ownerId": 1,
                        "createDateTime": "2023-04-13T16:06:21.569617Z",
                        "editDateTime": "2023-05-06T01:02:18.329121Z",
                        "owner": {
                            "id": 1,
                            "login": "admin",
                            "avatar": "/api/storage/public/user/avatar/1_AVATAR_200x200.jpg?time=1676930657",
                            "shortInfo": "Admin account"
                        },
                        "canEdit": true,
                        "canDelete": true,
                        "fileItemUuid": "17b13878-2a0d-4860-a6d6-2ad8545901b0",
                        "embedMessage": null,
                        "pinned": false,
                        "my": false,
                    },
                    {
                        "id": 2824,
                        "text": "<b>Ipsum</b>",
                        "chatId": 34,
                        "ownerId": 1,
                        "createDateTime": "2023-04-13T16:06:21.569617Z",
                        "editDateTime": "2023-05-06T01:02:18.329121Z",
                        "owner": {
                            "id": 1,
                            "login": "admin",
                            "avatar": "/api/storage/public/user/avatar/1_AVATAR_200x200.jpg?time=1676930657",
                            "shortInfo": "Admin account"
                        },
                        "canEdit": true,
                        "canDelete": true,
                        "fileItemUuid": "17b13878-2a0d-4860-a6d6-2ad8545901b0",
                        "embedMessage": null,
                        "pinned": false,
                        "my": true,
                    },
                    {
                        "id": 2825,
                        "text": "<p>Dolor</p>",
                        "chatId": 34,
                        "ownerId": 1,
                        "createDateTime": "2023-04-13T16:06:21.569617Z",
                        "editDateTime": "2023-05-06T01:02:18.329121Z",
                        "owner": {
                            "id": 1,
                            "login": "admin",
                            "avatar": "/api/storage/public/user/avatar/1_AVATAR_200x200.jpg?time=1676930657",
                            "shortInfo": "Admin account"
                        },
                        "canEdit": true,
                        "canDelete": true,
                        "fileItemUuid": "17b13878-2a0d-4860-a6d6-2ad8545901b0",
                        "embedMessage": null,
                        "pinned": false,
                        "my": true,
                    },
                ]
            }
        },
        components: {
            MessageItem
        }
    }
</script>

<style lang="stylus">
    @import "common.styl"
</style>