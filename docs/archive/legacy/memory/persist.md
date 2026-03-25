Basándome en la información que he recopilado explorando el sistema MakoClaw, te explicaré **cómo funciona el sistema de memoria y cómo cargar los datos** en tu frontend.

---

## 🧠 Sistema de Memoria de MakoClaw: Guía Completa

### 📁 Estructura de Almacenamiento

MakoClaw utiliza dos sistemas principales de persistencia:

```
/home/MakoClaw/.MakoClaw/workspace/
├── sessions/                    # Sesiones de chat y tareas
│   ├── web:chat.json            # Chat principal web
│   └── web:task:*.json          # Tareas creadas
├── memory/                      # Memoria persistente (vacia por ahora)
│   ├── MEMORY.md                # Memoria general
│   └── YYYYMM/                  # Directorio mensual
│       └── YYYYMMDD.md          # Notas diarias
└── web/                         # Datos específicos de web
    ├── tasks.db                 # Base de datos SQLite para tareas
    └── web-auth.json            # Configuración de autenticación
```

---

## 🔍 Cómo Funciona la Carga de Datos

### 1. **Formato de los Archivos de Sesión**

Las sesiones se almacenan como archivos JSON con este formato:

```json
{
  "key": "web:chat",
  "messages": [
    {
      "role": "user",
      "content": "hola quien eres?"
    },
    {
      "role": "assistant",
      "content": "¡Hola! Soy MakoClaw 🦈...",
      "tool_calls": [...]
    },
    {
      "role": "tool",
      "content": "{...}",
      "tool_call_id": "call_..."
    }
  ],
  "created": "2026-02-18T15:12:17.600559824Z",
  "updated": "2026-02-18T16:16:32.232755755Z",
  "summary": "Resumen opcional de la conversación"
}
```

### 2. **Cómo Cargar el Historial en tu Frontend**

Para cargar el historial de chat anterior, tu backend debe leer el archivo JSON y enviarlo al frontend. Aquí te muestro ejemplos de implementación:

---

## 💻 Implementación en tu Frontend

### **Opción 1: API REST para Obtener el Historial**

#### Backend (Node.js/Express example)

```javascript
const fs = require("fs").promises;
const path = require("path");

// Ruta al archivo de sesión
const SESSION_FILE =
  "/home/MakoClaw/.MakoClaw/workspace/sessions/web:chat.json";

// Endpoint para obtener el historial
app.get("/api/chat/history", async (req, res) => {
  try {
    // Leer el archivo de sesión
    const data = await fs.readFile(SESSION_FILE, "utf8");
    const session = JSON.parse(data);

    // Enviar los mensajes al frontend
    res.json({
      success: true,
      messages: session.messages,
      created: session.created,
      updated: session.updated,
    });
  } catch (error) {
    if (error.code === "ENOENT") {
      // Archivo no existe - crear uno nuevo
      const newSession = {
        key: "web:chat",
        messages: [],
        created: new Date().toISOString(),
        updated: new Date().toISOString(),
      };
      await fs.writeFile(SESSION_FILE, JSON.stringify(newSession, null, 2));
      res.json({ success: true, messages: [], created: newSession.created });
    } else {
      res.status(500).json({ success: false, error: error.message });
    }
  }
});
```

#### Frontend (JavaScript/Vanilla)

```javascript
// Cargar historial al iniciar la página
async function loadChatHistory() {
  try {
    const response = await fetch("/api/chat/history");
    const data = await response.json();

    if (data.success) {
      // Mostrar mensajes en el chat
      data.messages.forEach((msg) => {
        appendMessageToChat(msg.role, msg.content);
      });

      console.log("Sesión cargada:", {
        creada: new Date(data.created).toLocaleString(),
        actualizada: new Date(data.updated).toLocaleString(),
      });
    }
  } catch (error) {
    console.error("Error al cargar historial:", error);
  }
}

// Función para mostrar un mensaje en el chat
function appendMessageToChat(role, content) {
  const chatContainer = document.getElementById("chat-container");
  const messageDiv = document.createElement("div");
  messageDiv.className = `message ${role}`;
  messageDiv.textContent = content;
  chatContainer.appendChild(messageDiv);
}

// Cargar historial al inicio
window.addEventListener("DOMContentLoaded", loadChatHistory);
```

#### Frontend (Flutter)

