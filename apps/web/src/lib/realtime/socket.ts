import { createRequestId } from '$lib/api/client';
import type { ClientMessage } from '../../../../../contracts/generated/typescript/realtime/client';
import type { ServerMessage } from '../../../../../contracts/generated/typescript/realtime/server';

export interface Checkpoint {
  roomCode: string;
  roomVersion?: number;
  matchId?: string;
  matchEventNumber?: number;
}

export class RoomSocket {
  private socket: WebSocket | null = null;
  private clientSequence = 0;
  private reconnectAttempt = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private statusTimers = new Map<string, number>();
  private stopped = false;

  constructor(
    private readonly checkpoint: () => Checkpoint,
    private readonly onMessage: (message: ServerMessage) => void,
    private readonly onConnection: (state: 'connecting' | 'reconnecting' | 'disconnected') => void,
    private readonly onCommandUncertain: (requestId: string) => void = () => {},
  ) {}

  connect(): void {
    this.stopped = false;
    this.onConnection(this.reconnectAttempt === 0 ? 'connecting' : 'reconnecting');
    const scheme = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    this.socket = new WebSocket(`${scheme}//${window.location.host}/ws`);
    this.socket.addEventListener('open', () => {
      this.reconnectAttempt = 0;
      const checkpoint = this.checkpoint();
      this.send({
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
    });
    this.socket.addEventListener('message', (event) => {
      try {
        const message = JSON.parse(String(event.data)) as ServerMessage;
        if (
          (message.type === 'command.acknowledged' ||
            message.type === 'command.rejected' ||
            message.type === 'command.status') &&
          message.payload.requestId
        ) {
          this.clearStatusTimer(message.payload.requestId);
        }
        this.onMessage(message);
      } catch {
        this.onConnection('disconnected');
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
    this.socket?.close(1000, 'page closed');
  }

  roomCommand(
    type:
      | 'room.set_ready'
      | 'room.change_settings'
      | 'room.start_countdown'
      | 'room.cancel_countdown'
      | 'room.leave'
      | 'room.transfer_host',
    roomId: string,
    expectedVersion: number,
    payload: ClientMessage['payload'],
  ): string {
    const requestId = createRequestId();
    this.send({
      schemaVersion: 1,
      requestId,
      clientSequence: this.nextSequence(),
      target: { kind: 'Room', id: roomId, expectedVersion },
      type,
      payload,
    });
    this.scheduleStatusQuery(requestId);
    return requestId;
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
  ): string {
    const requestId = createRequestId();
    this.send({
      schemaVersion: 1,
      requestId,
      clientSequence: this.nextSequence(),
      target: { kind: 'Match', id: matchId, expectedVersion },
      type,
      payload,
    });
    this.scheduleStatusQuery(requestId);
    return requestId;
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
    const checkpoint = this.checkpoint();
    this.send({
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

  private queryStatus(requestId: string): void {
    if (this.socket?.readyState !== WebSocket.OPEN) return;
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

  private send(message: ClientMessage): void {
    if (this.socket?.readyState !== WebSocket.OPEN) return;
    this.socket.send(JSON.stringify(message));
  }

  private nextSequence(): number {
    this.clientSequence += 1;
    return this.clientSequence;
  }

  private scheduleReconnect(): void {
    if (this.stopped) return;
    this.onConnection('reconnecting');
    const delays = [500, 1_000, 2_000, 4_000, 8_000, 10_000];
    const base = delays[Math.min(this.reconnectAttempt, delays.length - 1)] ?? 10_000;
    this.reconnectAttempt += 1;
    const jitter = Math.floor(base * (0.85 + Math.random() * 0.3));
    this.reconnectTimer = setTimeout(() => this.connect(), jitter);
  }
}
