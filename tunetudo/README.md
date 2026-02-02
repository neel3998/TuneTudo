# TuneTudo - Music Streaming Application

A web-based music streaming application built with Go Fiber v2 following MVC architecture pattern.

## Features

- 🎵 Search for songs, artists, and albums
- 📝 Create and manage playlists
- 🎧 Stream music in real-time
- 📤 Upload personal tracks
- 🗂️ Browse music by categories
- 👤 User profile management
- 🔐 JWT-based authentication
- 👨‍💼 Admin portal for catalog management

## Technology Stack

- **Backend Framework**: Go Fiber v2
- **Database**: SQLite3
- **Authentication**: JWT (JSON Web Tokens)
- **Password Hashing**: sha1
- **Architecture**: MVC (Model-View-Controller)

## Project Structure

```
tunetudo/
├── main.go                 # Application entry point
├── main_test.go
├── Dockerfile
├── Makefile
├──tunetudo.db
├── config/                 # Configuration management
│   └── config.go
├── database/               # Database setup and migrations
│   └── database.go
├── models/                 # Data models
│   └── models.go
├── services/               # Business logic layer
│   ├── auth_service.go
│   ├── search_service.go
│   ├── playlist_service.go
│   ├── playback_service.go
│   ├── user_service.go
│   ├── admin_service.go
│   ├── auth_service_test.go
│   ├── playback_service_test.go
│   ├──  playlist_service_test.go
│   ├──  search_service_test.go
│   └── test_helper.go
├── controllers/            # HTTP request handlers
│   └── controllers.go
├── middleware/             # Custom middleware
│   └── middleware.go
├── routes/                 # Route definitions
│   └── routes.go
├── storage/                # File storage
│   ├── media/
│   │   ├── songs/
│   │   └── uploads/
│   └── images/
│       ├── profiles/
│       └── covers/
├── go.mod
├── go.sum
└── .env
```

## Installation

### Prerequisites

- Go 1.21 or higher
- SQLite3

### Setup Steps

1. **Clone the repository**
```bash
git clone <repository-url>
cd tunetudo
```

2. **Install dependencies**
```bash
go mod download
```

3. **Create environment file**
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. **Create storage directories**
```bash
mkdir -p storage/media/songs storage/media/uploads storage/images/profiles storage/images/covers
```

5. **Run the application**
```bash
go run main.go
```

The server will start on `http://localhost:2701` by default.

## API Endpoints

### Authentication

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/auth/register` | Register new user | No |
| POST | `/api/auth/login` | Login user | No |
| POST | `/api/auth/logout` | Logout user | Yes |
| GET | `/api/profile` | Get user profile | Yes |

### Search & Browse

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/search?q={query}` | Search songs, artists, albums | No |
| GET | `/api/categories` | Get all categories | No |
| GET | `/api/categories/:id/songs` | Get songs by category | No |
| GET | `/api/songs/recent` | Get recently added songs | No |
| GET | `/api/songs/:id` | Get song details | No |
| GET | `/api/songs/:id/stream` | Stream song audio | No |

### Playlists

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/playlists` | Get user playlists | Yes |
| POST | `/api/playlists` | Create new playlist | Yes |
| GET | `/api/playlists/:id` | Get playlist details | Yes |
| POST | `/api/playlists/:id/songs` | Add song to playlist | Yes |
| DELETE | `/api/playlists/:id/songs/:songId` | Remove song from playlist | Yes |
| DELETE | `/api/playlists/:id` | Delete playlist | Yes |

### User Operations

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| PUT | `/api/profile/picture` | Upload profile picture | Yes |
| POST | `/api/upload` | Upload personal track | Yes |
| GET | `/api/uploads` | Get user uploads | Yes |

### Admin Operations

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/admin/songs` | Upload new song to catalog | Admin |
| DELETE | `/api/admin/songs/:id` | Delete song from catalog | Admin |
| GET | `/api/admin/songs` | Get all songs (paginated) | Admin |
| GET | `/api/admin/users` | Get all users | Admin |

## API Usage Examples

### Register a User

```bash
curl -X POST http://localhost:2701/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "SecurePass123"
  }'
```

### Login

```bash
curl -X POST http://localhost:2701/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "SecurePass123"
  }'
```