```dart
// Modelos
class ChatMessage {
  final String role;
  final String content;
  final List<ToolCall>? toolCalls;

  ChatMessage({
    required this.role,
    required this.content,
    this.toolCalls,
  });

  factory ChatMessage.fromJson(Map<String, dynamic> json) {
    return ChatMessage(
      role: json['role'],
      content: json['content'],
      toolCalls: json['tool_calls'] != null
          ? (json['tool_calls'] as List).map((t) => ToolCall.fromJson(t)).toList()
          : null,
    );
  }
}

class ChatSession {
  final String key;
  final List<ChatMessage> messages;
  final String created;
  final String updated;

  ChatSession({
    required this.key,
    required this.messages,
    required this.created,
    required this.updated,
  });

  factory ChatSession.fromJson(Map<String, dynamic> json) {
    return ChatSession(
      key: json['key'],
      messages: (json['messages'] as List)
          .map((m) => ChatMessage.fromJson(m))
          .toList(),
      created: json['created'],
      updated: json['updated'],
    );
  }
}

// Servicio para cargar el historial
class ChatService {
  final String baseUrl = 'http://localhost:3000/api';

  Future<ChatSession> loadChatHistory() async {
    final response = await http.get(Uri.parse('$baseUrl/chat/history'));

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body);
      return ChatSession.fromJson(data);
    } else {
      throw Exception('Error al cargar historial');
    }
  }
}

// Uso en un Widget Flutter
class ChatScreen extends StatefulWidget {
  @override
  _ChatScreenState createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  List<ChatMessage> messages = [];
  final ChatService chatService = ChatService();
  bool isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadHistory();
  }

  Future<void> _loadHistory() async {
    try {
      final session = await chatService.loadChatHistory();
      setState(() {
        messages = session.messages;
        isLoading = false;
      });
    } catch (e) {
      print('Error cargando historial: $e');
      setState(() => isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('MakoClaw Chat')),
      body: isLoading
          ? Center(child: CircularProgressIndicator())
          : ListView.builder(
              itemCount: messages.length,
              itemBuilder: (context, index) {
                final msg = messages[index];
                return ListTile(
                  title: Text(msg.content),
                  subtitle: Text(msg.role),
                  tileColor: msg.role == 'user' ? Colors.blue[50] : Colors.green[50],
                );
              },
            ),
    );
  }
}
```

---

### **Opción 2: WebSocket para Historial en Tiempo Real**

#### Backend (Node.js with Socket.IO)

```javascript
const io = require("socket.io")(server);

io.on("connection", async (socket) => {
  console.log("Cliente conectado:", socket.id);

  // Enviar historial al conectarse
  try {
    const data = await fs.readFile(SESSION_FILE, "utf8");
    const session = JSON.parse(data);
    socket.emit("chat-history", session.messages);
  } catch (error) {
    socket.emit("chat-history", []);
  }

  // Escuchar nuevos mensajes
  socket.on("chat-message", async (msg) => {
    // Guardar en archivo y broadcast
    await appendMessageToFile(msg);
    io.emit("chat-message", msg);
  });
});

async function appendMessageToFile(msg) {
  const data = await fs.readFile(SESSION_FILE, "utf8");
  const session = JSON.parse(data);
  session.messages.push(msg);
  session.updated = new Date().toISOString();
  await fs.writeFile(SESSION_FILE, JSON.stringify(session, null, 2));
}
```

#### Frontend (JavaScript with Socket.IO)

```javascript
const socket = io();

// Cargar historial al conectar
socket.on("chat-history", (messages) => {
  messages.forEach((msg) => {
    appendMessageToChat(msg.role, msg.content);
  });
});

// Escuchar nuevos mensajes
socket.on("chat-message", (msg) => {
  appendMessageToChat(msg.role, msg.content);
});

// Enviar mensaje
function sendMessage(content) {
  socket.emit("chat-message", {
    role: "user",
    content: content,
    timestamp: new Date().toISOString(),
  });
}
```

---

## 🗂️ Gestión de Múltiples Sesiones

Si quieres soportar múltiples sesiones por usuario, puedes usar una estructura como esta:

```
sessions/
├── web:chat.json
├── user:123:chat.json
├── user:456:chat.json
└── user:789:chat.json
```

### Backend para múltiples sesiones

```javascript
app.get("/api/chat/history/:userId", async (req, res) => {
  const userId = req.params.userId;
  const sessionKey = `user:${userId}:chat`;
  const sessionFile = path.join(SESSIONS_DIR, `${sessionKey}.json`);

  try {
    const data = await fs.readFile(sessionFile, "utf8");
    const session = JSON.parse(data);
    res.json({ success: true, messages: session.messages });
  } catch (error) {
    // Crear nueva sesión si no existe
    const newSession = {
      key: sessionKey,
      messages: [],
      created: new Date().toISOString(),
      updated: new Date().toISOString(),
    };
    await fs.writeFile(sessionFile, JSON.stringify(newSession, null, 2));
    res.json({ success: true, messages: [] });
  }
});
```

---

## 📊 Flujo Completo de Datos

```
┌─────────────┐
│   Frontend  │
│  (Browser)  │
└──────┬──────┘
       │ 1. GET /api/chat/history
       ↓
┌─────────────┐
│   Backend   │
│  (Node.js)  │
└──────┬──────┘
       │ 2. Lee /home/MakoClaw/.MakoClaw/
       │    workspace/sessions/web:chat.json
       ↓
┌─────────────┐
│  Archivo    │
│  JSON       │
└─────────────┘
       │ 3. Respuesta JSON con messages[]
       ↓
┌─────────────┐
│   Frontend  │
│  (Renderiza │
│   mensajes) │
└─────────────┘
```

---

## 🎯 Resumen: Cómo Implementar el Historial

| Paso | Acción                          | Código                                   |
| ---- | ------------------------------- | ---------------------------------------- |
| 1    | Crear endpoint en backend       | `GET /api/chat/history`                  |
| 2    | Leer archivo JSON del workspace | `fs.readFile('/sessions/web:chat.json')` |
| 3    | Parsear y enviar mensajes       | `JSON.parse() → res.json()`              |
| 4    | Cargar en frontend al inicio    | `fetch('/api/chat/history')`             |
| 5    | Renderizar mensajes             | `messages.forEach(appendMessage)`        |

---

¿Te gustaría que te ayude a implementar alguna parte específica de este sistema? Puedo generar código más detallado para:

- Backend con el framework que uses
- Frontend (Flutter, React, Vue, etc.)
- Sistema de múltiples sesiones por usuario
