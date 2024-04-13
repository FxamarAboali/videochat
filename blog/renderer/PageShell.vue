<template>
    <v-app>
        <v-app-bar color='indigo' dark :density="getDensity()">
            <v-breadcrumbs
                :items="getBreadcrumbs()"
            />
            <v-spacer></v-spacer>

            <template v-if="true">
                <CollapsedSearch :provider="getProvider()"/>
            </template>
        </v-app-bar>

        <v-main>
            <slot />
        </v-main>
    </v-app>
</template>

<script setup>
    import {hasLength, isMobileBrowser, SEARCH_MODE_POSTS} from "./utils.js";
    import {blog, root} from "./router/routes.js";
    import { navigate } from 'vike/client/router';
    import {usePageContext} from "./usePageContext.js";
    import CollapsedSearch from "./CollapsedSearch.vue";
    import { useData } from '#root/renderer/useData';
    import { computed } from 'vue';
    import { useLocale } from 'vuetify';

    const { t } = useLocale();

    const pageContext = usePageContext();
    const data = useData();

    const searchStringFacade = computed({
        get() {
            return pageContext.urlParsed.search[SEARCH_MODE_POSTS];
        },
        set(newVal) {
            if (hasLength(newVal)) {
                navigate(blog + '?' + SEARCH_MODE_POSTS + "=" + newVal)
            } else {
                navigate(blog)
            }
        }
    });

    function getDensity() {
        return isMobileBrowser() ? "comfortable" : "compact";
    }

    function getBreadcrumbs() {
        const ret = [
            {
                title: 'Videochat',
                disabled: false,
                href: root,
            },
            {
                title: 'Blog',
                disabled: false,
                exactPath: true,
                href: blog,
            },
        ];
        // if (this.$route.name == blog_post_name) {
        //     ret.push(
        //         {
        //             title: 'Post #' + this.$route.params.id,
        //             disabled: false,
        //             to: blog_post + "/" + this.$route.params.id,
        //         },
        //     )
        // }
        return ret
    }

    function getModelValue() {
        return searchStringFacade.value
    }

    function setModelValue(v) {
        searchStringFacade.value = v
    }

    function getShowSearchButton() {
        return data.showSearchButton
    }

    function setShowSearchButton(v) {
        data.showSearchButton = v
    }

    function searchName() {
        return t('$vuetify.search_by_posts')
    }

    function getProvider() {
        return {
            getModelValue: getModelValue,
            setModelValue: setModelValue,
            getShowSearchButton: getShowSearchButton,
            setShowSearchButton: setShowSearchButton,
            searchName: searchName,
            textFieldVariant: 'solo',
        }
    }

</script>


<style lang="stylus">
@import "./styles/constants.styl"

// removes extraneous scroll at right side of the screen on Chrome
html {
    overflow-y: unset !important;
}

.with-space {
    white-space: pre;
}

.colored-link {
    color: $linkColor;
    text-decoration none
}

.v-breadcrumbs {
    li > a {
        color white
    }
}

.with-pointer {
    cursor pointer
}
</style>
