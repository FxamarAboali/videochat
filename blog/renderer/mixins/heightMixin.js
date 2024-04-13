import { computed } from 'vue';
import {isMobileBrowser} from "../utils.js";

const heightWithoutAppBar = computed(()=>{
    if (isMobileBrowser()) {
        return 'height: calc(100dvh - 56px)'
    } else {
        return 'height: calc(100dvh - 48px)'
    }
})

export {heightWithoutAppBar}
