import debounce from "lodash/debounce";
import {reactive, nextTick, inject} from "vue";

const directionTop = 'top';
const directionBottom = 'bottom';

export const initialDirectionSymbol = Symbol();

// expects getMaxItemsLength(),
// bottomElementSelector(), topElementSelector(), getItemId(id),
// load(), onFirstLoad(), initialDirection(), saveScroll(), scrollerSelector(),
// reduceTop(), reduceBottom()
// onScrollCallback(), afterScrollRestored()
// onScroll() should be called from template

const initialDirection = inject(initialDirectionSymbol);

const infiniteScrollData = reactive(
    {
        items: [],
        observer: null,

        isFirstLoad: true,

        scrollerDiv: null,

        loadedTop: false,
        loadedBottom: false,

        aDirection: initialDirection(),

        scrollerProbeCurrent: 0,
        scrollerProbePrevious: 0,

        preservedScroll: 0,
        timeout: null,
    }
);

function cssStr(el) {
    return el.tagName.toLowerCase() + (el.id ? '#' + el.id : "") + '.' + (Array.from(el.classList)).join('.')
}

async function reduceListIfNeed() {
    if (infiniteScrollData.items.length > getMaxItemsLength()) {
        await nextTick();
        if (isTopDirection()) {
            reduceBottom();
            infiniteScrollData.loadedBottom = false;
        } else {
            reduceTop();
            infiniteScrollData.loadedTop = false;
        }
        console.log("Reduced to", getMaxItemsLength(), infiniteScrollData.loadedBottom, infiniteScrollData.loadedTop);
    }
}

function isTopDirection() {
    return infiniteScrollData.aDirection === directionTop
}

function trySwitchDirection() {
    if (infiniteScrollData.scrollerProbeCurrent != 0 && infiniteScrollData.scrollerProbeCurrent > infiniteScrollData.scrollerProbePrevious && isTopDirection()) {
        infiniteScrollData.aDirection = directionBottom;
        // console.debug("Infinity scrolling direction has been changed to bottom");
    } else if (infiniteScrollData.scrollerProbeCurrent != 0 && infiniteScrollData.scrollerProbePrevious > infiniteScrollData.scrollerProbeCurrent && !isTopDirection()) {
        infiniteScrollData.aDirection = directionTop;
        // console.debug("Infinity scrolling direction has been changed to top");
    } else {
        // console.debug("Infinity scrolling direction has been remained untouched", this.aDirection);
    }
}

function onScroll(e) {
    if (onScrollCallback) {
        onScrollCallback();
    }

    infiniteScrollData.scrollerProbePrevious = infiniteScrollData.scrollerProbeCurrent;
    infiniteScrollData.scrollerProbeCurrent = infiniteScrollData.scrollerDiv.scrollTop;
    // console.debug("onScroll in", name, " prev=", this.scrollerProbePrevious, "cur=", this.scrollerProbeCurrent);

    trySwitchDirection();
}

function restoreScroll(top) {
    const restored = infiniteScrollData.preservedScroll;
    const q = scrollerSelector() + " " + "#"+getItemId(restored);
    const el = document.querySelector(q);
    console.debug("Restored scroll to element id", restored, "selector", q, "element", el);
    el?.scrollIntoView({behavior: 'instant', block: top ? "start": "end"});
    if (afterScrollRestored) {
        afterScrollRestored(el)
    }
}

function resetInfiniteScrollVars(skipResetting) {
    if (!skipResetting) {
        infiniteScrollData.items = [];
    }
    infiniteScrollData.isFirstLoad = true;
    infiniteScrollData.loadedTop = false;
    infiniteScrollData.loadedBottom = false;
    infiniteScrollData.aDirection = initialDirection();
    infiniteScrollData.scrollerProbePrevious = 0;
    infiniteScrollData.scrollerProbeCurrent = 0;
    infiniteScrollData.preservedScroll = null;
}

