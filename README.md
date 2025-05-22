# K-12 Educational Platform

## Overview

This project is intended to be a comprehensive educational platform for K-12 students, teachers, and administrators. It aims to provide tools for learning, teaching, assessment, and collaboration.

This repository contains the foundational setup for the platform, including a Go backend and a React Spectrum frontend.

## Backend (Go)

The backend is developed in Go and provides the API services for the platform.

### Building and Running the Backend

1.  **Navigate to the backend directory:**
    ```bash
    cd backend
    ```

2.  **Build the application:**
    ```bash
    go build -o k12platform_backend
    ```
    (Note: The output filename `-o k12platform_backend` is specified for clarity. `go build` alone will also work, creating an executable named `backend` on Linux/macOS or `backend.exe` on Windows, based on the module name or directory.)

3.  **Run the executable:**
    On Linux or macOS:
    ```bash
    ./k12platform_backend
    ```
    On Windows:
    ```bash
    .\k12platform_backend.exe
    ```
    The server will start on `http://localhost:8080`.

### Backend ORM (entgo)

This project uses [entgo](https://entgo.io/) as the Object-Relational Mapper (ORM) for the Go backend to manage database interactions.

**Defining Schemas:**
Entity schemas (the structure of your data) are defined as Go code in the `backend/ent/schema/` directory. For example, `backend/ent/schema/user.go` defines the schema for users.

**Generating ORM Code:**
After you modify or add new schemas, you need to regenerate the ORM code. Navigate to the backend directory and run:
```bash
cd backend
go generate ./...
```
If you encounter issues with the `ent` command not being found, you might need to install the `ent` CLI tool locally:
```bash
go install entgo.io/ent/cmd/ent@latest
```
**Important Note:** During the automated setup of this project, the `go generate ./...` and `go install entgo.io/ent/cmd/ent@latest` commands have consistently timed out. This means the `ent` ORM code was likely not generated automatically. If you are setting this up manually, these commands are crucial.

**Database Migrations:**
For development purposes, database schema migrations are handled automatically on application startup by the `client.Schema.Create(...)` call in `backend/main.go`. This will create or update tables according to the defined schemas. For production environments, a more robust migration strategy would be required.

### Backend Access Control (Casbin)

This project uses [Casbin](https://casbin.org/) to manage access control and authorization for the Go backend.

**Model Configuration:**
The access control model (e.g., RBAC - Role-Based Access Control) is defined in `backend/casbin_model.conf`. This file specifies the structure of requests, policies, roles, and how they are matched.

**Policy Rules:**
Policy rules are defined in `backend/casbin_policy.csv`. You can add rules to grant or deny permissions.
- To grant a role permission to access a path with a specific HTTP method:
  `p, <role_name>, <path>, <http_method>`
  Example: `p, teacher, /courses, GET`
- To assign a user to a role (using placeholder User IDs for now):
  `g, <user_id>, <role_name>`
  Example: `g, user_teacher_001, teacher`

**Middleware Integration:**
Access control is enforced via middleware in `backend/main.go`:
1.  `BasicAuthMiddleware`: This is a **placeholder authentication middleware**. For testing, it reads `X-User-ID` and `X-User-Role` headers to determine the user's identity and role. If specific test user IDs (e.g., `user_admin_001`, `user_teacher_001`) are provided in `X-User-ID`, corresponding roles ("admin", "teacher") are assigned. Otherwise, it defaults to an "anonymous" user and role.
2.  `Authorizer` Middleware: This middleware uses the Casbin enforcer to check if the authenticated user's role has permission to access the requested resource (path) and action (HTTP method) based on the policies defined in `casbin_policy.csv`.

**Dependencies:**
If you haven't already, you'll need to ensure Casbin dependencies are downloaded locally by running `go mod tidy` in the `backend` directory. **Important Note:** During the automated setup of this project, `go mod tidy` commands have consistently timed out. If you are setting this up manually, this command is crucial for fetching Casbin and other backend dependencies.

## Frontend (React Spectrum)

The frontend is developed using React and Adobe's React Spectrum component library for a rich user interface.

### Building and Running the Frontend

1.  **Navigate to the frontend directory:**
    ```bash
    cd frontend
    ```

2.  **Install dependencies:**
    ```bash
    npm install
    ```
    **Important Note:** During the automated setup of this project, this step might have timed out. If you are setting this up manually, running `npm install` is crucial for the frontend to work. It will download all necessary React and React Spectrum packages defined in `frontend/package.json`.

3.  **Start the development server:**
    ```bash
    npm start
    ```
    This will typically open the application in your default web browser at `http://localhost:3000`.

## Future Development

This is a foundational setup. Significant development is required to implement the full feature set of the K-12 Educational Platform. This includes:
-   Defining and implementing robust APIs in the backend.
-   Developing various UI components and views in the frontend.
-   Setting up databases.
-   Implementing user authentication and authorization.
-   Adding features for course management, content delivery, assessments, etc.
-   Writing comprehensive tests.
-   Setting up CI/CD pipelines.

---

*This README was generated as part of an automated setup process.*
