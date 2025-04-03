# Authentication System

A simple authentication and user management system built with Go and Gin framework.

## Features

- User authentication with login page
- Dashboard with navigation options
- User management (Add, Edit, Delete users)
- Permission levels
- Secure password storage (SHA-256 hashing)
- Integration with third-party platforms
- Persistent user data storage in JSON file

## Prerequisites

- Go 1.20 or higher

## Installation

1. Clone the repository:
```
git clone https://github.com/yourusername/auth-system.git
cd auth-system
```

2. Install dependencies:
```
go mod tidy
```

## Usage

1. Run the application:
```
go run main.go
```

Or with environment variables:
```
THIRD_PARTY_URL=https://your-platform-url.com go run main.go
```

2. Open your browser and navigate to:
```
http://localhost:8080
```

3. Login with the default admin credentials:
   - Username: `asguarder`
   - Password: `Sw@123456`

## Data Storage

User data is stored persistently in `data/users.json`. This file is automatically created when the application first runs. The file contains all user information in JSON format, including:
- User IDs
- Usernames
- Hashed passwords
- Permission levels

## Environment Variables

The application supports the following environment variables:

- `THIRD_PARTY_URL`: URL for the third-party platform integration (default: "https://example.com/third-party")

## Default Users

The system comes with one pre-configured administrator account:

- **Username**: asguarder
- **Password**: Sw@123456
- **Permission Level**: 2 (Administrator)

## Permission Levels

- **Level 1**: Regular user (can access dashboard and third-party platform)
- **Level 2**: Administrator (can access all features including user management)

## Security Notes

- Passwords are hashed using SHA-256 before storage
- This application does not use HTTPS by design (as per requirements)
- In a production environment, it is recommended to use HTTPS for secure communication
- User data is stored in a JSON file with appropriate file permissions (0644) 