<template>
    <v-row justify="center">
        <v-dialog v-model="show" max-width="800" height="100%" scrollable>
          <v-card :title="title()">
                <v-card-text>
                    <v-row
                      dense
                      v-if="!loading"
                      align="center"
                      justify="start"
                    >
                        <template
                          v-if="dto.count > 0"
                          v-for="(mediaFile, i) in dto.files"
                          :key="mediaFile.id"
                        >
                            <v-col :cols="isMobile() ? 12 : 6">
                                <v-hover v-slot="{ isHovering, props }">

                                    <v-card v-bind="props" @click="accept(mediaFile)">

                                            <v-img
                                                :src="mediaFile.previewUrl"
                                                gradient="to bottom, rgba(0,0,0,.1), rgba(0,0,0,.5)"
                                                class="align-end"
                                                height="200px"
                                                cover
                                            >
                                                <v-card-title class="card-title-wrapper">
                                                    <span v-text="mediaFile.filename" class="file-title text-white"></span>
                                                </v-card-title>
                                            </v-img>

                                            <!-- Even transition="false" doesn't actually disable the transition, it fixes breakage of the markup of hover -->
                                            <v-overlay
                                                :model-value="isHovering"
                                                :transition="false"
                                                contained
                                                class="align-center justify-center cursor-pointer"
                                            >
                                                <div class="text-white">
                                                    {{ $vuetify.locale.t('$vuetify.click_to_choose') }}
                                                </div>
                                            </v-overlay>

                                    </v-card>
                                </v-hover>
                            </v-col>
                        </template>
                        <template v-else>
                            <v-card-text>{{ $vuetify.locale.t('$vuetify.no_files') }}</v-card-text>
                        </template>
                    </v-row>

                    <v-progress-circular
                        class="my-4"
                        v-else
                        indeterminate
                        color="primary"
                    ></v-progress-circular>
                </v-card-text>

                <v-card-actions class="d-flex flex-wrap flex-row">

                  <!-- Pagination is shuddering / flickering on the second page without this wrapper -->
                  <v-row no-gutters class="ma-0 pa-0 d-flex flex-row">
                      <v-col class="ma-0 pa-0 flex-grow-1 flex-shrink-0" :class="isMobile() ? 'mb-2' : ''">
                          <v-pagination
                              variant="elevated"
                              active-color="primary"
                              density="comfortable"
                              v-if="shouldShowPagination"
                              v-model="page"
                              :length="pagesCount"
                              :total-visible="getTotalVisible()"
                          ></v-pagination>
                      </v-col>
                      <v-col class="ma-0 pa-0 d-flex flex-row flex-grow-1 flex-shrink-0 align-self-end justify-end">
                          <v-btn variant="outlined" @click="fromUrl()" min-width="0" :title="$vuetify.locale.t('$vuetify.from_link')"><v-icon size="large">mdi-link-variant</v-icon></v-btn>
                          <v-btn color="primary" variant="flat" @click="fromDisk()"><v-icon color="white">mdi-file-upload</v-icon>{{ $vuetify.locale.t('$vuetify.choose_file_from_disk') }}</v-btn>
                          <v-btn color="red" variant="flat" @click="closeModal()">{{ $vuetify.locale.t('$vuetify.close') }}</v-btn>
                      </v-col>
                  </v-row>
                </v-card-actions>

          </v-card>
        </v-dialog>
    </v-row>
</template>

