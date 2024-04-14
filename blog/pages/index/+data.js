import axios from "axios";
import { PAGE_SIZE, getApiHost, SEARCH_MODE_POSTS } from "#root/renderer/utils";
import {infiniteScrollData} from "#root/renderer/mixins/hashMixin";

export { data };

async function data(pageContext) {
    const apiHost = getApiHost();
    const searchString = pageContext.urlParsed.search[SEARCH_MODE_POSTS];
    const response = await axios.get(apiHost + '/api/blog', {
        params: {
            size: PAGE_SIZE,
            reverse: false,
            searchString: searchString,
            hasHash: false,
        },
    });

    infiniteScrollData.items = response.data;

    return {
        ...infiniteScrollData,
        markInstance: null,
    }
}
