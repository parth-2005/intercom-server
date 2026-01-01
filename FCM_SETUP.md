# FCM Wake-on-LAN Implementation

## 🔥 What Changed

### New Files
- **[fcm.go](fcm.go)** - Firebase Cloud Messaging integration
- **[serviceAccountKey.json.example](serviceAccountKey.json.example)** - Template for Firebase credentials

### Updated Files
- **[hub.go](hub.go)** - Added persistent FCM token storage and offline call handling
- **[main.go](main.go)** - Firebase initialization and FCM token capture
- **[go.mod](go.mod)** - Firebase dependencies added

---

## 🧠 How It Works

### 1. Token Registration
When a Flutter client connects:
```
ws://<server>:8081/ws?user=dad&fcm_token=xyz123...
```
The server stores the FCM token in `Hub.fcmTokens` (persistent across disconnects).

### 2. Call Flow

**Online User (Connected via WebSocket):**
```
Caller → Server → [WebSocket] → Target
```

**Offline User (App backgrounded/killed):**
```
Caller → Server → [FCM Push] → Target Device → CallKit UI
```

### 3. FCM Payload Structure
```json
{
  "data": {
    "type": "call_initiate",
    "uuid": "call-uuid-123",
    "callerName": "dad",
    "callerHandle": "dad"
  }
}
```

This is a **data-only message** so `flutter_callkit_incoming` can intercept it in the background.

---

## 🚀 Setup Instructions

### 1. Get Firebase Service Account Key
1. Go to [Firebase Console](https://console.firebase.google.com)
2. Project Settings → Service Accounts
3. Generate new private key
4. Save as `serviceAccountKey.json` in this directory

### 2. Update `.gitignore`
Already done! `serviceAccountKey.json` is ignored.

### 3. Run Server
```bash
go run .
# or
./signaling.exe
```

If the service account file is missing, server will log a warning but continue running (without FCM support).

---

## 📱 Flutter Integration

### Connect with FCM Token
```dart
final fcmToken = await FirebaseMessaging.instance.getToken();
final uri = 'ws://192.168.1.100:8081/ws?user=dad&fcm_token=$fcmToken';
```

### Handle Background Push
```dart
FirebaseMessaging.onBackgroundMessage(_firebaseMessagingBackgroundHandler);

Future<void> _firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  final data = message.data;
  if (data['type'] == 'call_initiate') {
    await FlutterCallkitIncoming.showCallkitIncoming(
      CallKitParams(
        id: data['uuid'],
        nameCaller: data['callerName'],
        handle: data['callerHandle'],
        type: 1, // audio call
      ),
    );
  }
}
```

---

## 🔒 Thread Safety

The `fcmTokens` map is protected by `sync.RWMutex`:
- `SetFCMToken()` - Write lock
- `GetFCMToken()` - Read lock

Safe for concurrent access from multiple goroutines.

---

## 🛠️ Testing

### 1. Simulate Offline User
1. Connect user A
2. Disconnect user B (but keep their FCM token registered)
3. User A calls user B
4. Check server logs for FCM send confirmation

### 2. Verify FCM Message
Use Firebase Console → Cloud Messaging → Send test message with the stored token.

---

## 🎯 What This Solves

✅ Calls work even when app is killed/backgrounded  
✅ Native CallKit UI on iOS  
✅ Native incoming call UI on Android  
✅ Zero polling/battery drain  
✅ Instant wake-up  

This is **production-grade** for family-scale (4-6 users).
