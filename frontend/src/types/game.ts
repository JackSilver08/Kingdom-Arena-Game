import type { UnitType } from './constants';

export interface Position {
  col: number;
  row: number;
}

export interface Resources {
  gold: number;
  wood: number;
  stone: number;
}

export interface Building {
  id: string;
  type: 'castle' | 'fortress' | 'village';
  owner: string;
  position: Position;
  level: number;
  hp: number;
  maxHp: number;
}

export interface Unit {
  id: string;
  type: UnitType;
  owner: string;
  position: Position;
  targetPosition?: Position;
  targetId?: string;
  targetBuildingId?: string;
  hp: number;
  maxHp: number;
  atk: number;
  state: 'idle' | 'moving' | 'attacking' | 'dead';
  lastAttackTime: number;
}

export interface GameState {
  tick: number;
  timeRemaining: number;
  phase: 'waiting' | 'playing' | 'finished';
  winner?: string;
  players: {
    [playerId: string]: {
      resources: Resources;
      buildings: Building[];
      units: Unit[];
    };
  };
  board?: (string | null)[][]; // Optional in authoritative server snapshots.
}
