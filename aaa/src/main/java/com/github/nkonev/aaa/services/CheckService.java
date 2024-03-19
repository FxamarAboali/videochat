package com.github.nkonev.aaa.services;

import com.github.nkonev.aaa.dto.EditUserDTO;
import com.github.nkonev.aaa.entity.jdbc.UserAccount;
import com.github.nkonev.aaa.exception.UserAlreadyPresentException;
import com.github.nkonev.aaa.repository.jdbc.UserAccountRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

@Service
public class CheckService {
    private static final Logger LOGGER = LoggerFactory.getLogger(CheckService.class);

    @Autowired
    private UserAccountRepository userAccountRepository;

    public void checkLoginIsFree(EditUserDTO userAccountDTO, UserAccount exists) {
        if (!exists.username().equals(userAccountDTO.login()) && userAccountRepository.findByUsername(userAccountDTO.login()).isPresent()) {
            throw new UserAlreadyPresentException("User with login '" + userAccountDTO.login() + "' is already present");
        }
    }

    public void checkLoginIsFree(EditUserDTO userAccountDTO) {
        if(userAccountRepository.findByUsername(userAccountDTO.login()).isPresent()){
            throw new UserAlreadyPresentException("User with login '" + userAccountDTO.login() + "' is already present");
        }
    }

    public boolean checkEmailIsFree(EditUserDTO userAccountDTO, UserAccount exists) {
        if (exists.email() != null && !exists.email().equals(userAccountDTO.email()) && userAccountDTO.email() != null && userAccountRepository.findByEmail(userAccountDTO.email()).isPresent()) {
            LOGGER.warn("user with email '{}' already present. exiting...", exists.email());
            return false;
        } else {
            return true;
        }
    }

    public boolean checkEmailIsFree(String email) {
        if (userAccountRepository.findByEmail(email).isPresent()) {
            LOGGER.warn("user with email '{}' already present. exiting...", email);
            return false;
        } else {
            return true;
        }
    }

    public boolean checkEmailIsFree(EditUserDTO userAccountDTO) {
        if(userAccountRepository.findByEmail(userAccountDTO.email()).isPresent()){
            LOGGER.warn("Skipping sent registration email '{}' because this user already present", userAccountDTO.email());
            return false; // we care for user email leak
        } else {
            return true;
        }
    }

}