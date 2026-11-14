import type { UnitType } from './constants';
import type { GameState, Resources } from './game';

// Client -> Server
export type ClientMessage =
  | { type: 'DEPLOY_UNIT'; payload: { unitType: UnitType; col: number; row: number } }
  | { type: 'SET_TARGET'; payload: { unitId: string; targetId: string } }
  | { type: 'UPGRADE'; payload: { buildingId: string } }
  | { type: 'SURRENDER'; payload: Record<string, never> }
  | { type: 'PING'; payload: { timestamp: number } };

// Server -> Client
export type ServerMessage =
  | { type: 'GAME_START'; payload: { gameState: GameState; yourSide: string } }
  | { type: 'TICK_UPDATE'; payload: TickDelta }
  | { type: 'UNIT_SPAWNED'; payload: { unit: { id: string; type: UnitType; owner: string; col: number; row: number; hp: number } } }
  | { type: 'UNIT_DIED'; payload: { unitId: string; killedBy: string } }
  | { type: 'BUILDING_DMG'; payload: { buildingId: string; hp: number } }
  | { type: 'BUILDING_UP'; payload: { buildingId: string; level: number } }
  | { type: 'GAME_OVER'; payload: { winner: string; reason: string } }
  | { type: 'PONG'; payload: { timestamp: number; serverTime: number } };

export interface TickDelta {
  tick: number;
  timeRemaining: number;
  resources?: { [playerId: string]: Resources };
  unitUpdates?: Array<{ id: string; col: number; row: number; hp: number; state: string }>;
  buildingUpdates?: Array<{ id: string; hp: number; level: number }>;
}
