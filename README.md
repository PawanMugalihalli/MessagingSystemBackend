# Messaging System Backend

A secure, real-time messaging backend built with Go (Gin). Supports user auth, DMs, group chats, message editing with handling of concurrency issues, and chat summarization using LLMs.

---

## Features

- JWT-based authentication
- Direct messages (DMs)
- Group chat with admin/member roles
- Chat summarization using LLMs
- Edit messages (DM and group)
- View chat previews and history
- Dockerized for easy setup

---

## Getting Started

### 1. Clone the repository

git clone https://github.com/PawanMugalihalli/MessagingSystemBackend.git  
cd MessagingSystemBackend

### 2. Set up .env file

Create a `.env` file:

PORT=3000  
DB_HOST=localhost  
DB_USER=your_db_user  
DB_PASSWORD=your_db_password  
DB_NAME=messaging_db  
JWT_SECRET=your_jwt_secret  

### 3. Run with Docker

docker-compose up --build

App runs at: http://localhost:3000

---

## API Endpoints

### Authentication

- POST /signup - Register a new user  
- POST /login - Log in (JWT is set in cookie)  
- GET /validate - Check current user session  
- GET /logout - Log out user  

### Direct Messages

- POST /dm/:id - Send a direct message to a user  
- GET /dm/:id - Get messages with a user  
- PUT /dm/message/:id - Edit a direct message  

### Group Messaging

- POST /groups/create - Create a new group  
- POST /groups/:id/message - Send message to group  
- POST /groups/:id/add-member - Add member to group  
- POST /groups/:id/add-admin - Promote member to admin  
- GET /groups/:id - Get messages from group  
- GET /groups/:id/summary - Summarize group messages  
- PUT /groups/message/:id - Edit a group message  

### Chat Views

- GET /view/dms - Preview DM conversations  
- GET /view/groups - Preview group chats  
- GET /view/chat/dm/:id - View DM history  
- GET /view/chat/group/:id - View group history  

---

## Group Summarization

- Requires LLM API integration (e.g., OpenAI)
- Endpoint: GET /groups/:id/summary

---

## Assumptions

- JWT is stored in cookie named 'Authorization'
- All /dm, /groups, and /view routes require auth
- Only message authors can edit their messages
- Groups are private to members

---

## Project Structure

MessagingSystemBackend/  
├── internal/  
│   ├── controllers/  
│   ├── middleware/  
│   ├── models/  
│   └── initializers/  
├── main.go  
├── Dockerfile  
├── docker-compose.yml  
├── .env  

---

## Author

Pawan Mugalihalli  
