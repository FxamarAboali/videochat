import {computed, reactive, nextTick} from "vue";
import {hasLength} from "#root/renderer/utils";
import {isTopDirection, reloadItems, infiniteScrollData} from "./infiniteScrollMixin.js";

// expects methods: doDefaultScroll(), getPositionFromStore(). isTopDirection() - from infiniteScrollMixin.js

const hashMixinData = reactive({
    startingFromItemIdTop: null,
    startingFromItemIdBottom: null,
    // those two doesn't play in reset() in order to survive after reload()
    hasInitialHash: false, // do we have hash in address line (message id)
    loadedHash: null, // keeps loaded message id from localstore the most top visible message - preserves scroll between page reload or switching between chats
})

const highlightItemId = computed(()=>{
    return null // return this.getIdFromRouteHash(this.$route.hash); // TODO
})

function getDefaultItemId() {
    return isTopDirection() ? hashMixinData.startingFromItemIdTop : hashMixinData.startingFromItemIdBottom;
}

function setHashes() {
    hashMixinData.hasInitialHash = hasLength(highlightItemId);
    this.loadedHash = getPositionFromStore();
}

function prepareHashesForLoad() {
    let startingFromItemId;
    let hasHash;
    if (hashMixinData.hasInitialHash) { // we need it here - it shouldn't be computable in order to be reset. The resetted value is need when we press "arrow down" after reload
        // how to check:
        // 1. click on hash
        // 2. reload page
        // 3. press "arrow down" (Scroll down)
        // 4. It is going to invoke this load method which will use cashed and reset hasInitialHash = false
        startingFromItemId = highlightItemId;
        hasHash = true;
    } else if (hashMixinData.loadedHash) {
        startingFromItemId = hashMixinData.loadedHash;
        hasHash = true;
    } else {
        startingFromItemId = getDefaultItemId();
        hasHash = false;
    }
    return {startingFromItemId, hasHash}
}

async function scrollTo(newValue) {
    await nextTick();
    const el = document.querySelector(newValue);
    el?.scrollIntoView({behavior: 'instant', block: "start"});
    return el
}

async function doScrollOnFirstLoad(prefix) {
    if (highlightItemId) {
        await scrollTo(prefix + highlightItemId);
    } else if (hashMixinData.loadedHash) {
        await scrollTo(prefix + hashMixinData.loadedHash);
    } else {
        await doDefaultScroll(); // we need it to prevent browser's scrolling
    }
    hashMixinData.loadedHash = null;
    hashMixinData.hasInitialHash = false;
}

async function setHashAndReloadItems(skipResetting) {
    setHashes();
    await reloadItems(skipResetting);
}

async function scrollToOrLoad(newValue) {
    const res = await scrollTo(newValue);
    if (!res) {
        console.log("Didn't scrolled, resetting");
        await setHashAndReloadItems();
    }
}

function getMaximumItemId() {
    return infiniteScrollData.items.length ? Math.max(...infiniteScrollData.items.map(it => it.id)) : null
}
function getMinimumItemId() {
    return infiniteScrollData.items.length ? Math.min(...infiniteScrollData.items.map(it => it.id)) : null
}
function clearRouteHash() {
    // console.log("Cleaning hash");
    // this.$router.push({ hash: null, query: this.$route.query }) // TODO
}

export {
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
}

