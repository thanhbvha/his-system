class WSClient {
  private ws: WebSocket | null = null;
  connect(url: string): void {
    console.log("Connecting to WS:", url);
    // TODO: Implement actual WS connection logic
  }
  disconnect(): void {
    console.log("Disconnecting WS");
    if (this.ws) this.ws.close();
  }
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  on(event: string, handler: (data: any) => void): void {
    console.log("Registering handler for:", event);
    // TODO: Implement listener registry
  }
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  send(event: string, data: any): void {
    console.log("Sending WS message:", event, data);
  }
}

export const wsClient = new WSClient();
