import { createRequestId } from '$lib/api/client';
import type { ClientMessage } from '../../../../../contracts/generated/typescript/realtime/client';
import type { ServerMessage } from '../../../../../contracts/generated/typescript/realtime/server';

export interface Checkpoint {
  roomCode: string;
  roomVersion?: number;
  matchId?: string;
  matchEventNumber?: number;
}

export type SocketConnectionState =
  | 'connecting'
  | 'offline'
  | 'reconnecting'
  | 'synchronizing'
  | 'connected'
  | 'read_only'
  | 'maintenance'
  | 'recovery_failed';

export class RoomSocket {
  private socket: WebSocket | null = null;
  private clientSequence = 0;
  private reconnectAttempt = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private statusTimers = new Map<string, number>();
  private pendingRequestIds = new Set<string>();
  private stopped = false;
  private terminal = false;
  private isController = false;
  private networkListenersAttached = false;
  private readonly handleOffline = () => {
    if (this.stopped || this.terminal) return;
    this.onConnection('offline');
    this.socket?.close();
  };

  constructor(
    private readonly checkpoint: () => Checkpoint,
    private readonly onMessage: (message: ServerMessage) => void,
    private readonly onConnection: (state: SocketConnectionState) => void,
    private readonly onCommandUncertain: (requestId: string) => void = () => {},
  ) {}

  connect(): void {
    if (this.terminal) return;
    this.stopped = false;
    if (!this.networkListenersAttached) {
      window.addEventListener('offline', this.handleOffline);
      this.networkListenersAttached = true;
    }
    if (!navigator.onLine) {
      this.onConnection('offline');
      this.waitUntilOnline();
      return;
    }
    this.onConnection(this.reconnectAttempt === 0 ? 'connecting' : 'reconnecting');
    const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    this.socket = new WebSocket(`${scheme}//${window.location.host}/ws`);
    this.socket.addEventListener('open', () => {
      this.reconnectAttempt = 0;
      this.onConnection('synchronizing');
      this.sendInitialize();
      for (const requestId of this.pendingRequestIds) this.queryStatus(requestId);
    });
    this.socket.addEventListener('message', (event) => {
      try {
        const message = JSON.parse(String(event.data)) as ServerMessage;
        this.handleProtocolState(message);
        this.onMessage(message);
      } catch {
        this.socket?.close();
      }
    });
    this.socket.addEventListener('close', () => this.scheduleReconnect());
    this.socket.addEventListener('error', () => this.socket?.close());
  }

  close(): void {
    this.stopped = true;
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    for (const timer of this.statusTimers.values()) clearTimeout(timer);
    this.statusTimers.clear();
    window.removeEventListener('offline', this.handleOffline);
    this.networkListenersAttached = false;
    this.socket?.close(1000, 'page closed');
  }

  roomCommand(
    type:
      | 'room.set_ready'
      | 'room.change_settings'
      | 'room.start_countdown'
      | 'room.cancel_countdown'
      | 'room.leave'
      | 'room.transfer_host'
      | 'room.prepare_rematch',
    roomId: string,
    expectedVersion: number,
    payload: ClientMessage['payload'],
  ): string | null {
    return this.authoritativeCommand({
      schemaVersion: 1,
      requestId: createRequestId(),
      clientSequence: this.nextSequence(),
      target: { kind: 'Room', id: roomId, expectedVersion },
      type,
      payload,
    });
  }

  matchCommand(
    type:
      | 'match.place_value'
      | 'match.erase_value'
      | 'match.add_note'
      | 'match.remove_note'
      | 'match.use_hint'
      | 'match.ping',
    matchId: string,
    expectedVersion: number,
    payload: ClientMessage['payload'],
  ): string | null {
    return this.authoritativeCommand({
      schemaVersion: 1,
      requestId: createRequestId(),
      clientSequence: this.nextSequence(),
      target: { kind: 'Match', id: matchId, expectedVersion },
      type,
      payload,
    });
  }

  publishFocus(cell: number, focused: boolean): void {
    this.send({
      schemaVersion: 1,
      requestId: createRequestId(),
      clientSequence: this.nextSequence(),
      type: focused ? 'match.focus_cell' : 'match.release_focus',
      payload: { cell },
    });
  }

  sendReaction(reaction: 'agree' | 'nice_move'): void {
    this.send({
      schemaVersion: 1,
      requestId: createRequestId(),
      clientSequence: this.nextSequence(),
      type: 'match.reaction',
      payload: { reaction },
    });
  }

