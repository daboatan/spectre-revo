# Spectre (Revolution)

Spectre is a modern, privacy-focused pastebin application written in Go. It is a feature-rich evolution of the Ghostbin platform, designed for performance, security, and ease of deployment.

## Features

- **Encrypted Pastes:** Secure client-side and server-side encryption options.
- **Syntax Highlighting:** Automatic detection and support for hundreds of languages.
- **Paste Expiration:** Set pastes to expire after a specific duration (minutes to days).
- **User Accounts:** Optional accounts to track and manage your "My Pastes" list across devices.
- **Admin Dashboard:** Tools for managing reports, promoting users, and moderating content.
- **Low Footprint:** No heavy database required; uses a high-performance filesystem-based storage system.
- **Responsive UI:** Built with LESS and optimized for both desktop and mobile.

## Getting Started

### Prerequisites

- [Go](https://golang.org/dl/) 1.21 or higher
- [Node.js & NPM](https://nodejs.org/) (for asset compilation)
- [GCC](https://gcc.gnu.org/) (required for some security dependencies like `scrypt`)

### Quick Start

1. **Install Dependencies:**
   ```bash
   go mod download
   npm install
   ```

2. **Initialize Keys:**
   The server will automatically generate `session.key` and `client_session_enc.key` on its first run if they don't exist.

3. **Run the Server:**
   ```bash
   go run .
   ```
   The app will be available at `http://localhost:8080`.

## Development Workflow

To develop effectively without having to restart the server for every HTML change:

### Template Hot-Reloading
Run the server with the `-rebuild` flag. This forces the Go engine to re-parse `.tmpl` files on every request.
```bash
go run . -rebuild
```

### Debugging Sessions
To see detailed identity tracing and session logs (useful for debugging login issues):
```bash
go run . -v=2 -logtostderr
```

### Asset Management
The project uses **Grunt** to manage frontend assets.
- `grunt`: Builds and minifies CSS/JS for production.
- `less.js`: Is used in development mode to compile styles in-browser.

## Configuration

Environment variables can be used to customize the instance:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `SPECTRE_ENV` | Set to `production` to enable Secure cookies and minification. | `dev` |
| `SPECTRE_BRAND` | Changes the site name displayed in the UI. | `Spectre` |

## Storage Architecture

Spectre avoids complex database setups by using the filesystem efficiently:
- `accounts/`: User profiles and permissions (mangled for privacy).
- `pastes/`: The raw content of pastes.
- `sessions/`: Server-side session storage.

## License

This project is licensed under the GPL License - see the [LICENSE](LICENSE) file for details.
