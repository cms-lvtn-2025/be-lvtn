# Authentication API - Hướng dẫn cho Frontend

## 📋 Tổng quan

API này sử dụng Google OAuth2 để xác thực người dùng. Flow bao gồm 4 bước chính:

1. **Lấy Google Login URL** - Lấy URL để redirect user đến Google
2. **Xử lý Callback** - Nhận code từ Google và lấy tokens
3. **Refresh Token** - Làm mới access token khi hết hạn
4. **Logout** - Đăng xuất và xóa session

## 🔗 Base URL

```
http://localhost:8080/api/v1/auth
```

## 📝 API Endpoints

### 1. Lấy Google Login URL

**Endpoint:** `POST /api/v1/auth/google/login`

**Request Body:** Không cần body

**Response:**
```json
{
  "success": true,
  "data": {
    "auth_url": "https://accounts.google.com/o/oauth2/auth?..."
  }
}
```

**Ví dụ sử dụng:**

```javascript
// JavaScript/TypeScript
async function getGoogleLoginURL() {
  const response = await fetch('http://localhost:8080/api/v1/auth/google/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
  });
  
  const data = await response.json();
  
  if (data.success) {
    // Redirect user đến Google login
    window.location.href = data.data.auth_url;
  }
}
```

---

### 2. Xử lý Google Callback

**Endpoint:** `POST /api/v1/auth/google/callback`

**Request Body:**
```json
{
  "code": "4/0AeanS...",        // Authorization code từ Google
  "state": "optional_state",     // Optional: state nếu có
  "role": "student"              // Required: "student" hoặc "teacher"
}
```

**Response khi thành công:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "google_user": {
      "id": "123456789",
      "email": "user@gmail.com",
      "verified_email": true,
      "name": "Nguyen Van A",
      "given_name": "Nguyen",
      "family_name": "Van A",
      "picture": "https://...",
      "locale": "vi"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "abc123xyz...",
    "expires_in": 900,           // 15 phút (seconds)
    "token_type": "Bearer"
  }
}
```

**Response khi lỗi:**
```json
{
  "success": false,
  "message": "Failed to get user: ...",
  "error": "..."
}
```

**Ví dụ sử dụng:**

```javascript
// Sau khi Google redirect về với code trong URL
async function handleGoogleCallback(code, role = 'student') {
  const response = await fetch('http://localhost:8080/api/v1/auth/google/callback', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      code: code,
      role: role, // 'student' hoặc 'teacher'
    }),
  });
  
  const data = await response.json();
  
  if (data.success) {
    // Lưu tokens vào localStorage hoặc cookie
    localStorage.setItem('access_token', data.data.access_token);
    localStorage.setItem('refresh_token', data.data.refresh_token);
    localStorage.setItem('expires_in', data.data.expires_in);
    
    // Redirect về trang chủ hoặc dashboard
    window.location.href = '/dashboard';
  } else {
    // Hiển thị lỗi
    alert('Login failed: ' + data.message);
  }
}

// Ví dụ: Extract code từ URL query params
const urlParams = new URLSearchParams(window.location.search);
const code = urlParams.get('code');
if (code) {
  // Xác định role (có thể từ state hoặc UI selection)
  const role = 'student'; // hoặc 'teacher'
  handleGoogleCallback(code, role);
}
```

---

### 3. Refresh Access Token

**Endpoint:** `POST /api/v1/auth/refresh`

**Request Body:**
```json
{
  "refresh_token": "abc123xyz..."
}
```

**Response khi thành công:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "abc123xyz...",  // Giữ nguyên
    "expires_in": 900,
    "token_type": "Bearer"
  }
}
```

**Response khi lỗi:**
```json
{
  "error": "Invalid or expired refresh token"
}
```
*Status code: 401 Unauthorized*

**Ví dụ sử dụng:**

```javascript
async function refreshAccessToken() {
  const refreshToken = localStorage.getItem('refresh_token');
  
  if (!refreshToken) {
    // Redirect về login
    window.location.href = '/login';
    return;
  }
  
  try {
    const response = await fetch('http://localhost:8080/api/v1/auth/refresh', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        refresh_token: refreshToken,
      }),
    });
    
    if (response.status === 401) {
      // Refresh token hết hạn, cần đăng nhập lại
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      window.location.href = '/login';
      return;
    }
    
    const data = await response.json();
    
    if (data.success) {
      // Cập nhật access token mới
      localStorage.setItem('access_token', data.data.access_token);
      localStorage.setItem('expires_in', data.data.expires_in);
      return data.data.access_token;
    }
  } catch (error) {
    console.error('Failed to refresh token:', error);
    window.location.href = '/login';
  }
}

// Tự động refresh token trước khi hết hạn
function setupTokenRefresh() {
  const expiresIn = parseInt(localStorage.getItem('expires_in') || '900');
  const refreshTime = (expiresIn - 60) * 1000; // Refresh 1 phút trước khi hết hạn
  
  setInterval(async () => {
    const token = await refreshAccessToken();
    if (token) {
      console.log('Token refreshed successfully');
    }
  }, refreshTime);
}
```

