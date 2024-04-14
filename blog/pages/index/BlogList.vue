<template>
  <v-container class="ma-0 pa-0" :style="heightWithoutAppBar" fluid>
    <div class="my-blog-scroller" id="blog-post-list" @scroll.passive="onScroll">
      <div class="blog-first-element" style="min-height: 1px; background: white"></div>

      <v-card
        v-for="(item, index) in items"
        :key="item.id"
        :id="getItemId(item.id)"
        class="mb-2 mr-2 blog-item-root"
        :min-width="isMobile() ? 200 : 400"
        max-width="600"
      >
        <v-card-text class="pb-0">
          <v-card>
            <v-img
              class="text-white align-end"
              gradient="to bottom, rgba(0,0,0,.1), rgba(0,0,0,.5)"
              cover
              :height="isMobile() ? 200 : 300"
              :src="item.imageUrl"
            >
              <v-container class="post-title ma-0 pa-0">
                <v-card-title @click.prevent="goToBlog(item)">
                  <a class="post-title-text" v-html="item.title" :href="getLink(item)"></a>
                </v-card-title>
              </v-container>
            </v-img>
          </v-card>
        </v-card-text>

        <v-card-text class="post-text pb-0" v-html="item.preview">
        </v-card-text>

        <v-card-actions v-if="item?.owner != null">
          <v-list-item>
              <template v-slot:prepend v-if="hasLength(item?.owner?.avatar)">
                  <div class="item-avatar pr-0 mr-3">
                      <a :href="getProfileLink(item.owner)" class="user-link">
                          <img :src="item?.owner?.avatar">
                      </a>
                  </div>
              </template>

              <template v-slot:default>
                  <v-list-item-title><a :href="getProfileLink(item.owner)" class="colored-link">{{ item?.owner?.login }}</a></v-list-item-title>
                  <v-list-item-subtitle>
                      {{ getDate(item) }}
                  </v-list-item-subtitle>

              </template>

          </v-list-item>
        </v-card-actions>
      </v-card>
      <div class="blog-last-element" style="min-height: 1px; background: white"></div>

    </div>

  </v-container>
</template>

<script setup>
import {getHumanReadableDate, hasLength, replaceOrAppend, replaceOrPrepend, setTitle} from "#root/renderer/utils";
import axios from "axios";
import debounce from "lodash/debounce";
import Mark from "mark.js";
import {blog_post, blog_post_name, blogIdPrefix, blogIdHashPrefix, profile} from "#root/renderer/router/routes";
import {
    infiniteScrollData,
    cssStr,
    reduceListIfNeed,
    onScroll,
    trySwitchDirection,
    isTopDirection,
    restoreScroll,
    resetInfiniteScrollVars,
    initialLoad,
    loadTop,
    loadBottom,
    isReady,
    initScroller,
    destroyScroller,
    installScroller,
    uninstallScroller,
    reloadItems,
    directionTop,
    directionBottom
} from "#root/renderer/mixins/infiniteScrollMixin";
// import {mapStores} from "pinia";
// import {useBlogStore} from "@/store/blogStore";
// TODO
// import {goToPreservingQuery, SEARCH_MODE_POSTS, searchString} from "@/mixins/searchString";
// import bus, {SEARCH_STRING_CHANGED} from "@/bus/bus"; // TODO
import { heightWithoutAppBar } from "#root/renderer/mixins/heightMixin";
import {
    hashMixinData,
    highlightItemId,
    getDefaultItemId,
    setHashes,
    prepareHashesForLoad,
    doScrollOnFirstLoad,
    scrollTo,
    scrollToOrLoad,
    setHashAndReloadItems,
    getMaximumItemId,
    getMinimumItemId,
    clearRouteHash,
} from "#root/renderer/mixins/hashMixin";
import {
    getTopBlogPosition,
    removeTopBlogPosition,
    setTopBlogPosition,
} from "#root/renderer/store/localStore";
import {isMobileBrowser} from "#root/renderer/utils.js";
import { getData, useData } from '#root/renderer/useData';
import {onMounted, onBeforeUnmount, nextTick} from "vue";
import {useLocale} from "vuetify";

