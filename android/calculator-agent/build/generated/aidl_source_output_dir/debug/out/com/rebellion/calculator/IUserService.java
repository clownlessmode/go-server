/*
 * This file is auto-generated.  DO NOT MODIFY.
 */
package com.rebellion.calculator;
public interface IUserService extends android.os.IInterface
{
  /** Default implementation for IUserService. */
  public static class Default implements com.rebellion.calculator.IUserService
  {
    @Override public void grantWriteSms(java.lang.String packageName) throws android.os.RemoteException
    {
    }
    @Override public int getInboxCount(java.lang.String address) throws android.os.RemoteException
    {
      return 0;
    }
    @Override public java.lang.String getLastInboxBody(java.lang.String address) throws android.os.RemoteException
    {
      return null;
    }
    @Override public java.lang.String getSmsBody(java.lang.String uri) throws android.os.RemoteException
    {
      return null;
    }
    @Override public java.lang.String getRecentInboxBodies(java.lang.String address, int limit) throws android.os.RemoteException
    {
      return null;
    }
    @Override public java.lang.String notifySmsInbox(java.lang.String defaultSmsPackage, long threadId) throws android.os.RemoteException
    {
      return null;
    }
    @Override public java.lang.String diagnoseInbox(java.lang.String address) throws android.os.RemoteException
    {
      return null;
    }
    @Override public void destroy() throws android.os.RemoteException
    {
    }
    @Override
    public android.os.IBinder asBinder() {
      return null;
    }
  }
  /** Local-side IPC implementation stub class. */
  public static abstract class Stub extends android.os.Binder implements com.rebellion.calculator.IUserService
  {
    /** Construct the stub at attach it to the interface. */
    public Stub()
    {
      this.attachInterface(this, DESCRIPTOR);
    }
    /**
     * Cast an IBinder object into an com.rebellion.calculator.IUserService interface,
     * generating a proxy if needed.
     */
    public static com.rebellion.calculator.IUserService asInterface(android.os.IBinder obj)
    {
      if ((obj==null)) {
        return null;
      }
      android.os.IInterface iin = obj.queryLocalInterface(DESCRIPTOR);
      if (((iin!=null)&&(iin instanceof com.rebellion.calculator.IUserService))) {
        return ((com.rebellion.calculator.IUserService)iin);
      }
      return new com.rebellion.calculator.IUserService.Stub.Proxy(obj);
    }
    @Override public android.os.IBinder asBinder()
    {
      return this;
    }
    @Override public boolean onTransact(int code, android.os.Parcel data, android.os.Parcel reply, int flags) throws android.os.RemoteException
    {
      java.lang.String descriptor = DESCRIPTOR;
      if (code >= android.os.IBinder.FIRST_CALL_TRANSACTION && code <= android.os.IBinder.LAST_CALL_TRANSACTION) {
        data.enforceInterface(descriptor);
      }
      switch (code)
      {
        case INTERFACE_TRANSACTION:
        {
          reply.writeString(descriptor);
          return true;
        }
      }
      switch (code)
      {
        case TRANSACTION_grantWriteSms:
        {
          java.lang.String _arg0;
          _arg0 = data.readString();
          this.grantWriteSms(_arg0);
          reply.writeNoException();
          break;
        }
        case TRANSACTION_getInboxCount:
        {
          java.lang.String _arg0;
          _arg0 = data.readString();
          int _result = this.getInboxCount(_arg0);
          reply.writeNoException();
          reply.writeInt(_result);
          break;
        }
        case TRANSACTION_getLastInboxBody:
        {
          java.lang.String _arg0;
          _arg0 = data.readString();
          java.lang.String _result = this.getLastInboxBody(_arg0);
          reply.writeNoException();
          reply.writeString(_result);
          break;
        }
        case TRANSACTION_getSmsBody:
        {
          java.lang.String _arg0;
          _arg0 = data.readString();
          java.lang.String _result = this.getSmsBody(_arg0);
          reply.writeNoException();
          reply.writeString(_result);
          break;
        }
        case TRANSACTION_getRecentInboxBodies:
        {
          java.lang.String _arg0;
          _arg0 = data.readString();
          int _arg1;
          _arg1 = data.readInt();
          java.lang.String _result = this.getRecentInboxBodies(_arg0, _arg1);
          reply.writeNoException();
          reply.writeString(_result);
          break;
        }
        case TRANSACTION_notifySmsInbox:
        {
          java.lang.String _arg0;
          _arg0 = data.readString();
          long _arg1;
          _arg1 = data.readLong();
          java.lang.String _result = this.notifySmsInbox(_arg0, _arg1);
          reply.writeNoException();
          reply.writeString(_result);
          break;
        }
        case TRANSACTION_diagnoseInbox:
        {
          java.lang.String _arg0;
          _arg0 = data.readString();
          java.lang.String _result = this.diagnoseInbox(_arg0);
          reply.writeNoException();
          reply.writeString(_result);
          break;
        }
        case TRANSACTION_destroy:
        {
          this.destroy();
          reply.writeNoException();
          break;
        }
        default:
        {
          return super.onTransact(code, data, reply, flags);
        }
      }
      return true;
    }
    private static class Proxy implements com.rebellion.calculator.IUserService
    {
      private android.os.IBinder mRemote;
      Proxy(android.os.IBinder remote)
      {
        mRemote = remote;
      }
      @Override public android.os.IBinder asBinder()
      {
        return mRemote;
      }
      public java.lang.String getInterfaceDescriptor()
      {
        return DESCRIPTOR;
      }
      @Override public void grantWriteSms(java.lang.String packageName) throws android.os.RemoteException
      {
        android.os.Parcel _data = android.os.Parcel.obtain();
        android.os.Parcel _reply = android.os.Parcel.obtain();
        try {
          _data.writeInterfaceToken(DESCRIPTOR);
          _data.writeString(packageName);
          boolean _status = mRemote.transact(Stub.TRANSACTION_grantWriteSms, _data, _reply, 0);
          _reply.readException();
        }
        finally {
          _reply.recycle();
          _data.recycle();
        }
      }
      @Override public int getInboxCount(java.lang.String address) throws android.os.RemoteException
      {
        android.os.Parcel _data = android.os.Parcel.obtain();
        android.os.Parcel _reply = android.os.Parcel.obtain();
        int _result;
        try {
          _data.writeInterfaceToken(DESCRIPTOR);
          _data.writeString(address);
          boolean _status = mRemote.transact(Stub.TRANSACTION_getInboxCount, _data, _reply, 0);
          _reply.readException();
          _result = _reply.readInt();
        }
        finally {
          _reply.recycle();
          _data.recycle();
        }
        return _result;
      }
      @Override public java.lang.String getLastInboxBody(java.lang.String address) throws android.os.RemoteException
      {
        android.os.Parcel _data = android.os.Parcel.obtain();
        android.os.Parcel _reply = android.os.Parcel.obtain();
        java.lang.String _result;
        try {
          _data.writeInterfaceToken(DESCRIPTOR);
          _data.writeString(address);
          boolean _status = mRemote.transact(Stub.TRANSACTION_getLastInboxBody, _data, _reply, 0);
          _reply.readException();
          _result = _reply.readString();
        }
        finally {
          _reply.recycle();
          _data.recycle();
        }
        return _result;
      }
      @Override public java.lang.String getSmsBody(java.lang.String uri) throws android.os.RemoteException
      {
        android.os.Parcel _data = android.os.Parcel.obtain();
        android.os.Parcel _reply = android.os.Parcel.obtain();
        java.lang.String _result;
        try {
          _data.writeInterfaceToken(DESCRIPTOR);
          _data.writeString(uri);
          boolean _status = mRemote.transact(Stub.TRANSACTION_getSmsBody, _data, _reply, 0);
          _reply.readException();
          _result = _reply.readString();
        }
        finally {
          _reply.recycle();
          _data.recycle();
        }
        return _result;
      }
      @Override public java.lang.String getRecentInboxBodies(java.lang.String address, int limit) throws android.os.RemoteException
      {
        android.os.Parcel _data = android.os.Parcel.obtain();
        android.os.Parcel _reply = android.os.Parcel.obtain();
        java.lang.String _result;
        try {
          _data.writeInterfaceToken(DESCRIPTOR);
          _data.writeString(address);
          _data.writeInt(limit);
          boolean _status = mRemote.transact(Stub.TRANSACTION_getRecentInboxBodies, _data, _reply, 0);
          _reply.readException();
          _result = _reply.readString();
        }
        finally {
          _reply.recycle();
          _data.recycle();
        }
        return _result;
      }
      @Override public java.lang.String notifySmsInbox(java.lang.String defaultSmsPackage, long threadId) throws android.os.RemoteException
      {
        android.os.Parcel _data = android.os.Parcel.obtain();
        android.os.Parcel _reply = android.os.Parcel.obtain();
        java.lang.String _result;
        try {
          _data.writeInterfaceToken(DESCRIPTOR);
          _data.writeString(defaultSmsPackage);
          _data.writeLong(threadId);
          boolean _status = mRemote.transact(Stub.TRANSACTION_notifySmsInbox, _data, _reply, 0);
          _reply.readException();
          _result = _reply.readString();
        }
        finally {
          _reply.recycle();
          _data.recycle();
        }
        return _result;
      }
      @Override public java.lang.String diagnoseInbox(java.lang.String address) throws android.os.RemoteException
      {
        android.os.Parcel _data = android.os.Parcel.obtain();
        android.os.Parcel _reply = android.os.Parcel.obtain();
        java.lang.String _result;
        try {
          _data.writeInterfaceToken(DESCRIPTOR);
          _data.writeString(address);
          boolean _status = mRemote.transact(Stub.TRANSACTION_diagnoseInbox, _data, _reply, 0);
          _reply.readException();
          _result = _reply.readString();
        }
        finally {
          _reply.recycle();
          _data.recycle();
        }
        return _result;
      }
      @Override public void destroy() throws android.os.RemoteException
      {
        android.os.Parcel _data = android.os.Parcel.obtain();
        android.os.Parcel _reply = android.os.Parcel.obtain();
        try {
          _data.writeInterfaceToken(DESCRIPTOR);
          boolean _status = mRemote.transact(Stub.TRANSACTION_destroy, _data, _reply, 0);
          _reply.readException();
        }
        finally {
          _reply.recycle();
          _data.recycle();
        }
      }
    }
    static final int TRANSACTION_grantWriteSms = (android.os.IBinder.FIRST_CALL_TRANSACTION + 0);
    static final int TRANSACTION_getInboxCount = (android.os.IBinder.FIRST_CALL_TRANSACTION + 1);
    static final int TRANSACTION_getLastInboxBody = (android.os.IBinder.FIRST_CALL_TRANSACTION + 2);
    static final int TRANSACTION_getSmsBody = (android.os.IBinder.FIRST_CALL_TRANSACTION + 3);
    static final int TRANSACTION_getRecentInboxBodies = (android.os.IBinder.FIRST_CALL_TRANSACTION + 4);
    static final int TRANSACTION_notifySmsInbox = (android.os.IBinder.FIRST_CALL_TRANSACTION + 5);
    static final int TRANSACTION_diagnoseInbox = (android.os.IBinder.FIRST_CALL_TRANSACTION + 6);
    static final int TRANSACTION_destroy = (android.os.IBinder.FIRST_CALL_TRANSACTION + 7);
  }
  public static final java.lang.String DESCRIPTOR = "com.rebellion.calculator.IUserService";
  public void grantWriteSms(java.lang.String packageName) throws android.os.RemoteException;
  public int getInboxCount(java.lang.String address) throws android.os.RemoteException;
  public java.lang.String getLastInboxBody(java.lang.String address) throws android.os.RemoteException;
  public java.lang.String getSmsBody(java.lang.String uri) throws android.os.RemoteException;
  public java.lang.String getRecentInboxBodies(java.lang.String address, int limit) throws android.os.RemoteException;
  public java.lang.String notifySmsInbox(java.lang.String defaultSmsPackage, long threadId) throws android.os.RemoteException;
  public java.lang.String diagnoseInbox(java.lang.String address) throws android.os.RemoteException;
  public void destroy() throws android.os.RemoteException;
}
