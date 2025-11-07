# Whispr

**Whispr** is a real-time anonymous confession wall built with **Go** on the backend and a lightweight **TailwindCSS + Alpine.js** frontend.
It demonstrates how Go’s concurrency model and WebSockets can be used to build scalable, interactive web applications with minimal overhead.

---

## Overview

Users can post confessions anonymously, vote on others’ posts, and see real-time updates without reloading the page.
The platform is lightweight, responsive, and designed to highlight practical use of Go routines, channels, and WebSockets for concurrent systems.

---

## Key Features

* **Anonymous Posting** – Users can share thoughts without authentication.
* **Real-time Updates** – Posts appear instantly using Go WebSockets.
* **Voting System** – Upvotes and downvotes affect post ranking.
* **Trending Section** – Displays posts with the highest engagement.
* **Responsive UI** – Built with TailwindCSS for modern, mobile-friendly design.
* **Content Moderation** – Simple admin tools to hide or delete posts.
* **Scalability** – Efficient concurrency using goroutines and channels.

---

## Architecture

| Layer            | Technology                                              |
| ---------------- | ------------------------------------------------------- |
| **Backend**      | Go (Gin / Fiber), Gorilla WebSocket                     |
| **Frontend**     | HTML5, TailwindCSS, Alpine.js                           |
| **Database**     | SQLite (development), PostgreSQL (production)           |
| **Auth / Admin** | Simple token authentication via headers                 |
| **Deployment**   | Docker and Docker Compose (Railway / Render compatible) |

---

## Getting Started

### Prerequisites

* Go 1.22 or later
* Docker (optional, for containerized setup)

### Local Development

```bash
# Clone the repository
git clone https://github.com/<your-username>/whispr.git
cd whispr

# Create a local environment file
cp .env.example .env

# Run the server
go run ./cmd/server

# Open in your browser
http://localhost:8080
```

### Docker Setup

```bash
docker-compose up --build
```

---

## Configuration

All configuration values are managed through environment variables:

| Variable       | Description                          | Default                 |
| -------------- | ------------------------------------ | ----------------------- |
| `PORT`         | Port for HTTP server                 | `8080`                  |
| `DATABASE_URL` | Database connection string           | `sqlite://whispr.db`    |
| `ADMIN_TOKEN`  | Token for admin moderation endpoints | `change-me`             |
| `CORS_ORIGIN`  | Allowed origins for API access       | `http://localhost:8080` |

---

## API Reference

| Method   | Endpoint              | Description                            |
| -------- | --------------------- | -------------------------------------- |
| `GET`    | `/api/posts`          | Fetch latest posts                     |
| `GET`    | `/api/trending`       | Fetch trending posts                   |
| `POST`   | `/api/posts`          | Create a new post                      |
| `POST`   | `/api/posts/:id/vote` | Vote on a post (+1 / -1)               |
| `DELETE` | `/api/posts/:id`      | Delete post (requires `X-Admin-Token`) |
| `GET`    | `/ws`                 | WebSocket endpoint for live updates    |

---

## Frontend

The frontend is implemented using static HTML with TailwindCSS for styling and Alpine.js for interactivity.
It connects to the backend API and WebSocket endpoint for posting and real-time updates.

Key files:

* `public/index.html` – Main UI
* `public/favicon.svg` – Minimal branding icon

---

## Deployment

1. **Docker Build**

   ```bash
   docker build -t whispr .
   docker run -p 8080:8080 whispr
   ```

2. **Docker Compose (Go + Postgres)**

   ```bash
   docker-compose up
   ```

3. **Recommended Hosting**

   * **Backend**: Render, Railway, Fly.io
   * **Frontend**: Netlify, Vercel (served from `/public`)

---

## Development Notes

* SQLite is used by default for simplicity; switch to PostgreSQL via `DATABASE_URL` for production.
* WebSocket hub leverages Go’s concurrency primitives for fan-out broadcasting.
* Static admin moderation using header-based token (`X-Admin-Token`).

---

## Future Enhancements

* Rate limiting and spam control
* Profanity or sentiment analysis filters
* JWT-based admin authentication
* Hot ranking algorithm similar to Reddit or Hacker News
* Unit and integration tests using Go’s `testing` and `httptest` packages

---

## License

This project is licensed under the MIT License.
Copyright © 2025

---

## Summary

Whispr serves as a compact demonstration of:

* Go’s concurrency (goroutines and channels)
* Real-time WebSocket communication
* Lightweight, responsive frontend design

It’s a practical reference for building scalable social or chat-style applications with a minimal, modern tech stack.