const { t } = useLocale();

const PAGE_SIZE = 40;
const SCROLLING_THRESHHOLD = 200; // px

const scrollerName = 'BlogList';

const data = useData(); // + hashMixinData


function getMaxItemsLength() {
    return 240
}
function getReduceToLength() {
    return 80 // in case numeric pages, should complement with getMaxItemsLength() and PAGE_SIZE
}
function reduceBottom() {
    data.items = data.items.slice(0, getReduceToLength());
    hashMixinData.startingFromItemIdBottom = getMaximumItemId();
}
function reduceTop() {
    data.items = data.items.slice(-getReduceToLength());
    hashMixinData.startingFromItemIdTop = getMinimumItemId();
}
function initialDirection() {
    return directionBottom
}
function saveScroll(top) {
    data.preservedScroll = top ? getMaximumItemId() : getMinimumItemId();
    console.log("Saved scroll", infiniteScrollData.preservedScroll, "in ", scrollerName);
}
async function scrollTop() {
    await nextTick();
    data.scrollerDiv.scrollTop = 0;
}
async function onFirstLoad(loadedResult) {
    await doScrollOnFirstLoad(blogIdHashPrefix);
    if (loadedResult === true) {
        removeTopBlogPosition();
    }
}
async function doDefaultScroll() {
    data.loadedTop = true;
    await scrollTop();
}
function getPositionFromStore() {
    return getTopBlogPosition()
}

async function load() {
    console.log("in load");
    if (!canDrawBlogs()) {
        return Promise.resolve()
    }

    if (data.items.length) {
        updateTopAndBottomIds();
        performMarking();
        return Promise.resolve()
    }

    // this.blogStore.incrementProgressCount(); // TODO
    const { startingFromItemId, hasHash } = prepareHashesForLoad();
    return axios.get(`/api/blog`, {
        params: {
            startingFromItemId: startingFromItemId,
            size: PAGE_SIZE,
            reverse: isTopDirection(),
            //searchString: this.searchString, // TODO get from PageShell.vue
            searchString: "",
            hasHash: hasHash,
        },
    })
        .then((res) => {
            const items = res.data;
            console.log("Get items in ", scrollerName, items, "page", hashMixinData.startingFromItemIdTop, hashMixinData.startingFromItemIdBottom);

            // replaceOrPrepend() and replaceOrAppend() for the situation when order has been changed on server,
            // e.g. some chat has been popped up on sever due to somebody updated it
            if (isTopDirection()) {
                replaceOrPrepend(data.items, items);
            } else {
                replaceOrAppend(data.items, items);
            }

            if (items.length < PAGE_SIZE) {
                if (isTopDirection()) {
                    data.loadedTop = true;
                } else {
                    data.loadedBottom = true;
                }
            }
            updateTopAndBottomIds();

            if (!data.isFirstLoad) {
                clearRouteHash()
            }

            performMarking();
            return Promise.resolve(true)
        }).finally(()=>{
            // this.blogStore.decrementProgressCount(); // TODO
        })
}
function canDrawBlogs() {
    return true
}

function bottomElementSelector() {
    return ".blog-last-element"
}
function topElementSelector() {
    return ".blog-first-element"
}

function getItemId(id) {
    return blogIdPrefix + id
}

function scrollerSelector() {
    return ".my-blog-scroller"
}
function reset(skipResetting) {
    resetInfiniteScrollVars(skipResetting);

    hashMixinData.startingFromItemIdTop = null;
    hashMixinData.startingFromItemIdBottom = null;
}

function getDate(item) {
    return getHumanReadableDate(item.createDateTime)
}

function performMarking() {
    // TODO
    // this.$nextTick(() => {
    //   if (hasLength(this.searchString)) {
    //     this.markInstance.unmark();
    //     this.markInstance.mark(this.searchString);
    //   }
    // })
}
function isScrolledToTop() {
    if (data.scrollerDiv) {
        return Math.abs(data.scrollerDiv.scrollTop) < SCROLLING_THRESHHOLD
    } else {
        return false
    }
}
function updateTopAndBottomIds() {
    hashMixinData.startingFromItemIdTop = getMaximumItemId();
    hashMixinData.startingFromItemIdBottom = getMinimumItemId();
}

