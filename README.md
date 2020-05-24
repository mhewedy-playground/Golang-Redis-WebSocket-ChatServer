# Golang-Redis-WebSocket-ChatServer
Chat service using Redis, Golang and websockets


## Design Decisions
1. The chat service allows users to subscribe to multiple channels.
1. Redis client opens a TCP connection per subscription which requires opening too many connections from the chat service to Redis if we open a connection for each channel as user subscribers to.
To avoid this, whenever the user subscribes to a new channel, I have to cancel the old subscription and subscribe to all user channels at once. (Redis allow subscribe to multiple channels at once)
1. The chat service registers all connected users in two channels by default “general” and “random” (following slack conventions). The user can subscribe to any arbitrary channel at any time, and only users subscribed to such channels can see the messages.
1. I used WebSockets to handle the communication between the JavaScript client and the Golang Service.
1. The javascript client sends messages in JSON to the Websocket API which uses the following structure `{"command": <0, 1 or 2>, "channel": "channel name", "content": "content text"}` (will know more later)
1. Users and channels are saved on Redis, so we can horizontally scale the Golang Service without warring about using WebSocket the stateful protocol.
1. The Service provides a couple of REST APIs to get the active users and the channels a particular user subscribed to. (would help if we will build a GUI on top of the service)