  requestControl(): void {
    this.send({
      schemaVersion: 1,
      requestId: createRequestId(),
      clientSequence: this.nextSequence(),
      type: 'connection.request_control',
      payload: {},
    });
  }

  synchronize(): void {
    if (this.sendInitialize()) this.onConnection('synchronizing');
  }

  restoreUncertain(requestIds: string[]): void {
    for (const requestId of requestIds.slice(0, 32)) {
      this.pendingRequestIds.add(requestId);
      this.onCommandUncertain(requestId);
    }
  }

  private authoritativeCommand(message: ClientMessage): string | null {
    if (!this.send(message)) return null;
    const requestId = String(message.requestId);
    this.pendingRequestIds.add(requestId);
    this.scheduleStatusQuery(requestId);
    return requestId;
  }

  private sendInitialize(): boolean {
    const checkpoint = this.checkpoint();
    return this.send({
      schemaVersion: 1,
      requestId: createRequestId(),
      clientSequence: this.nextSequence(),
      type: 'connection.initialize',
      payload: {
        roomCode: checkpoint.roomCode,
        lastRoomVersion: checkpoint.roomVersion,
        lastMatchId: checkpoint.matchId,
        lastMatchEventNumber: checkpoint.matchEventNumber,
      },
    });
  }

  private handleProtocolState(message: ServerMessage): void {
    if (
      (message.type === 'command.acknowledged' ||
        message.type === 'command.rejected' ||
        message.type === 'command.status') &&
      message.payload.requestId
    ) {
      this.clearStatusTimer(message.payload.requestId);
      if (message.type !== 'command.status' || message.payload.status !== 'pending') {
        this.pendingRequestIds.delete(message.payload.requestId);
      }
    }
    if (message.type === 'connection.accepted' && message.payload.isController !== undefined) {
      this.isController = message.payload.isController;
    }
    if (message.type === 'connection.rejected') {
      const code = message.payload.code;
      if (code === 'SESSION_INVALID' || code === 'SESSION_EXPIRED' || code === 'ROOM_NOT_FOUND') {
        this.terminal = true;
        this.onConnection('recovery_failed');
      } else {
        this.onConnection('maintenance');
      }
      return;
    }
    if (
      message.type === 'connection.read_only' ||
      message.type === 'connection.controller_revoked'
    ) {
      this.isController = false;
      this.onConnection('read_only');
      return;
    }
    if (message.type === 'connection.status') {
      if (message.payload.connectionState === 'maintenance') {
        this.onConnection('maintenance');
        return;
      }
      if (message.payload.isController !== undefined) {
        this.isController = message.payload.isController;
      }
      this.onConnection(this.isController ? 'connected' : 'read_only');
    }
  }

  private queryStatus(requestId: string): void {
    this.send({
      schemaVersion: 1,
      requestId,
      clientSequence: this.nextSequence(),
      type: 'command.status',
      payload: {},
    });
  }

  private scheduleStatusQuery(requestId: string): void {
    this.clearStatusTimer(requestId);
    this.statusTimers.set(
      requestId,
      window.setTimeout(() => {
        this.statusTimers.delete(requestId);
        this.onCommandUncertain(requestId);
        this.queryStatus(requestId);
      }, 5_000),
    );
  }

  private clearStatusTimer(requestId: string): void {
    const timer = this.statusTimers.get(requestId);
    if (timer) clearTimeout(timer);
    this.statusTimers.delete(requestId);
  }

  private send(message: ClientMessage): boolean {
    if (this.socket?.readyState !== WebSocket.OPEN) return false;
    this.socket.send(JSON.stringify(message));
    return true;
  }

  private nextSequence(): number {
    this.clientSequence += 1;
    return this.clientSequence;
  }

  private scheduleReconnect(): void {
    if (this.stopped || this.terminal) return;
    if (!navigator.onLine) {
      this.onConnection('offline');
      this.waitUntilOnline();
      return;
    }
    this.onConnection('reconnecting');
    const delays = [500, 1_000, 2_000, 4_000, 8_000, 10_000];
    const base = delays[Math.min(this.reconnectAttempt, delays.length - 1)] ?? 10_000;
    this.reconnectAttempt += 1;
    const jitter = Math.floor(base * (0.85 + Math.random() * 0.3));
    this.reconnectTimer = setTimeout(() => this.connect(), jitter);
  }

  private waitUntilOnline(): void {
    const resume = () => {
      if (!this.stopped && !this.terminal) this.connect();
    };
    window.addEventListener('online', resume, { once: true });
  }
}