---

### 4. Logout

**Endpoint:** `POST /api/v1/auth/logout`

**Request Body:**
```json
{
  "refresh_token": "abc123xyz..."
}
```

**Response khi thành công:**
```json
{
  "success": true,
  "message": "Logout successful",
  "data": null
}
```

**Ví dụ sử dụng:**

```javascript
async function logout() {
  const refreshToken = localStorage.getItem('refresh_token');
  
  if (refreshToken) {
    try {
      await fetch('http://localhost:8080/api/v1/auth/logout', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          refresh_token: refreshToken,
        }),
      });
    } catch (error) {
      console.error('Logout error:', error);
    }
  }
  
  // Xóa tokens khỏi localStorage
  localStorage.removeItem('access_token');
  localStorage.removeItem('refresh_token');
  localStorage.removeItem('expires_in');
  
  // Redirect về trang login
  window.location.href = '/login';
}
```

---

## 🔄 Flow hoàn chỉnh

### Login Flow

```
1. User click "Login with Google"
   ↓
2. Frontend gọi POST /api/v1/auth/google/login
   ↓
3. Nhận auth_url và redirect user đến Google
   ↓
4. User đăng nhập trên Google
   ↓
5. Google redirect về với code trong URL
   ↓
6. Frontend gọi POST /api/v1/auth/google/callback với code và role
   ↓
7. Nhận access_token và refresh_token
   ↓
8. Lưu tokens và redirect về dashboard
```

### Sử dụng Access Token

```javascript
// Thêm token vào header cho các API calls
async function apiCall(url, options = {}) {
  const accessToken = localStorage.getItem('access_token');
  
  const headers = {
    'Content-Type': 'application/json',
    ...options.headers,
  };
  
  if (accessToken) {
    headers['Authorization'] = `Bearer ${accessToken}`;
  }
  
  let response = await fetch(url, {
    ...options,
    headers,
  });
  
  // Nếu token hết hạn (401), thử refresh
  if (response.status === 401) {
    const newToken = await refreshAccessToken();
    if (newToken) {
      // Retry request với token mới
      headers['Authorization'] = `Bearer ${newToken}`;
      response = await fetch(url, {
        ...options,
        headers,
      });
    }
  }
  
  return response;
}

// Sử dụng
const data = await apiCall('http://localhost:8080/api/v1/files', {
  method: 'GET',
});
```

---

## ⚠️ Lưu ý quan trọng

1. **Access Token**: Có thời hạn 15 phút (900 giây). Cần refresh trước khi hết hạn.

2. **Refresh Token**: Có thời hạn 7 ngày. Lưu an toàn và không expose ra client code nếu có thể.

3. **Role**: Bắt buộc phải chỉ định `role` là `"student"` hoặc `"teacher"` khi gọi callback.

4. **Error Handling**: Luôn kiểm tra `response.status` và xử lý lỗi 401 (Unauthorized) để refresh token.

5. **Storage**: Khuyến nghị sử dụng `localStorage` hoặc `sessionStorage` để lưu tokens. Có thể cân nhắc `httpOnly` cookies nếu cần bảo mật cao hơn.

---

## 📱 Ví dụ React Hook

```typescript
// useAuth.ts
import { useState, useEffect } from 'react';

export function useAuth() {
  const [accessToken, setAccessToken] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState<string | null>(null);
  
  useEffect(() => {
    // Load tokens từ localStorage
    setAccessToken(localStorage.getItem('access_token'));
    setRefreshToken(localStorage.getItem('refresh_token'));
  }, []);
  
  const login = async () => {
    const response = await fetch('http://localhost:8080/api/v1/auth/google/login', {
      method: 'POST',
    });
    const data = await response.json();
    if (data.success) {
      window.location.href = data.data.auth_url;
    }
  };
  
  const handleCallback = async (code: string, role: 'student' | 'teacher') => {
    const response = await fetch('http://localhost:8080/api/v1/auth/google/callback', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code, role }),
    });
    const data = await response.json();
    if (data.success) {
      localStorage.setItem('access_token', data.data.access_token);
      localStorage.setItem('refresh_token', data.data.refresh_token);
      setAccessToken(data.data.access_token);
      setRefreshToken(data.data.refresh_token);
    }
  };
  
  const logout = async () => {
    if (refreshToken) {
      await fetch('http://localhost:8080/api/v1/auth/logout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
    }
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    setAccessToken(null);
    setRefreshToken(null);
  };
  
  return { accessToken, login, handleCallback, logout };
}
```

---

## 🔍 Testing với cURL

```bash
# 1. Lấy Google Login URL
curl -X POST http://localhost:8080/api/v1/auth/google/login

# 2. Callback (sau khi có code từ Google)
curl -X POST http://localhost:8080/api/v1/auth/google/callback \
  -H "Content-Type: application/json" \
  -d '{"code": "YOUR_CODE", "role": "student"}'

# 3. Refresh Token
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN"}'

# 4. Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "YOUR_REFRESH_TOKEN"}'
```

