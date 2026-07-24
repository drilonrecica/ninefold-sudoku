import type { Room } from '$lib/api/client';
import type { ServerMessage } from '../../../../../contracts/generated/typescript/realtime/server';

export type { ServerMessage };

export type ConnectionState =
  'connecting' | 'connected' | 'reconnecting' | 'read_only' | 'disconnected';

export interface RoomClientState {
  room: Room | null;
  selfParticipantId: string | null;
  isController: boolean;
  connection: ConnectionState;
  recoveryRequired: boolean;
}

export const initialRoomState: RoomClientState = {
  room: null,
  selfParticipantId: null,
  isController: false,
  connection: 'connecting',
  recoveryRequired: false,
};

export function applyRoomMessage(state: RoomClientState, message: ServerMessage): RoomClientState {
  switch (message.type) {
    case 'connection.accepted':
    case 'connection.status':
      return {
        ...state,
        selfParticipantId: message.payload.identity?.participantId ?? state.selfParticipantId,
        isController: message.payload.isController ?? state.isController,
        connection: message.payload.isController === false ? 'read_only' : 'connected',
      };
    case 'connection.read_only':
    case 'connection.controller_revoked':
      return { ...state, isController: false, connection: 'read_only' };
    case 'room.snapshot': {
      const nextRoom = message.payload.room as Room | undefined;
      if (!nextRoom) return { ...state, recoveryRequired: true };
      if (state.room && nextRoom.version < state.room.version) return state;
      return { ...state, room: nextRoom, recoveryRequired: false };
    }
    case 'connection.rejected':
      return { ...state, connection: 'disconnected' };
    default:
      return state;
  }
}