<script>
    import axios from "axios";

    import bus, {
        FILE_CREATED, FILE_REMOVED,
        FILE_UPDATED,
        LOGGED_OUT,
        OPEN_MESSAGE_EDIT_LINK,
        OPEN_MESSAGE_EDIT_MEDIA, PREVIEW_CREATED
    } from "./bus/bus";
    import {
        findIndex,
        link_dialog_type_add_media_by_link,
        media_audio,
        media_image,
        media_video,
        replaceOrPrepend
    } from "@/utils";

    const firstPage = 1;
    const pageSize = 20;

    const dtoFactory = () => {return {files: []} };

    export default {
        data () {
            return {
                show: false,
                type: '',
                fromDiskCallback: null,
                setExistingMediaCallback: null,
                loading: false,
                dto: dtoFactory(),
                page: firstPage,
                dataLoaded: false,
                isDirty: false,
            }
        },
        computed: {
            pagesCount() {
                const count = Math.ceil(this.dto.count / pageSize);
                // console.debug("Calc pages count", count);
                return count;
            },
            shouldShowPagination() {
                return this.dto != null && this.dto.files && this.dto.count > pageSize
            },
            chatId() {
                return this.$route.params.id
            },
        },
        methods: {
            showModal({type, fromDiskCallback, setExistingMediaCallback}) {

                if (this.type != type) {
                    this.reset();
                }

                if (this.isDirty) {
                    this.reset();
                    this.isDirty = false;
                }

                this.$data.show = true;

                this.type = type;
                this.fromDiskCallback = fromDiskCallback;
                this.setExistingMediaCallback = setExistingMediaCallback;

                if (!this.dataLoaded) {
                    this.updateFiles();
                } else {
                    //
                }
            },
            accept(item) {
                if (this.setExistingMediaCallback) {
                    this.setExistingMediaCallback(item.url, item.previewUrl)
                }
                this.closeModal();
            },
            clear() {
                this.closeModal();
            },
            title() {
                switch (this.type) {
                    case media_video:
                        return this.$vuetify.locale.t('$vuetify.message_edit_video')
                    case media_image:
                        return this.$vuetify.locale.t('$vuetify.message_edit_image')
                    case media_audio:
                        return this.$vuetify.locale.t('$vuetify.message_edit_audio')
                }
            },
            fromUrl() {
                bus.emit(OPEN_MESSAGE_EDIT_LINK, {dialogType: link_dialog_type_add_media_by_link, mediaType: this.type});
                this.closeModal();
            },
            fromDisk() {
                if (this.fromDiskCallback) {
                    this.fromDiskCallback();
                }
                this.closeModal();
            },
            translatePage() {
                return this.page - 1;
            },
            updateFiles() {
                if (!this.show) {
                    return
                }
                this.loading = true;
                axios.get(`/api/storage/${this.chatId}/embed/candidates`, {
                    params: {
                        page: this.translatePage(),
                        size: pageSize,
                        type: this.type
                    },
                })
                    .then((response) => {
                        this.dto = response.data;
                    })
                    .finally(() => {
                        this.loading = false;
                        this.dataLoaded = true;
                    })
            },
            getTotalVisible() {
                if (!this.isMobile()) {
                    return 7
                } else if (this.page == firstPage || this.page == this.pagesCount) {
                    return 3
                } else {
                    return 1
                }
            },

            removeItem(dto) {
                console.log("Removing item", dto);
                const idxToRemove = findIndex(this.dto.files, dto);
                this.dto.files.splice(idxToRemove, 1);
            },
            replaceItem(dto) {
                console.log("Replacing item", dto);
                replaceOrPrepend(this.dto.files, [dto]);
            },
            onFileCreated(dto) {
                if (!this.dataLoaded) {
                    return
                }
                if (!this.show) {
                    this.reset()
                } else {
                    this.isDirty = true;
                }
            },
            onFileUpdated(dto) {
                if (!this.dataLoaded) {
                    return
                }
                if (!this.show) {
                    this.reset()
                } else {
                    this.isDirty = true;
                }
            },
            onFileRemoved(dto) {
                if (!this.dataLoaded) {
                    return
                }
                if (!this.show) {
                    this.reset()
                } else {
                    this.isDirty = true;
                }
            },
            onLogout() {
                this.reset();
                this.closeModal();
            },
            closeModal() {
                this.show = false;
                this.fromDiskCallback = null;
                this.setExistingMediaCallback = null;
            },
            reset() {
                this.type = '';
                this.dataLoaded = false;
                this.loading = false;
                this.dto = dtoFactory();
                this.page = firstPage;
            },
        },
        watch: {
            show(newValue) {
                if (!newValue) {
                    this.closeModal();
                }
            },
            page(newValue) {
                if (this.show) {
                    console.debug("SettingNewPage", newValue);
                    this.dto = dtoFactory();
                    this.updateFiles();
                }
            },
            '$route.params.id': function (newValue, oldValue) {
                if (newValue != oldValue) {
                    this.reset();
                }
            }
        },
        mounted() {
            bus.on(OPEN_MESSAGE_EDIT_MEDIA, this.showModal);
            bus.on(FILE_CREATED, this.onFileCreated);
            bus.on(FILE_UPDATED, this.onFileUpdated);
            bus.on(FILE_REMOVED, this.onFileRemoved);
            bus.on(LOGGED_OUT, this.onLogout);
        },
        beforeUnmount() {
            bus.off(OPEN_MESSAGE_EDIT_MEDIA, this.showModal);
            bus.off(FILE_CREATED, this.onFileCreated);
            bus.off(FILE_UPDATED, this.onFileUpdated);
            bus.off(FILE_REMOVED, this.onFileRemoved);
            bus.off(LOGGED_OUT, this.onLogout);
        },
    }
</script>

<style lang="stylus" scoped>
  .breaks {
    white-space: break-spaces;
  }
  .cursor-pointer {
    cursor pointer
  }

  .card-title-wrapper {
    line-height 1.25em

    .file-title {
      white-space break-spaces
    }
  }

</style>
