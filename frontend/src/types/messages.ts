import type { UnitType } from './constants';
import type { GameState, Resources } from './game';

// Client -> Server
export type ClientMessage =
  | { type: 'DEPLOY_UNIT'; payload: { unitType: UnitType; col: number; row: number; commandId?: string } }
  | { type: 'MOVE_TO'; payload: { unitId: string; col: number; row: number; commandId?: string } }
  | { type: 'SET_TARGET'; payload: { unitId: string; targetId: string; commandId?: string } }
  | { type: 'UPGRADE'; payload: { buildingId: string; commandId?: string } }
  | { type: 'SURRENDER'; payload: Record<string, never> }
  | { type: 'PING'; payload: { timestamp: number } };

// Server -> Client
export type ServerMessage =
  | { type: 'GAME_START'; payload: { gameState: GameState; yourSide: string } }
  | { type: 'STATE_UPDATE'; payload: { gameState: GameState } }
  | { type: 'COMMAND_REJECTED'; payload: { message: string } }
  | { type: 'PONG'; payload: { timestamp: number; serverTime: number } };

export interface TickDelta {
  tick: number;
  timeRemaining: number;
  resources?: { [playerId: string]: Resources };
  unitUpdates?: Array<{ id: string; col: number; row: number; hp: number; state: string }>;
  buildingUpdates?: Array<{ id: string; hp: number; level: number }>;
}