Response:
```json
{
  "error": false,
  "message": "login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "johndoe",
      "email": "john@example.com",
      "is_admin": false
    }
  }
}
```

### Search for Songs

```bash
curl http://localhost:2701/api/search?q=yellow
```

### Create a Playlist

```bash
curl -X POST http://localhost:2701/api/playlists \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "My Favorites",
    "description": "Collection of my favorite songs"
  }'
```

### Add Song to Playlist

```bash
curl -X POST http://localhost:2701/api/playlists/1/songs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "song_id": 5
  }'
```

### Upload Personal Track

```bash
curl -X POST http://localhost:2701/api/upload \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@/path/to/song.mp4"
```

### Admin: Upload Song to Catalog

```bash
curl -X POST http://localhost:2701/api/admin/songs \
  -H "Authorization: Bearer ADMIN_JWT_TOKEN" \
  -F "file=@/path/to/song.mp4" \
  -F "title=Song Title" \
  -F "artist=Artist Name" \
  -F "album=Album Title" \
  -F "category_id=1" \
  -F "duration=240"
```

## Database Schema

### Tables

- **users** - User accounts
- **artists** - Music artists
- **albums** - Music albums
- **categories** - Genre/category classifications
- **songs** - Song catalog
- **playlists** - User playlists
- **playlist_songs** - Songs in playlists (junction table)
- **uploads** - User file upload records
- **songs_fts** - Full-text search virtual table

## Security Features

- JWT token-based authentication
- Password hashing with bcrypt
- Role-based access control (User/Admin)
- File type validation for uploads
- File size limits
- SQL injection protection via parameterized queries

## File Upload Limits

- **Audio Files**: 50MB maximum (.mp4, .wav, .mp3)
- **Images**: 5MB maximum (.jpg, .jpeg, .png)

## Testing the Application

### Manual Testing

1. **Register a user**
2. **Login to get JWT token**
3. **Use the token in Authorization header** for protected endpoints
4. **Test all functional requirements** as per the test cases in the design specification

### Creating an Admin User

Run this SQL directly on the database:

```sql
-- First, register a user normally through the API
-- Then update the user to admin:
UPDATE users SET is_admin = 1 WHERE username = 'adminuser';
```

Or use SQLite CLI:

```bash
sqlite3 tunetudo.db "UPDATE users SET is_admin = 1 WHERE username = 'adminuser';"
```

## Error Handling

The API returns consistent error responses:

```json
{
  "error": true,
  "message": "Error description here"
}
```

Success responses:

```json
{
  "error": false,
  "message": "Success message",
  "data": { /* response data */ }
}
```

## Development

### Running in Development Mode

```bash
# Install air for hot reload (optional)
go install github.com/cosmtrek/air@latest

# Run with air
air
```

### Running Tests

```bash
go test ./...
```

## Production Deployment

1. **Update environment variables** in `.env`
   - Change `JWT_SECRET` to a strong random string
   - Set `ALLOWED_ORIGINS` to your frontend domain
   - Update `DATABASE_PATH` if needed

2. **Build the application**
```bash
go build -o tunetudo main.go
```

3. **Run the binary**
```bash
./tunetudo
```

4. **Use a reverse proxy** (Nginx recommended)
   - Configure SSL/TLS certificates
   - Set up proper caching headers
   - Configure rate limiting

### Nginx Configuration Example

```nginx
server {
    listen 80;
    server_name yourdomain.com;

    location / {
        proxy_pass http://localhost:2701;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    location /storage/ {
        alias /path/to/tunetudo/storage/;
        expires 30d;
        add_header Cache-Control "public, immutable";
    }
}
```

## Troubleshooting

### Database locked error
- Ensure only one instance is running
- Check file permissions on the database file

### File upload fails
- Verify storage directory permissions
- Check file size limits
- Confirm file type is supported

### Authentication errors
- Verify JWT_SECRET is set correctly
- Check token expiration (7 days by default)
- Ensure Authorization header format: `Bearer TOKEN`

## License

This project is part of an academic assignment for UMD.

## Author

**Neel Patel**  
Email: neelp27@umd.edu  
UMD Directory ID: neelp27

## References

- Go Fiber Documentation: https://docs.gofiber.io
- SQLite FTS5: https://www.sqlite.org/fts5.html
- JWT Best Practices: https://jwt.io/introduction