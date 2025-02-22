# TODO App - (Golang + GIN)

## Overview
This is a TODO application built using Golang and the GIN framework. The project follows best practices for clean architecture, API design, and automated testing. The application is containerized and supports deployment via Docker and Kubernetes.

## 📌 Features
- CRUD operations for todo items
- S3-based attachment handling
- Kubernetes deployment with scalability
- Local development and debugging support

## 🚀 How to Run?

### Local Setup
1. Clone the repository:
   ```sh
   git clone https://github.com/deepjyotk/todo-with-golang.git
   ```
2. Run the application:
   ```sh
   go run cmd/server/main.go
   ```

### Local Debugging
1. Open VS Code.
2. Create a `.vscode` folder and add a `launch.json` file with the following content:
   ```json
   {
     "version": "0.2.0",
     "configurations": [
       {
         "name": "Debug Gin Server",
         "type": "go",
         "request": "launch",
         "mode": "auto",
         "program": "${workspaceFolder}/cmd/server/main.go",
         "cwd": "${workspaceFolder}",
         "env": {
           "GIN_MODE": "debug"
         },
         "dlvFlags": ["--check-go-version=false"]
       }
     ]
   }
   ```
3. Click **Run > Start Debugging** in VS Code.

### Running with Docker 🐳
1. Ensure you have the necessary environment variables configured.
2. Run:
   ```sh
   docker-compose up
   ```

### Running with Kubernetes (Local Cluster)
1. Install `kubectl`.
2. Navigate to the Kubernetes folder:
   ```sh
   cd ./local/
   ```
3. Create the cluster:
   ```sh
   kind create cluster --config kind-ingress-cluster.yaml
   kind create cluster --config metrics-server.yaml
   ```
4. Apply Ingress:
   ```sh
   kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
   ```
5. Deploy the application:
   ```sh
   cd ./deployments/kubernetes
   kubectl apply -f .
   ```

### Horizontal Pod Autoscaler (HPA) Configuration
- **Pod Range**: Maintains between 1 and 2 pods for the `todo-app` deployment.
- **Scaling Triggers**:
  - Scales up when average CPU utilization exceeds 30%.
  - Scales up when memory usage exceeds 256Mi.
- **Stabilization Window**: A 300-second delay prevents rapid fluctuations.


## 🌐 API Endpoints

### Health Check
- **GET** `/health`
  - **Description**: Check if the service is running.
  - **Response**: `200 OK`

### Authentication API (Public Endpoints)
- **POST** `/api/v1/auth/register`
  - **Description**: Register a new user.
  - **Request Body**:
    ```json
    {
      "username": "testuser",
      "email": "test@example.com",
      "password": "securepassword"
    }
    ```
  - **Response**:
    ```json
    {
      "message": "User registered successfully"
    }
    ```

- **POST** `/api/v1/auth/login`
  - **Description**: Authenticate user and return a JWT token.
  - **Request Body**:
    ```json
    {
      "email": "test@example.com",
      "password": "securepassword"
    }
    ```
  - **Response**:
    ```json
    {
      "token": "your-jwt-token"
    }
    ```

### Todo API (Protected Routes - Requires JWT Authentication)
- **POST** `/api/v1/todos`
  - **Description**: Create a new todo item.
  - **Headers**: `Authorization: Bearer <JWT_TOKEN>`
  - **Request Body**:
    ```json
    {
      "title": "Complete project",
      "description": "Finish the backend integration"
    }
    ```
  - **Response**:
    ```json
    {
      "id": "1234",
      "title": "Complete project",
      "description": "Finish the backend integration",
      "status": "pending"
    }
    ```

- **GET** `/api/v1/todos/:id`
  - **Description**: Retrieve a specific todo item by ID.
  - **Headers**: `Authorization: Bearer <JWT_TOKEN>`
  - **Response**:
    ```json
    {
      "id": "1234",
      "title": "Complete project",
      "description": "Finish the backend integration",
      "status": "pending"
    }
    ```

- **PUT** `/api/v1/todos/:id`
  - **Description**: Update an existing todo item.
  - **Headers**: `Authorization: Bearer <JWT_TOKEN>`
  - **Request Body**:
    ```json
    {
      "title": "Updated Project",
      "description": "New details",
      "status": "completed"
    }
    ```
  - **Response**:
    ```json
    {
      "message": "Todo updated successfully"
    }
    ```

- **DELETE** `/api/v1/todos/:id`
  - **Description**: Delete a todo item.
  - **Headers**: `Authorization: Bearer <JWT_TOKEN>`
  - **Response**:
    ```json
    {
      "message": "Todo deleted successfully"
    }
    ```

- **GET** `/api/v1/todos/get-all`
  - **Description**: Retrieve all todo items for the authenticated user.
  - **Headers**: `Authorization: Bearer <JWT_TOKEN>`
  - **Response**:
    ```json
    [
      {
        "id": "1234",
        "title": "Complete project",
        "description": "Finish the backend integration",
        "status": "pending"
      },
      {
        "id": "5678",
        "title": "Review PR",
        "description": "Check code quality",
        "status": "in-progress"
      }
    ]
    ```

### S3 Presigned URL API (For File Uploads)
- **GET** `/api/v1/todos/presigned-url`
  - **Description**: Generate a presigned URL for uploading attachments.
  - **Headers**: `Authorization: Bearer <JWT_TOKEN>`
  - **Query Parameters**:
    - `filename` (string) - The name of the file to be uploaded.
  - **Response**:
    ```json
    {
      "url": "https://todo-go-app-s3-bucket.s3.us-east-1.amazonaws.com/1/example.jpeg",
      "method": "PUT"
    }
    ```
  - **Usage**:
    1. Call this endpoint to get a **presigned URL**.
    2. Use the provided **PUT** URL to upload the file.

## 📖 Notes
- Ensure the frontend uploads attachments **before** calling `updateTodo`.
- Swagger documentation is available at `localhost/swagger/index.html`.

## 📜 License
This project is licensed under the MIT License.