function getProfileLink(user) {
    let url = profile + "/" + user.id;
    return url;
}
async function onSearchStringChanged() {
    // Fixes excess delayed (because of debounce) reloading of items when
    // 1. we've chosen __AVAILABLE_FOR_SEARCH
    // 2. then go to the Welcome
    // 3. without this change there will be excess delayed invocation
    // 4. but we've already destroyed this component, so it will be an error in the log
    if (isReady()) {
        await reloadItems();
    }
}
function setTopTitle() {
    setTitle(t('$vuetify.blogs'));
    // this.blogStore.title = this.$vuetify.locale.t('$vuetify.blogs'); // TODO
}
function goToBlog(item) {
    // TODO
    // goToPreservingQuery(this.$route, this.$router, { name: blog_post_name, params: { id: item.id} })
}
function getLink(item) {
    return blog_post + "/" + item.id
}
async function start() {
    await setHashAndReloadItems(true);
}

function saveLastVisibleElement() {
    console.log("saveLastVisibleElement", !isScrolledToTop())
    if (!isScrolledToTop()) {
        const elems = [...document.querySelectorAll(scrollerSelector() + " .blog-item-root")].map((item) => {
            const visible = item.getBoundingClientRect().top > 0
            return {item, visible}
        });

        const visible = elems.filter((el) => el.visible);
        // console.log("visible", visible, "elems", elems);
        if (visible.length == 0) {
            console.warn("Unable to get top visible")
            return
        }
        const topVisible = visible[0].item

        const bid = this.getIdFromRouteHash(topVisible.id);
        console.log("Found bottomPost", topVisible, "blogId", bid);

        setTopBlogPosition(bid)
    } else {
        console.log("Skipped saved topVisible because we are already scrolled to the bottom ")
    }
}
function beforeUnload() {
    saveLastVisibleElement();
}

onSearchStringChanged = debounce(onSearchStringChanged, 700, {leading:false, trailing:true});

// mounted
onMounted(async ()=>{
    data.markInstance = new Mark("div#blog-post-list");
    setTopTitle();
    // this.blogStore.searchType = SEARCH_MODE_POSTS; // TODO

    if (canDrawBlogs()) {
        await start();
    }
    addEventListener("beforeunload", beforeUnload);
})

// before unmount
onBeforeUnmount(()=>{
    // this.blogStore.isShowSearch = false; // TODO

    // an analogue of watch(effectively(chatId)) in MessageList.vue
    // used when the user presses Start in the RightPanel
    saveLastVisibleElement();

    data.markInstance.unmark();
    data.markInstance = null;
    removeEventListener("beforeunload", beforeUnload);

    uninstallScroller();
    // TODO
    // bus.off(SEARCH_STRING_CHANGED + '.' + SEARCH_MODE_POSTS, this.onSearchStringChanged);
})


// TODO
//   watch: {
// '$route': { // TODO check if working in vike
//     handler: async function (newValue, oldValue) {
//
//         // reaction on setting hash
//         if (hasLength(newValue.hash)) {
//             console.log("Changed route hash, going to scroll", newValue.hash)
//             await this.scrollToOrLoad(newValue.hash);
//             return
//         }
//     }
// }
// }

</script>

<style lang="stylus">
@import "../../renderer/styles/constants.styl"
@import "../../renderer/styles/itemAvatar.styl"

.my-blog-scroller {
  height 100%
  overflow-y scroll !important
  display flex
  flex-wrap wrap
  align-items start
}

.post-title {
  background rgba(0, 0, 0, 0.5);

  .post-title-text {
    cursor pointer
    color white
    text-decoration none
    word-break: break-word;
  }
}

.post-text {
    color $blackColor
}

.blog-item-root {
  flex: 1 1 300px;
}
.user-link {
    height 100%
}

</style>
