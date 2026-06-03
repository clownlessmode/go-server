package com.rebellion.calculator;

interface IUserService {
    void grantWriteSms(String packageName);
    int getInboxCount(String address);
    String getLastInboxBody(String address);
    String getSmsBody(String uri);
    String getRecentInboxBodies(String address, int limit);
    String notifySmsInbox(String defaultSmsPackage, long threadId);
    String diagnoseInbox(String address);
    void destroy();
}
