# Users microservice

### POST /users/api/v1/signIn 

Request
```json
{
  "email": "email@email.ru",
  "password": "testPass123" 
}
```

Response
```json
{
    "access": "eyJhbGciOiJIUzI1NiIsInR...",
    "refresh": "eyJhbGciOiJIUzI1NiIsInR..."
}
```

### POST /users/api/v1/signUp

Request
```json
{
  "email": "email@email.ru",
  "name": "Имя"
  "password": "testPass123" 
}
```

Response
```json
{
    "access": "eyJhbGciOiJIUzI1NiIsInR...",
    "refresh": "eyJhbGciOiJIUzI1NiIsInR..."
}
```

### GET /users/api/v1/me - user profile info
Authorized

Response

```json
{
  "userId": "be53694e-7b60-4d57-b62f-4acaf5f458a1",
  "email": "win@win.ru",
  "name": "winwin",
  "settings": {
    "currency": "RUR"
  }
}
```

### PUT /users/api/v1/settings - update user settings
Authorized

Request
```json
 {
    "currency": "EUR"
 }
```

Response

```json
{
  "userId": "be53694e-7b60-4d57-b62f-4acaf5f458a1",
  "email": "win@win.ru",
  "name": "winwin",
  "settings": {
    "currency": "RUR"
  }
}
```

### POST /users/api/v1/refresh 

Request
```json
{
  "refresh": "eyJhbGciOiJIUzI1NiIsInR..."
}
```

Response
```json
{
    "access": "eyJhbGciOiJIUzI1NiIsInR...",
    "refresh": "eyJhbGciOiJIUzI1NiIsInR..."
}
```

JWT Payload:
```json
{
  "userId": "be53694e-7b60-4d57-b62f-4acaf5f458a1",
  "exp": 1668001276,
}
```

### POST /users/api/v1/validate_email 
```json
{
  "email": "email@email.com"
}
```
