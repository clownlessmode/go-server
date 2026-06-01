package com.rebellion.calculator;

interface IUserService {
    String insertSms(String address, String body);
    String diagnoseInbox(String address);
    void destroy();
}
