# WebShark Frontend - Vue 3 + Vite + TypeScript

This is the frontend for WebShark, a web-based network packet capture and analysis tool.

## 🚀 Getting Started

### Prerequisites
- Node.js 18+ 
- npm or yarn

### Installation

```bash
# Install dependencies
npm install
```

### Development

```bash
# Start development server
npm run dev
```

The application will be available at `http://localhost:38081`

### Build

```bash
# Build for production
npm run build
```

### Preview Production Build

```bash
# Preview the production build
npm run preview
```

### Type Check

```bash
# Run TypeScript type checking
npm run type-check
```

## 📁 Project Structure

```
src/
├── components/          # Vue components
│   ├── CaptureSection.vue
│   ├── ConnectionSection.vue
│   ├── FilterSection.vue
│   ├── Header.vue
│   ├── InterfaceSection.vue
│   └── StatusMessage.vue
├── services/           # API and WebSocket services
│   ├── api.service.ts
│   └── websocket.service.ts
├── styles/             # Global styles
│   └── main.css
├── types/              # TypeScript type definitions
│   └── index.ts
├── App.vue             # Root component
└── main.ts             # Application entry point
```

## 🔧 Technology Stack

- **Vue 3** - Progressive JavaScript framework
- **TypeScript** - Type-safe JavaScript
- **Vite** - Next generation frontend tooling
- **WebSocket API** - Real-time packet streaming
- **Fetch API** - HTTP requests

## 📝 Features

- Connect to remote hosts via SSH
- Select network interfaces
- Configure BPF and Wireshark filters
- Real-time packet capture and display
- Packet detail viewer
- WebSocket-based real-time streaming

## 🎯 Component Overview

### Components
- **Header**: Application header with title
- **ConnectionSection**: Remote host connection form
- **InterfaceSection**: Network interface selection with checkboxes
- **FilterSection**: BPF and Wireshark filter configuration
- **CaptureSection**: Packet capture results display
- **StatusMessage**: Toast notifications for user feedback

### Services
- **ApiService**: HTTP API communication
- **WebSocketService**: WebSocket connection management

## 🔌 Backend Integration

This frontend connects to a Go backend running on port 38081 with the following endpoints:

- `GET /api/interfaces` - Get network interfaces
- `POST /api/capture/start` - Start packet capture
- `POST /api/capture/stop` - Stop packet capture
- `WS /ws/capture` - WebSocket for real-time packet streaming

## 🎨 Styling

The application uses custom CSS with:
- Modern gradient backgrounds
- Responsive layout
- Custom scrollbar styling
- Smooth animations
- Card-based UI design

## 📄 License

This project is part of the WebShark network packet capture tool.