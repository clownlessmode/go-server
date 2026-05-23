package com.rebellion.calculator;

interface IUserService {
    int insertSms(String address, String body);
    void destroy();
}
