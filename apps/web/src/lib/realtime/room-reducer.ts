import type { Room } from '$lib/api/client';
import type { ServerMessage } from '../../../../../contracts/generated/typescript/realtime/server';

export type { ServerMessage };

export type ConnectionState =
  | 'connecting'
  | 'offline'
  | 'connected'
  | 'reconnecting'
  | 'synchronizing'
  | 'read_only'
  | 'maintenance'
  | 'recovery_failed';

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
      return {
        ...state,
        selfParticipantId: message.payload.identity?.participantId ?? state.selfParticipantId,
        isController: message.payload.isController ?? state.isController,
        connection: message.payload.isController === false ? 'read_only' : 'synchronizing',
      };
    case 'connection.status': {
      const isController = message.payload.isController ?? state.isController;
      return {
        ...state,
        selfParticipantId: message.payload.identity?.participantId ?? state.selfParticipantId,
        isController,
        connection:
          message.payload.connectionState === 'maintenance'
            ? 'maintenance'
            : !isController
              ? 'read_only'
              : 'connected',
      };
    }
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
      return { ...state, connection: 'recovery_failed' };
    default:
      return state;
  }
}
