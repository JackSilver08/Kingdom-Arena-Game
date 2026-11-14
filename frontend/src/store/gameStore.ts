import { create } from 'zustand';
import type { GameState, Resources, Building, Unit } from '../types/game';
import {
  INITIAL_GOLD,
  INITIAL_WOOD,
  INITIAL_STONE,
  BOARD_COLS,
  BOARD_ROWS,
  BUILDING_POSITIONS,
} from '../types/constants';

interface GameStore {
  gameState: GameState | null;
  mySide: string | null;
  initGame: (side: string) => void;
  updateResources: (playerId: string, resources: Resources) => void;
  addUnit: (unit: Unit) => void;
  removeUnit: (unitId: string) => void;
  updateTick: (tick: number, timeRemaining: number) => void;
  upgradeBuilding: (buildingId: string) => void;
  setGameState: (state: GameState) => void;
  setMySide: (side: string) => void;
}

function createInitialBuildings(side: string): Building[] {
  const pos = side === 'player1' ? BUILDING_POSITIONS.player1 : BUILDING_POSITIONS.player2;
  return [
    { id: `${side}_castle`, type: 'castle', owner: side, position: pos.castle, level: 1, hp: 500, maxHp: 500 },
    { id: `${side}_fortress1`, type: 'fortress', owner: side, position: pos.fortress1, level: 1, hp: 300, maxHp: 300 },
    { id: `${side}_fortress2`, type: 'fortress', owner: side, position: pos.fortress2, level: 1, hp: 300, maxHp: 300 },
    { id: `${side}_village1`, type: 'village', owner: side, position: pos.village1, level: 1, hp: 150, maxHp: 150 },
    { id: `${side}_village2`, type: 'village', owner: side, position: pos.village2, level: 1, hp: 150, maxHp: 150 },
    { id: `${side}_village3`, type: 'village', owner: side, position: pos.village3, level: 1, hp: 150, maxHp: 150 },
  ];
}

export const useGameStore = create<GameStore>((set) => ({
  gameState: null,
  mySide: null,

  initGame: (side: string) => {
    const board: (string | null)[][] = Array.from({ length: BOARD_COLS }, () =>
      Array.from({ length: BOARD_ROWS }, () => null)
    );
    const p1Buildings = createInitialBuildings('player1');
    const p2Buildings = createInitialBuildings('player2');
    [...p1Buildings, ...p2Buildings].forEach((building) => {
      board[building.position.col][building.position.row] = building.id;
    });

    set({
      gameState: {
        tick: 0,
        timeRemaining: 600,
        phase: 'playing',
        players: {
          player1: {
            resources: { gold: INITIAL_GOLD, wood: INITIAL_WOOD, stone: INITIAL_STONE },
            buildings: p1Buildings,
            units: [],
          },
          player2: {
            resources: { gold: INITIAL_GOLD, wood: INITIAL_WOOD, stone: INITIAL_STONE },
            buildings: p2Buildings,
            units: [],
          },
        },
        board,
      },
      mySide: side,
    });
  },

  updateResources: (playerId, resources) =>
    set((state) => {
      if (!state.gameState) return state;
      return {
        gameState: {
          ...state.gameState,
          players: {
            ...state.gameState.players,
            [playerId]: { ...state.gameState.players[playerId], resources },
          },
        },
      };
    }),

  addUnit: (unit) =>
    set((state) => {
      if (!state.gameState) return state;
      const player = state.gameState.players[unit.owner];
      if (!player) return state;
      return {
        gameState: {
          ...state.gameState,
          players: {
            ...state.gameState.players,
            [unit.owner]: { ...player, units: [...player.units, unit] },
          },
        },
      };
    }),

  removeUnit: (unitId) =>
    set((state) => {
      if (!state.gameState) return state;
      const players = { ...state.gameState.players };
      for (const playerId of Object.keys(players)) {
        players[playerId] = {
          ...players[playerId],
          units: players[playerId].units.filter((unit) => unit.id !== unitId),
        };
      }
      return { gameState: { ...state.gameState, players } };
    }),

  upgradeBuilding: (buildingId) =>
    set((state) => {
      if (!state.gameState) return state;
      const players = { ...state.gameState.players };
      for (const playerId of Object.keys(players)) {
        players[playerId] = {
          ...players[playerId],
          buildings: players[playerId].buildings.map((building) =>
            building.id === buildingId ? { ...building, level: Math.min(3, building.level + 1) } : building
          ),
        };
      }
      return { gameState: { ...state.gameState, players } };
    }),

  updateTick: (tick, timeRemaining) =>
    set((state) => {
      if (!state.gameState) return state;
      return { gameState: { ...state.gameState, tick, timeRemaining } };
    }),

  setGameState: (gameState) => set({ gameState }),
  setMySide: (mySide) => set({ mySide }),
}));
