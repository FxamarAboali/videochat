package name.nkonev.video.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import name.nkonev.video.dto.out.Typed;
import name.nkonev.video.dto.out.WithUserSession;
import okhttp3.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.net.URI;

@Component
public class ChatRequestService {

    @Autowired
    private OkHttpClient okHttpClient;

    @Value("${chat.url}")
    private String chatUrl;

    @Autowired
    private ObjectMapper objectMapper;

    private static final Logger LOGGER = LoggerFactory.getLogger(ChatRequestService.class);

    public void sendToWebsocketForSession(String sessionId, Typed data) {
        try {
            final RequestBody requestBody = RequestBody.create(
                    MediaType.get("application/json;charset=UTF-8"),
                    objectMapper.writeValueAsBytes(data)
            );
            final Request request = new Request.Builder()
                    .url(chatUrl + "?toUser=" + sessionId)
                    .post(requestBody)
                    .build();
            final Response response = okHttpClient.newCall(request).execute();
            final String json = response.body().string();
            LOGGER.debug("chat response: {}", json);
        } catch (Exception e) {
            LOGGER.error("failed chat invocation", e);
        }
    }

}