async function initialLoad() {
    if (infiniteScrollData.scrollerDiv == null) {
        infiniteScrollData.scrollerDiv = document.querySelector(scrollerSelector());
    }
    const loadedResult = await load();
    await nextTick();
    await onFirstLoad(loadedResult);
    infiniteScrollData.isFirstLoad = false;
}

async function loadTop() {
    console.log("going to load top");
    saveScroll(true); // saves scroll between new portion load
    await load(); // restores scroll after new portion load
    await nextTick();
    await reduceListIfNeed();
    restoreScroll(true);
}

async function loadBottom() {
    console.log("going to load bottom");
    saveScroll(false);
    await load();
    await nextTick();
    await reduceListIfNeed();
    restoreScroll(false);
}

function isReady() {
    return infiniteScrollData.scrollerDiv != null
}

function initScroller() {
    if (!isReady()) {
        throw "You have to invoke initialLoad() first"
    }

    // https://developer.mozilla.org/en-US/docs/Web/API/Intersection_Observer_API
    const options = {
        root: infiniteScrollData.scrollerDiv,
        rootMargin: "0px",
        threshold: 0.0,
    };
    const observerCallback0 = async (entries, observer) => {
        const mappedEntries = entries.map((entry) => {
            return {
                entry,
                elementName: cssStr(entry.target)
            }
        });
        const lastElementEntries = mappedEntries.filter(en => en.entry.intersectionRatio > 0 && en.elementName.includes(topElementSelector()));
        const lastElementEntry = lastElementEntries.length ? lastElementEntries[lastElementEntries.length-1] : null;

        const firstElementEntries = mappedEntries.filter(en => en.entry.intersectionRatio > 0 && en.elementName.includes(bottomElementSelector()));
        const firstElementEntry = firstElementEntries.length ? firstElementEntries[firstElementEntries.length-1] : null;

        console.log("Invoking callback", mappedEntries);

        if (lastElementEntry && lastElementEntry.entry.isIntersecting) {
            console.debug("attempting to load top", !infiniteScrollData.loadedTop, isTopDirection());
            if (!infiniteScrollData.loadedTop && isTopDirection()) {
                await loadTop();
            }
        }
        if (firstElementEntry && firstElementEntry.entry.isIntersecting) {
            console.debug("attempting to load bottom", !infiniteScrollData.loadedBottom, !isTopDirection());
            if (!infiniteScrollData.loadedBottom && !isTopDirection()) {
                await loadBottom();
            }
        }
    };

    const observerCallback = debounce(observerCallback0, 200, {leading:false, trailing:true});

    infiniteScrollData.observer = new IntersectionObserver(observerCallback, options);
    infiniteScrollData.observer.observe(document.querySelector(scrollerSelector() + " " + bottomElementSelector()));
    infiniteScrollData.observer.observe(document.querySelector(scrollerSelector() + " " + topElementSelector()));
}

function destroyScroller() {
    infiniteScrollData.observer?.disconnect();
    infiniteScrollData.observer = null;
    infiniteScrollData.scrollerDiv = null;
}

function installScroller() {
    infiniteScrollData.timeout = setTimeout(async ()=> {
        await nextTick();
        initScroller();
        console.log("Scroller has been installed");
        infiniteScrollData.timeout = null;
    }, 1500); // must be > than debounce millis in observer (it seems this strange behavior can be explained by optimizations in Firefox)
    // tests in Firefox
    // a) refresh page 30 times
    // b) refresh page 30 times when the hash is present (#message-523)
    // c) input search string - search by messages
}

function uninstallScroller(skipResetting) {
    if (infiniteScrollData.timeout) {
        clearTimeout(infiniteScrollData.timeout);
        infiniteScrollData.timeout = null;
    }
    destroyScroller();
    reset(skipResetting);
    console.log("Scroller has been uninstalled");
}

async function reloadItems(skipResetting) {
    uninstallScroller(skipResetting);
    await initialLoad();
    await nextTick();
    installScroller();
}

export {
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
}
