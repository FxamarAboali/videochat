// This is your plugin object. It can be exported to be used anywhere.
import Centrifuge from "centrifuge";

export const setupCentrifuge = (centrifugeSessionFunction) => {
    // Create Centrifuge object with Websocket endpoint address set in main.go
    var url = ((window.location.protocol === "https:") ? "wss://" : "ws://") + window.location.host + "/api/chat/websocket";
    var centrifuge = new Centrifuge(url, {
        onRefresh: (ctx)=>{
            console.debug("Dummy refresh");
        }
    });
    centrifuge.on('connect', (ctx)=>{
        console.log("Connected response", ctx);
        centrifugeSessionFunction(ctx.client);
        console.log('My centrifuge session clientId :', ctx.client);
    });
    centrifuge.on('disconnect', (ctx)=>{
        console.log("Disconnected response", ctx);
    });
    centrifuge.connect();
    return centrifuge;
};
