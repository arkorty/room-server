# room-server

This project implements the backend server for a collaborative text editor using WebSockets, allowing multiple users to edit text in real-time within designated rooms.

## Features

- Real-time text updates among clients.
- Join and leave functionality for collaborative editing sessions.
- Automatic cleanup of inactive rooms.
- Persistent storage of room content.

## Technologies Used

- WebSockets
- SQLite

## Getting Started

### Prerequisites

- Go
- SQLite

### Usage

1. Clone the repository.
2. Deploy using Docker Compose.

### Using the Application

- Connect to the WebSocket endpoint.
- Send a JSON message to join a room.
- Send text updates in JSON format.

### Room Cleanup

- Inactive rooms are automatically deleted after a specified duration.

## Contributing

Contributions are welcome! Please fork the repository and create a pull request.

## License

This project is open-source and available under the [MIT License](LICENSE).

## Acknowledgments

- Thanks to Me!
