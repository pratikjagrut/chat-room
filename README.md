# Real-Time Chat Room

This is a simple, real-time chat application built with Go and WebSockets. 

## How to Run

### 1. Build the Docker Image

Open your terminal, navigate to the project's root directory (the one containing the `Dockerfile`), and run the following command to build the image:

```bash
docker build -t chat-room .
```

This command creates a new Docker image named `chat-room` based on the instructions in the `Dockerfile`.

### 2. Run the Docker Container

Once the image is built, run the following command to start a container from it:

```bash
docker run -p 8080:8080 chat-room
```

This command starts the container and maps port `8080` from your local machine to port `8080` inside the container.

### 3. Access the Application

Open your web browser and navigate to:

**[http://localhost:8080](http://localhost:8080)**

You can now enter a username to join the chat. To simulate a multi-user conversation, open incognito browser tab or join with different browser.
