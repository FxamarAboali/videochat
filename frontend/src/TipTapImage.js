import axios from "axios";

export const embedUploadFunction = (chatId, fileObj) => {
    const formData = new FormData();
    formData.append('embed_file_header', fileObj);
    return axios.post('/api/storage/'+chatId+'/embed', formData)
        .then((result) => {
            let url = result.data.relativeUrl; // Get url from response
            console.debug("got embed url", url);
            return url;
        })
}

export const buildImageHandler = (chatId) => {
    const MyImage = require('@tiptap/extension-image').Image;
    const prosemirrorState = require('prosemirror-state');

    MyImage.config.addProseMirrorPlugins = () => {
        return [
            new prosemirrorState.Plugin({
                key: new prosemirrorState.PluginKey('imageHandler'),
                props: {
                    handlePaste: (view, event) => {
                        const items = (event.clipboardData || event.originalEvent.clipboardData).items;
                        for (const item of items) {
                            if (item.type.indexOf("image") === 0) {
                                event.preventDefault();
                                const {schema} = view.state;

                                const image = item.getAsFile();

                                embedUploadFunction(chatId, image).then(src => {
                                    const node = schema.nodes.image.create({
                                        src: src,
                                    });
                                    const transaction = view.state.tr.replaceSelectionWith(node);
                                    view.dispatch(transaction)
                                });

                            }
                        }
                    },

                    handleDOMEvents: {
                        drop: (view, event) => {
                            const hasFiles =
                                event.dataTransfer &&
                                event.dataTransfer.files &&
                                event.dataTransfer.files.length;

                            if (!hasFiles) {
                                return false;
                            }

                            const images = Array.from(
                                event.dataTransfer?.files ?? []
                            ).filter((file) => /image/i.test(file.type));

                            if (images.length === 0) {
                                return false;
                            }

                            event.preventDefault();

                            const { schema } = view.state;
                            const coordinates = view.posAtCoords({
                                left: event.clientX,
                                top: event.clientY,
                            });
                            if (!coordinates) return false;

                            images.forEach(async (image) => {

                                const node = schema.nodes.image.create({
                                    src: await embedUploadFunction(chatId, image),
                                });
                                const transaction = view.state.tr.insert(coordinates.pos, node);
                                view.dispatch(transaction);

                            });

                            return true;
                        },
                    },
                }
            }),
        ];
    }
    return MyImage
};