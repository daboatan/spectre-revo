# Spectre (Revolution)

<img width="1536" height="395" alt="spectre-revo-banner-copilot" src="https://github.com/user-attachments/assets/b2787233-cddf-4493-a059-dc49987f8e38" />
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
   By default app will use port (8080), and be available at `http://localhost:8080`.

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

Environment variables can be used to customize the instance. You can set them in your environment or place them in a `.env` file in the root directory.

| Variable | Description | Default |
| :--- | :--- | :--- |
| `PORT` or `ADDR` | Binds the server to the specified port (e.g., `3000`) or address. | `0.0.0.0:8080` |
| `SPECTRE_ENV` | Set to `production` to enable Secure cookies and minification. | `dev` |
| `SPECTRE_BRAND` | Changes the site name displayed in the UI. | `Spectre` |

## Deployment

### Docker Compose
It is recommended to use Docker to run Spectre in production. The repository includes an `.env.example` file to help you get started.

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```
2. Start the container in detached mode:
   ```bash
   docker compose up -d
   ```
This setup automatically mounts the `data` directory for persistence and reads the environment variables defined in your `.env` file.

### Manual Deployment (VPS/Bare Metal)
If you prefer not to use Docker, you can run Spectre directly on your server.

1. **Clone the repository:**
   ```bash
   git clone https://github.com/borrougagnou/spectre-updated.git
   cd spectre-updated
   ```
2. **Install requirements:**
   You will need Go (>= 1.21) and Node.js installed.
3. **Configure Environment:**
   Copy the example environment file and edit it to suit your needs:
   ```bash
   cp .env.example .env
   ```
4. **Use the Install Script:**
   The repository includes an `install.sh` script to help install dependencies and build the binary:
   ```bash
   chmod +x install.sh
   ./install.sh
   ```
   *Note*: The application by default runs on port `8619` via the install script or standard Go port `8080` if run directly, but this can be changed in your `.env` file.

## Storage Architecture

Spectre avoids complex database setups by using the filesystem efficiently:
- `accounts/`: User profiles and permissions (mangled for privacy).
- `pastes/`: The raw content of pastes.
- `sessions/`: Server-side session storage.

## License

This project is licensed under the GPL License - see the [LICENSE](LICENSE) file for details.
