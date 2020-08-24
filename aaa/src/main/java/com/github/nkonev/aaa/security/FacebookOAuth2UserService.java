package com.github.nkonev.aaa.security;

import com.github.nkonev.aaa.converter.UserAccountConverter;
import com.github.nkonev.aaa.repository.jdbc.UserAccountRepository;
import com.github.nkonev.aaa.dto.UserAccountDetailsDTO;
import com.github.nkonev.aaa.entity.jdbc.UserAccount;
import com.github.nkonev.aaa.security.checks.AaaPostAuthenticationChecks;
import com.github.nkonev.aaa.security.checks.AaaPreAuthenticationChecks;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Scope;
import org.springframework.context.annotation.ScopedProxyMode;
import org.springframework.security.oauth2.client.userinfo.OAuth2UserRequest;
import org.springframework.security.oauth2.client.userinfo.OAuth2UserService;
import org.springframework.security.oauth2.core.OAuth2AuthenticationException;
import org.springframework.security.oauth2.core.user.OAuth2User;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.util.Assert;
import java.util.Map;
import java.util.Optional;


@Transactional
@Scope(proxyMode = ScopedProxyMode.TARGET_CLASS)
@Component
public class FacebookOAuth2UserService extends AbstractOAuth2UserService implements OAuth2UserService<OAuth2UserRequest, OAuth2User> {

    private static final Logger LOGGER = LoggerFactory.getLogger(FacebookOAuth2UserService.class);

    @Autowired
    private UserAccountRepository userAccountRepository;

    public static final String LOGIN_PREFIX = "facebook_";

    @Autowired
    private AaaPreAuthenticationChecks aaaPreAuthenticationChecks;

    @Autowired
    private AaaPostAuthenticationChecks aaaPostAuthenticationChecks;


    @Override
    public OAuth2User loadUser(OAuth2UserRequest userRequest) throws OAuth2AuthenticationException {
        OAuth2User oAuth2User = delegate.loadUser(userRequest);

        var map = oAuth2User.getAttributes();
        String facebookId = getId(map);
        Assert.notNull(facebookId, "facebookId cannot be null");


        UserAccountDetailsDTO resultPrincipal = mergeOauthIdToExistsUser(facebookId);
        if (resultPrincipal != null) {
            // ok
        } else {
            String login = getLogin(map);
            resultPrincipal = createOrGetExistsUser(facebookId, login, map);
        }

        aaaPreAuthenticationChecks.check(resultPrincipal);
        aaaPostAuthenticationChecks.check(resultPrincipal);
        return resultPrincipal;
    }


    private String getAvatarUrl(Map<String, Object> map){
        return null;
    }

    private String getLogin(Map<String, Object> map) {
        String login = (String) map.get("name");
        Assert.hasLength(login, "facebook name cannot be null");
        login = login.trim();
        login = login.replaceAll(" +", " ");
        return login;
    }

    private String getId(Map<String, Object> map) {
        return (String) map.get("id");
    }

    @Override
    protected Logger logger() {
        return LOGGER;
    }

    @Override
    protected String getOauthName() {
        return "facebook";
    }

    @Override
    protected Optional<UserAccount> findByOauthId(String oauthId) {
        return userAccountRepository.findByOauthIdentifiersFacebookId(oauthId);
    }

    @Override
    protected void setOauthIdToPrincipal(UserAccountDetailsDTO principal, String oauthId) {
        principal.getOauthIdentifiers().setFacebookId(oauthId);
    }

    @Override
    protected void setOauthIdToEntity(Long id, String oauthId) {
        UserAccount userAccount = userAccountRepository.findById(id).orElseThrow();
        userAccount.getOauthIdentifiers().setFacebookId(oauthId);
        userAccount = userAccountRepository.save(userAccount);
    }

    @Override
    protected UserAccount insertEntity(String oauthId, String login, Map<String, Object> map) {
        String maybeImageUrl = getAvatarUrl(map);
        UserAccount userAccount = UserAccountConverter.buildUserAccountEntityForFacebookInsert(oauthId, login, maybeImageUrl);
        userAccount = userAccountRepository.save(userAccount);
        LOGGER.info("Created facebook user id={} login='{}'", oauthId, login);

        return userAccount;
    }

    @Override
    protected String getLoginPrefix() {
        return LOGIN_PREFIX;
    }

    @Override
    protected Optional<UserAccount> findByUsername(String login) {
        return userAccountRepository.findByUsername(login);
    }

}
